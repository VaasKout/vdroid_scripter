package cv

import (
	"encoding/binary"
	"errors"
	"image"

	"gocv.io/x/gocv"
)

const (
	inkKernelDivisor  = 27
	inkMinKernel      = 3
	inkMinContrast    = 30
	solidReachDivisor = 80
	solidMinReach     = 2
	solidMaxShare     = 0.3
	labelSize         = 4
)

var fanDirections = []image.Point{
	{X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: -1, Y: 1},
	{X: -1, Y: 0}, {X: -1, Y: -1}, {X: 0, Y: -1}, {X: 1, Y: -1},
}

func (c *cvImpl) createTextEdges(img *gocv.Mat) (gocv.Mat, error) {
	if img.Empty() {
		return gocv.NewMat(), errors.New("createTextEdges img empty")
	}
	gray := gocv.NewMat()
	defer gray.Close()
	err := gocv.CvtColor(*img, &gray, gocv.ColorBGRToGray)
	if err != nil {
		return gocv.NewMat(), err
	}

	ink, err := localInk(&gray)
	defer ink.Close()
	if err != nil {
		return gocv.NewMat(), err
	}

	solid := solidRegions(&gray)
	defer solid.Close()
	err = dropSolidShapes(&ink, &solid)
	if err != nil {
		return gocv.NewMat(), err
	}

	edges := gocv.NewMat()
	gocv.BitwiseNot(ink, &edges)
	return edges, nil
}

func localInk(gray *gocv.Mat) (gocv.Mat, error) {
	ink := gocv.NewMat()
	kernel := oddKernel(gray.Cols())

	background := gocv.NewMat()
	defer background.Close()
	err := gocv.MedianBlur(*gray, &background, kernel)
	if err != nil {
		return ink, err
	}

	contrast := gocv.NewMat()
	defer contrast.Close()
	err = gocv.AbsDiff(*gray, background, &contrast)
	if err != nil {
		return ink, err
	}

	gocv.Threshold(contrast, &ink, inkMinContrast, MaxThreshHold, gocv.ThresholdBinary)
	return ink, nil
}

func oddKernel(width int) int {
	kernel := max(width/inkKernelDivisor, inkMinKernel)
	if kernel%2 == 0 {
		return kernel + 1
	}
	return kernel
}

func solidRegions(gray *gocv.Mat) gocv.Mat {
	reach := max(gray.Cols()/solidReachDivisor, solidMinReach)
	reaches := make([]gocv.Mat, 0, len(fanDirections))
	for _, direction := range fanDirections {
		reached := colorReaches(gray, direction, reach)
		reaches = append(reaches, reached)
	}

	solid := gocv.Zeros(gray.Rows(), gray.Cols(), gocv.MatTypeCV8UC1)
	fan := gocv.NewMat()
	defer fan.Close()
	for index, reached := range reaches {
		next := reaches[(index+1)%len(reaches)]
		afterNext := reaches[(index+2)%len(reaches)]
		gocv.BitwiseAnd(reached, next, &fan)
		gocv.BitwiseAnd(fan, afterNext, &fan)
		gocv.BitwiseOr(solid, fan, &solid)
	}

	for _, reached := range reaches {
		reached.Close()
	}
	return solid
}

func colorReaches(gray *gocv.Mat, direction image.Point, reach int) gocv.Mat {
	farOffset := direction.Mul(reach)
	nearOffset := direction.Mul(reach / 2)

	far := sameColorAt(gray, farOffset)
	near := sameColorAt(gray, nearOffset)
	defer near.Close()
	gocv.BitwiseAnd(far, near, &far)
	return far
}

func sameColorAt(gray *gocv.Mat, offset image.Point) gocv.Mat {
	same := gocv.Zeros(gray.Rows(), gray.Cols(), gocv.MatTypeCV8UC1)
	bounds := image.Rect(0, 0, gray.Cols(), gray.Rows())
	shifted := bounds.Sub(offset)
	here := bounds.Intersect(shifted)
	if here.Empty() {
		return same
	}
	there := here.Add(offset)

	hereRegion := gray.Region(here)
	defer hereRegion.Close()
	thereRegion := gray.Region(there)
	defer thereRegion.Close()
	contrast := gocv.NewMat()
	defer contrast.Close()
	gocv.AbsDiff(hereRegion, thereRegion, &contrast)

	sameRegion := same.Region(here)
	defer sameRegion.Close()
	gocv.Threshold(contrast, &sameRegion, inkMinContrast/2, MaxThreshHold, gocv.ThresholdBinaryInv)
	return same
}

func dropSolidShapes(ink *gocv.Mat, solid *gocv.Mat) error {
	labels := gocv.NewMat()
	defer labels.Close()
	count := gocv.ConnectedComponents(*ink, &labels)

	inkData, err := ink.DataPtrUint8()
	if err != nil {
		return err
	}
	solidData, err := solid.DataPtrUint8()
	if err != nil {
		return err
	}
	labelData := labels.ToBytes()

	sizes := make([]int, count)
	solidSizes := make([]int, count)
	for index, solidValue := range solidData {
		label := pixelLabel(labelData, index)
		sizes[label]++
		if solidValue != 0 {
			solidSizes[label]++
		}
	}

	for index := range inkData {
		label := pixelLabel(labelData, index)
		if label == 0 {
			continue
		}
		solidLimit := solidMaxShare * float64(sizes[label])
		if float64(solidSizes[label]) > solidLimit {
			inkData[index] = 0
		}
	}
	return nil
}

func pixelLabel(labelData []byte, index int) uint32 {
	offset := index * labelSize
	labelBytes := labelData[offset : offset+labelSize]
	return binary.NativeEndian.Uint32(labelBytes)
}
