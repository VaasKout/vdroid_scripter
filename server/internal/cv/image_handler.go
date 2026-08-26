package cv

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"sort"

	"gocv.io/x/gocv"
)

// ImageHandler ...
type ImageHandler interface {
	FindAllRectangles(img *gocv.Mat) ([]image.Rectangle, error)
	FindImages(img *gocv.Mat, template string) ([]image.Rectangle, error)
	DrawRectangles(
		img gocv.Mat,
		rectangles []image.Rectangle,
		transparent bool,
	) error
	CutZone(imgPath string, outputPath string, zone *image.Rectangle)
}

func (c *cvImpl) FindAllRectangles(img *gocv.Mat) ([]image.Rectangle, error) {
	if img == nil {
		return []image.Rectangle{}, errors.New("img empty")
	}
	imgRectangles, err := c.createRectangles(img)
	if err != nil {
		return []image.Rectangle{}, err
	}

	imgRectangles = c.filterRectangles(imgRectangles, img)
	return imgRectangles, nil
}

func (c *cvImpl) FindImages(
	img *gocv.Mat,
	template string,
) ([]image.Rectangle, error) {
	if img == nil || template == "" {
		return nil, errors.New("empty params")
	}

	templateMat := gocv.IMRead(template, gocv.IMReadColor)
	if templateMat.Empty() {
		return nil, errors.New("could not read template")
	}
	defer templateMat.Close()

	result := gocv.NewMat()
	defer result.Close()
	err := gocv.MatchTemplate(*img, templateMat, &result, gocv.TmCcoeffNormed, result)
	if err != nil {
		return nil, err
	}

	rectangles := []image.Rectangle{}
	for len(rectangles) < MaxTemplateMatches {
		_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)
		if maxVal < MatchCoefficient {
			break
		}

		rectangles = append(rectangles, image.Rect(
			maxLoc.X,
			maxLoc.Y,
			maxLoc.X+templateMat.Cols(),
			maxLoc.Y+templateMat.Rows(),
		))
		suppressMatch(&result, maxLoc, templateMat.Cols(), templateMat.Rows())
	}

	return rectangles, nil
}

func suppressMatch(result *gocv.Mat, loc image.Point, width int, height int) {
	x0 := max(loc.X-width/2, 0)
	y0 := max(loc.Y-height/2, 0)
	x1 := min(loc.X+width/2+1, result.Cols())
	y1 := min(loc.Y+height/2+1, result.Rows())

	region := result.Region(image.Rect(x0, y0, x1, y1))
	defer region.Close()
	region.SetTo(gocv.NewScalar(0, 0, 0, 0))
}

func (c *cvImpl) CutZone(imgPath string, outputPath string, zone *image.Rectangle) {
	img := gocv.IMRead(imgPath, gocv.IMReadColor)
	defer img.Close()
	if img.Empty() {
		return
	}

	cropped := img.Region(*zone)
	defer cropped.Close()
	gocv.IMWrite(outputPath, cropped)
}

func (c *cvImpl) createRectangles(img *gocv.Mat) ([]image.Rectangle, error) {
	if img == nil || img.Empty() {
		return []image.Rectangle{}, errors.New("img empty")
	}

	threshold := gocv.NewMat()
	defer threshold.Close()
	hierarchy := gocv.NewMat()
	defer hierarchy.Close()

	err := gocv.AdaptiveThreshold(
		*img,
		&threshold,
		255,
		gocv.AdaptiveThresholdGaussian, // Often better than Mean for icons
		gocv.ThresholdBinary,
		11, // Larger block size
		2,  // Higher C value
	)
	if err != nil {
		return []image.Rectangle{}, err
	}

	contours := gocv.FindContoursWithParams(
		threshold,
		&hierarchy,
		gocv.RetrievalList,
		gocv.ChainApproxSimple,
	)
	defer contours.Close()

	var rectangles = make([]image.Rectangle, contours.Size())
	for i := 0; i < contours.Size(); i++ {
		pvs := contours.At(i)
		rect := gocv.BoundingRect(pvs)
		rectangles[i] = rect
	}
	return rectangles, nil
}

func (c *cvImpl) filterRectangles(rects []image.Rectangle, img *gocv.Mat) []image.Rectangle {
	// sort rectangles from big to small
	sort.Slice(rects, func(i, j int) bool {
		areaI := rects[i].Dx() * rects[i].Dy()
		areaJ := rects[j].Dx() * rects[j].Dy()
		return areaI > areaJ
	})

	filtered := []image.Rectangle{}
	for _, rect := range rects {
		// remove too small rectangles and rectangles with borders on the edge of the screen
		area := float64(rect.Dx() * rect.Dy())
		if area < 1000 || rect.Size().X >= img.Cols() || rect.Size().Y >= img.Rows() {
			continue
		}

		shouldKeep := true
		for _, larger := range filtered {
			currentArea := rect.Dx() * rect.Dy()
			largerArea := larger.Dx() * larger.Dy()

			//remove rectangle only if it is smaller than any in filtered
			if currentArea < largerArea && isCloseToBorder(rect, larger) {
				shouldKeep = false
				break
			}
		}

		if shouldKeep {
			filtered = append(filtered, rect)
		}
	}

	return filtered
}

func isCloseToBorder(inner, outer image.Rectangle) bool {
	leftDist := inner.Min.X - outer.Min.X
	rightDist := outer.Max.X - inner.Max.X
	topDist := inner.Min.Y - outer.Min.Y
	bottomDist := outer.Max.Y - inner.Max.Y

	return ((leftDist >= 0 && leftDist <= MinBorderDistance) ||
		(rightDist >= 0 && rightDist <= MinBorderDistance) ||
		(topDist >= 0 && topDist <= MinBorderDistance) ||
		(bottomDist >= 0 && bottomDist <= MinBorderDistance)) && inner.Overlaps(outer)
}

func (c *cvImpl) DrawRectangles(
	img gocv.Mat,
	rectangles []image.Rectangle,
	transparent bool,
) error {
	if transparent {
		gocv.CvtColor(img, &img, gocv.ColorBGRToBGRA)
		img.SetTo(gocv.NewScalar(0, 0, 0, 0))
	}
	var redColor = color.RGBA{R: 255, A: 255}
	for _, rect := range rectangles {
		err := gocv.Rectangle(&img, rect, redColor, 2)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	return nil
}
