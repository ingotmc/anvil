package anvil

import (
	"fmt"
	"math"
	"os"
	"path"
)

// Dimension is a collection of region files. It's so called because minecraft stores dimensions in separate folders.
type Dimension struct {
	regionDir string
}

// OpenDimension returns a new Dimension with regionDir as the base directory for region files
func OpenDimension(regionDir string) *Dimension {
	return &Dimension{
		regionDir: regionDir,
	}
}

// Block returns the id of the block at (x, y, z) in the dimension
func (d *Dimension) Block(x, y, z int) (id string, err error) {
	colX, colZ := x>>4, z>>4
	regX, regZ := colX>>5, colZ>>5
	reg, err := d.loadRegion(regX, regZ)
	if err != nil {
		return
	}
	col, err := reg.Column(colX, colZ)
	if err != nil {
		return
	}
	chunkY := byte(math.Floor(float64(y) / 16))
	chunk, err := col.Chunk(chunkY)
	if err != nil {
		return
	}
	return chunk.Block(x&15, y&15, z&15)
}

func (d *Dimension) loadRegion(x, z int) (reg Region, err error) {
	regFilePath := path.Join(d.regionDir, fmt.Sprintf("r.%d.%d.mca", x, z))
	if _, err = os.Stat(regFilePath); err != nil {
		return
	}
	regFile, err := os.Open(regFilePath)
	if err != nil {
		return
	}
	defer regFile.Close()
	return ParseRegion(regFile)
}
