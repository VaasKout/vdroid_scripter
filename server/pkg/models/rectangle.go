package models

import (
	"android_vision_scripter/pkg/core/numutils"
	"image"
)

// Offset from rectangle borders
const (
	RectangleOffsetPercent = 20
)

// Rectangle ...
type Rectangle struct {
	LeftX   int    `json:"left_x"`
	RightX  int    `json:"right_x"`
	TopY    int    `json:"top_y"`
	BottomY int    `json:"bottom_y"`
	Label   string `json:"label,omitempty"`
}

// Center ...
func (r *Rectangle) Center() (int, int) {
	return (r.LeftX + r.RightX) / 2, (r.TopY + r.BottomY) / 2
}

// IsNotEmpty ...
func (r *Rectangle) IsNotEmpty() bool {
	return r != nil && (r.LeftX > 0 || r.RightX > 0 || r.BottomY > 0 || r.TopY > 0)
}

// IsEmpty ...
func (r *Rectangle) IsEmpty() bool {
	return r == nil || (r.LeftX <= 0 && r.RightX <= 0 && r.BottomY <= 0 && r.TopY <= 0)
}

// ImgRectangleToDomain ...
func ImgRectangleToDomain(imgRectangle *image.Rectangle) *Rectangle {
	if imgRectangle == nil {
		return &Rectangle{}
	}
	return &Rectangle{
		LeftX:   imgRectangle.Min.X,
		RightX:  imgRectangle.Max.X,
		TopY:    imgRectangle.Min.Y,
		BottomY: imgRectangle.Max.Y,
	}
}

// ImgRectanglesToDomain ...
func ImgRectanglesToDomain(imgRectangles []image.Rectangle) []Rectangle {
	var rectangles = make([]Rectangle, len(imgRectangles))
	for i, imgRectangle := range imgRectangles {
		rectangles[i] = Rectangle{
			LeftX:   imgRectangle.Min.X,
			RightX:  imgRectangle.Max.X,
			TopY:    imgRectangle.Min.Y,
			BottomY: imgRectangle.Max.Y,
		}
	}
	return rectangles
}

// ToImageRectangle ...
func (r *Rectangle) ToImageRectangle() *image.Rectangle {
	if r.IsEmpty() {
		return &image.Rectangle{}
	}
	return &image.Rectangle{
		Min: image.Point{
			X: r.LeftX,
			Y: r.TopY,
		},
		Max: image.Point{
			X: r.RightX,
			Y: r.BottomY,
		},
	}
}

// GetRandomXY ...
func GetRandomXY(r *image.Rectangle) (int, int) {
	var width = r.Dx()
	var height = r.Dy()

	var xOffset = numutils.GetPercentValue(width, RectangleOffsetPercent) / 2
	var yOffset = numutils.GetPercentValue(height, RectangleOffsetPercent) / 2

	var leftXWithOffset = r.Min.X + xOffset
	var rightXWithOffset = r.Max.X - xOffset
	var topYWithOffset = r.Min.Y + yOffset
	var bottomYWithOffset = r.Max.Y - yOffset

	var xToPress = 0
	var yToPress = 0

	if leftXWithOffset < rightXWithOffset {
		xToPress = numutils.RandInt(leftXWithOffset, rightXWithOffset)
	} else {
		xToPress = leftXWithOffset
	}

	if topYWithOffset < bottomYWithOffset {
		yToPress = numutils.RandInt(topYWithOffset, bottomYWithOffset)
	} else {
		yToPress = topYWithOffset
	}

	return xToPress, yToPress
}

// ImageRectIsEmpty ...
func ImageRectIsEmpty(r *image.Rectangle) bool {
	return r == nil || (r.Dy() == 0 && r.Dx() == 0)
}
