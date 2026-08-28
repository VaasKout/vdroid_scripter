package models

import (
	"encoding/binary"
	"encoding/json"
	"image"
	"math"
	"math/rand/v2"
)

// Size of control bytes buffer
const (
	ControlBytesSize = 32
)

// Touch event properties
const (
	TypeInjectTouchEvent byte = 2

	PointerIDGenericFinger uint64 = 0xFFFFFFFFFFFFFFFE
	PressureMax            uint16 = 0xFFFF

	TapDurationMinMs    = 50
	TapDurationJitterMs = 70

	LongTapDurationMinMs    = 700
	LongTapDurationJitterMs = 200

	SwipeDurationMinMs    = 300
	SwipeDurationJitterMs = 200
	SwipeMoveIntervalMs   = 12
	SwipeLengthRatio      = 0.5
	SwipeEdgeMarginRatio  = 0.1
	SwipeBowRatioMin      = 0.02
	SwipeBowRatioMax      = 0.04
	SwipeJitterPx         = 2
)

// Touch actions
const (
	ActionDown byte = 0
	ActionUp   byte = 1
	ActionMove byte = 2
)

// Event ...
type Event struct {
	Time int64        `json:"time"`
	Data ControlBytes `json:"data"`
}

// ControlBytes ...
type ControlBytes []byte

// GenerateTapEvents ...
func GenerateTapEvents(screenWidth int, screenHeight int) []Event {
	upTime := int64(TapDurationMinMs + rand.IntN(TapDurationJitterMs))
	return generateTapPair(screenWidth, screenHeight, upTime)
}

// GenerateLongTapEvents ...
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

// GenerateSwipeEvents ...
func GenerateSwipeEvents(direction string, screenWidth int, screenHeight int) []Event {
	start, end := swipeEndpoints(direction, screenWidth, screenHeight)
	control := swipeControlPoint(start, end)
	duration := int64(SwipeDurationMinMs + rand.IntN(SwipeDurationJitterMs))

	events := []Event{{
		Time: 0,
		Data: generatePositionedTouchData(
			ActionDown, start.X, start.Y, screenWidth, screenHeight, PressureMax,
		),
	}}

	for t := int64(SwipeMoveIntervalMs); t < duration; t += SwipeMoveIntervalMs {
		progress := easeInOut(float64(t) / float64(duration))
		point := bezierPoint(start, control, end, progress)
		events = append(events, Event{
			Time: t,
			Data: generatePositionedTouchData(
				ActionMove,
				jitterCoord(point.X),
				jitterCoord(point.Y),
				screenWidth, screenHeight, PressureMax,
			),
		})
	}

	return append(events, Event{
		Time: duration,
		Data: generatePositionedTouchData(
			ActionUp, end.X, end.Y, screenWidth, screenHeight, 0,
		),
	})
}

func swipeEndpoints(direction string, screenWidth int, screenHeight int) (image.Point, image.Point) {
	marginX := int(float64(screenWidth) * SwipeEdgeMarginRatio)
	marginY := int(float64(screenHeight) * SwipeEdgeMarginRatio)
	lengthX := int(float64(screenWidth) * SwipeLengthRatio)
	lengthY := int(float64(screenHeight) * SwipeLengthRatio)

	switch direction {
	case SwipeUpEvent:
		x := randRange(marginX, screenWidth-marginX)
		y := randRange(marginY+lengthY, screenHeight-marginY)
		return image.Pt(x, y), image.Pt(x, y-lengthY)
	case SwipeDownEvent:
		x := randRange(marginX, screenWidth-marginX)
		y := randRange(marginY, screenHeight-marginY-lengthY)
		return image.Pt(x, y), image.Pt(x, y+lengthY)
	case SwipeLeftEvent:
		x := randRange(marginX+lengthX, screenWidth-marginX)
		y := randRange(marginY, screenHeight-marginY)
		return image.Pt(x, y), image.Pt(x-lengthX, y)
	}
	x := randRange(marginX, screenWidth-marginX-lengthX)
	y := randRange(marginY, screenHeight-marginY)
	return image.Pt(x, y), image.Pt(x+lengthX, y)
}

func swipeControlPoint(start image.Point, end image.Point) image.Point {
	midX := (start.X + end.X) / 2
	midY := (start.Y + end.Y) / 2

	dx := float64(end.X - start.X)
	dy := float64(end.Y - start.Y)
	length := math.Hypot(dx, dy)
	if length == 0 {
		return image.Pt(midX, midY)
	}

	bowRatio := SwipeBowRatioMin + rand.Float64()*(SwipeBowRatioMax-SwipeBowRatioMin)
	bow := bowRatio * length
	if rand.IntN(2) == 0 {
		bow = -bow
	}
	return image.Pt(
		midX+int(-dy/length*bow),
		midY+int(dx/length*bow),
	)
}

func bezierPoint(start image.Point, control image.Point, end image.Point, t float64) image.Point {
	inv := 1 - t
	x := inv*inv*float64(start.X) + 2*inv*t*float64(control.X) + t*t*float64(end.X)
	y := inv*inv*float64(start.Y) + 2*inv*t*float64(control.Y) + t*t*float64(end.Y)
	return image.Pt(int(math.Round(x)), int(math.Round(y)))
}

func easeInOut(progress float64) float64 {
	return progress * progress * (3 - 2*progress)
}

func jitterCoord(value int) int {
	jittered := value + rand.IntN(2*SwipeJitterPx+1) - SwipeJitterPx
	if jittered < 0 {
		return 0
	}
	return jittered
}

func randRange(minValue int, maxValue int) int {
	if maxValue <= minValue {
		return minValue
	}
	return minValue + rand.IntN(maxValue-minValue)
}

func generatePositionedTouchData(
	action byte,
	x int,
	y int,
	screenWidth int,
	screenHeight int,
	pressure uint16,
) ControlBytes {
	data := generateTouchData(action, screenWidth, screenHeight, pressure)
	binary.BigEndian.PutUint32(data[10:14], uint32(x))
	binary.BigEndian.PutUint32(data[14:18], uint32(y))
	return data
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

// CountOffset ...
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
