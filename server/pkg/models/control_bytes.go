package models

import (
	"encoding/binary"
	"encoding/json"
	"image"
)

// Size of control bytes buffer
const (
	ControlBytesSize = 32
)

// ControlBytes ...
type ControlBytes []byte

// MarshalJSON - custom serialization
func (b ControlBytes) MarshalJSON() ([]byte, error) {
	ints := make([]int, len(b))
	for i, v := range b {
		ints[i] = int(v)
	}
	return json.Marshal(ints)
}

// UnmarshalJSON - custom deserialization
func (b *ControlBytes) UnmarshalJSON(data []byte) error {
	var ints []int
	if err := json.Unmarshal(data, &ints); err != nil {
		return err
	}

	result := make([]byte, len(ints))
	for i, v := range ints {
		result[i] = byte(v)
	}

	*b = result
	return nil
}

func (b *ControlBytes) CountOffset(
	stepZone *image.Rectangle,
) (int, int) {
	if ImageRectIsEmpty(stepZone) || b == nil || len(*b) != ControlBytesSize {
		return 0, 0
	}

	x := binary.BigEndian.Uint32((*b)[10:14])
	y := binary.BigEndian.Uint32((*b)[14:18])

	randX, randY := GetRandomXY(stepZone)
	return randX - int(x), randY - int(y)
}

// ApplyOffset ...
func (b *ControlBytes) ApplyOffset(x, y int) {
	if b == nil || len(*b) != ControlBytesSize || (x == 0 && y == 0) {
		return
	}

	var oldX = binary.BigEndian.Uint32((*b)[10:14])
	var oldY = binary.BigEndian.Uint32((*b)[14:18])

	var xWithOffset = int(oldX) + x
	if xWithOffset < 0 {
		xWithOffset = 0
	}

	var yWithOffset = int(oldY) + y
	if yWithOffset < 0 {
		yWithOffset = 0
	}

	binary.BigEndian.PutUint32((*b)[10:14], uint32(xWithOffset))
	binary.BigEndian.PutUint32((*b)[14:18], uint32(yWithOffset))
}
