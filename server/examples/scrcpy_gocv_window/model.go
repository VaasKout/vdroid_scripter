package main

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
)

// Output files for cv
const (
	EdgesPng  = "edges.png"
	OcrJSON   = "ocr_results.json"
	OutputTsv = "output"
)

// Offset from rectangle borders
const (
	RectangleOffsetPercent = 20
	MinBorderDistance      = 20
)

// Rectangle ...
type Rectangle struct {
	LeftX   int `json:"left_x"`
	RightX  int `json:"right_x"`
	TopY    int `json:"top_y"`
	BottomY int `json:"bottom_y"`
}

// ToJSON ...
func (r *Rectangle) ToJSON() string {
	result, err := json.Marshal(r)
	if err != nil {
		fmt.Println("Rectangle ToJson " + err.Error())
		return ""
	}
	return string(result)
}

// Center ...
func (r *Rectangle) Center() (int, int) {
	return (r.LeftX + r.RightX) / 2, (r.TopY + r.BottomY) / 2
}

// Equals ...
func (r *Rectangle) Equals(rect *Rectangle) bool {
	return r.LeftX == rect.LeftX && r.RightX == rect.RightX && r.TopY == rect.TopY && r.BottomY == rect.BottomY
}

// IsNotEmpty ...
func (r *Rectangle) IsNotEmpty() bool {
	return r != nil && (r.LeftX > 0 || r.RightX > 0 || r.BottomY > 0 || r.TopY > 0)
}

// IsEmpty ...
func (r *Rectangle) IsEmpty() bool {
	return r.LeftX <= 0 && r.RightX <= 0 && r.BottomY <= 0 && r.TopY <= 0
}

// Contains ...
func (r *Rectangle) Contains(rect *image.Rectangle) bool {
	return (rect.Min.X >= r.LeftX && rect.Max.X <= r.LeftX && rect.Min.Y >= r.TopY && rect.Max.Y <= r.BottomY) ||
		(r.LeftX >= rect.Min.X && r.RightX <= rect.Max.X && r.TopY >= rect.Min.Y && r.BottomY <= rect.Max.Y)
}

// ClosestRect ...
func ClosestRect(rects []Rectangle, x, y int) int {
	targetX, targetY := x, y
	bestDist := math.MaxInt
	bestIndex := -1

	for i, r := range rects {
		cx, cy := r.Center()
		dx := cx - targetX
		dy := cy - targetY
		dist := dx*dx + dy*dy // squared distance
		if dist < bestDist {
			bestDist = dist
			bestIndex = i
		}
	}
	return bestIndex
}

// ImgRectanglesToDomain ...
func ImgRectanglesToDomain(imgRectangles []image.Rectangle) []Rectangle {
	var rectangles []Rectangle
	for _, imgRectangle := range imgRectangles {
		rectangles = append(rectangles, Rectangle{
			LeftX:   imgRectangle.Min.X,
			RightX:  imgRectangle.Max.X,
			TopY:    imgRectangle.Min.Y,
			BottomY: imgRectangle.Max.Y,
		})
	}
	return rectangles
}

// TemplateResult ...
type TemplateResult struct {
	Rectangle Rectangle
	Path      string
}
