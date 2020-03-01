package anvil

import (
	"errors"
	"math"
)

// Chunk represents a world section of 16x16x16 blocks
type Chunk struct {
	Y       byte `nbt:"Y"`
	Palette []struct {
		Name string `nbt:"Name"`
	} `nbt:"Palette"`
	BlockLight  []byte  `nbt:"BlockLight"`
	BlockStates []int64 `nbt:"BlockStates"`
	SkyLight    []byte  `nbt:"SkyLight"`
}

// Block returns the id of the block at (x, y, z) **in chunk coordinates**
func (c Chunk) Block(x, y, z int) (id string, err error) {
	i := uint64(y*16*16 + z*16 + x)
	bitsPerBlock := uint64(math.Ceil(math.Log2(float64(len(c.Palette)))))
	if bitsPerBlock < 4 {
		bitsPerBlock = 4
	}
	bitIndex := i * bitsPerBlock
	longIndex := int(math.Floor(float64(bitIndex) / 64.0))
	if longIndex >= len(c.BlockStates) {
		err = errors.New("longIndex is out of bounds for the blockstates array")
		return
	}
	long := uint64(c.BlockStates[longIndex])
	longBitIndex := bitIndex % 64
	if bitsTillLongEnd := 64 - longBitIndex; bitsTillLongEnd < bitsPerBlock {
		if longIndex+1 >= len(c.BlockStates) {
			err = errors.New("parsing block required reading outside the bounds of the blockstates array")
			return
		}
		long >>= longBitIndex
		leftoverBytes := bitsPerBlock - bitsTillLongEnd
		long <<= leftoverBytes
		nextLong := uint64(c.BlockStates[longIndex+1])
		nextLong &= ^uint64(0xffffffffffffffff << leftoverBytes)
		long |= nextLong
		// our value is at the beginning of this long
		longBitIndex = 0
	}
	long >>= longBitIndex
	mask := ^uint64(0xffffffffffffffff << bitsPerBlock)
	long &= mask
	paletteIndex := uint16(long)
	if int(paletteIndex) >= len(c.Palette) {
		err = errors.New("palette out of bounds")
		return
	}
	id = c.Palette[paletteIndex].Name
	return
}
