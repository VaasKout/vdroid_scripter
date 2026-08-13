package models

import (
	"encoding/binary"
	"encoding/json"
	"image"
	"math/rand/v2"
)

// Size of control bytes buffer
const (
	ControlBytesSize = 32
)

const (
	TypeInjectTouchEvent byte = 2

	PointerIDGenericFinger uint64 = 0xFFFFFFFFFFFFFFFE
	PressureMax            uint16 = 0xFFFF

	TapDurationMinMs    = 50
	TapDurationJitterMs = 70

	LongTapDurationMinMs    = 700
	LongTapDurationJitterMs = 200
)

const (
	ActionDown byte = 0
	ActionUp   byte = 1
)

// Event ...
type Event struct {
	Time int64        `json:"time"`
	Data ControlBytes `json:"data"`
}

// ControlBytes ...
type ControlBytes []byte

func GenerateTapEvents(screenWidth int, screenHeight int) []Event {
	upTime := int64(TapDurationMinMs + rand.IntN(TapDurationJitterMs))
	return generateTapPair(screenWidth, screenHeight, upTime)
}

func GenerateLongTapEvents(screenWidth int, screenHeight int) []Event {
	upTime := int64(LongTapDurationMinMs + rand.IntN(LongTapDurationJitterMs))
	return generateTapPair(screenWidth, screenHeight, upTime)
}

func generateTapPair(screenWidth int, screenHeight int, upTime int64) []Event {
	return []Event{
		{
			Time: 0,
			Data: generateTouchData(ActionDown, screenWidth, screenHeight, PressureMax),
		},
		{
			Time: upTime,
			Data: generateTouchData(ActionUp, screenWidth, screenHeight, 0),
		},
	}
}

func generateTouchData(
	action byte,
	screenWidth int,
	screenHeight int,
	pressure uint16,
) ControlBytes {
	data := make(ControlBytes, ControlBytesSize)
	data[0] = TypeInjectTouchEvent
	data[1] = action
	binary.BigEndian.PutUint64(data[2:10], PointerIDGenericFinger)
	binary.BigEndian.PutUint16(data[18:20], uint16(screenWidth))
	binary.BigEndian.PutUint16(data[20:22], uint16(screenHeight))
	binary.BigEndian.PutUint16(data[22:24], pressure)
	return data
}

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
