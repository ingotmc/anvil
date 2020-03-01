package anvil

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"gitlab.com/guglicap/gomcsrv/pkg/nbt"
	"io"
	"io/ioutil"
)

// Region represents a section of world composed by 32x1x32 columns
type Region struct {
	X, Z       int
	Header     RegionHeader
	ColumnData []byte
}

// ParseRegion returns a Region decoded from the Reader r, eventually returning an error.
func ParseRegion(r io.Reader) (reg Region, err error) {
	if len(reg.Header) == 0 {
		reg.Header = make(RegionHeader, 8192)
	}
	_, err = io.ReadFull(r, reg.Header)
	if err != nil {
		return
	}
	reg.ColumnData, err = ioutil.ReadAll(r)
	return
}

var ErrInvalidColumnCoords = errors.New("specified column coordinates are invalid, check region file is correct")
var ErrInvalidColumnSize = errors.New("column exact size is greater than maximum column size specified in region header")

// Column returns the column at (x,z)
func (reg *Region) Column(x, z int) (c Column, err error) {
	if x > 31 || z > 31 {
		err = ErrInvalidColumnCoords
		return
	}
	cl, err := reg.Header.GetColumnLocation(x, z)
	if err != nil {
		return
	}
	// offset is from begging of file, which includes header, that's why we're subtracting 8192
	colOffset := cl.ColumnOffset() - 8192
	colMaxSize := cl.ColumnSize()
	col := reg.ColumnData[colOffset:(colOffset + colMaxSize)]
	colExactSize := binary.BigEndian.Uint32(col[:4])
	if uint(colExactSize) > colMaxSize {
		err = ErrInvalidColumnSize
		return
	}
	// we skip compression check because it's always zlib as of reference
	_ = col[4]
	compressedReader, err := zlib.NewReader(bytes.NewReader(col[5 : 5+(colExactSize-1)]))
	if err != nil {
		return
	}
	temp := struct {
		Level       Column `nbt:"Level"`
		DataVersion int32  `nbt:"DataVersion"`
	}{}
	err = nbt.NewDecoder(compressedReader).Decode(&temp)
	c = temp.Level
	return
}

type RegionHeader []byte

var ErrOffsetOutOfBounds = errors.New("column location offset is out of bounds")

// GetColumnLocation returns the location of the column at (x,z)
func (rh RegionHeader) GetColumnLocation(x, z int) (c ColumnLocation, err error) {
	offset := 4 * ((x & 31) + (z&31)*32)
	if (offset + 4) >= len(rh) {
		err = ErrOffsetOutOfBounds
		return
	}
	c = ColumnLocation(rh[offset : offset+4])
	return
}

type ColumnLocation []byte

func (cl ColumnLocation) ColumnOffset() (offset uint) {
	offset = ((uint(cl[0]) << 16) | (uint(cl[1]) << 8) | uint(cl[2])) * 4096
	return
}

func (cl ColumnLocation) ColumnSize() (size uint) {
	size = uint(cl[3]) * 4096
	return
}
