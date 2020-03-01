package anvil

import "errors"

// Column represents a world section composed by 1x16x1 chunks
type Column struct {
	XPos     int32   `nbt:"xPos"`
	ZPos     int32   `nbt:"zPos"`
	Sections []Chunk `nbt:"Sections"`
}

// Chunk returns the Chunk at height y (0 to 15) in the chunk column
func (c Column) Chunk(y byte) (chunk Chunk, err error) {
	for _, ch := range c.Sections {
		if ch.Y == y {
			chunk = ch
			return
		}
	}
	err = errors.New("no chunk found")
	return
}
