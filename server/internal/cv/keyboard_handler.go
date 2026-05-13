package cv

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"fmt"
	"image"
	"path/filepath"
	"strings"

	"gocv.io/x/gocv"
)

// Keyboard consts
const (
	SideMinDiff = -5
	SideMaxDiff = 5
)

// Keyboard special symbols
const (
	Space = "space"
)

var keyboardSizeMap = map[string]int{
	"nums": 10,
	"eng":  26,
	"rus":  32,
}

// KeyboardHandler ...
type KeyboardHandler interface {
	GetKeyboardKeys(
		keyboardButtons []string,
		img gocv.Mat,
	) []OCRResult
	ResetKeyboardKeys(
		keyboardDir string,
		tesseractDir string,
		screenshot string,
		lang string,
		upperCase bool,
	) []OCRResult
}

func (c *cvImpl) GetKeyboardKeys(
	keyboardButtons []string,
	img gocv.Mat,
) []OCRResult {
	var ocrResult = []OCRResult{}
	for _, buttonPath := range keyboardButtons {
		rect, err := c.FindImage(&img, buttonPath)
		if err != nil || models.ImageRectIsEmpty(rect) {
			continue
		}

		var imgRectangle = models.ImgRectangleToDomain(rect)
		if imgRectangle.IsEmpty() {
			continue
		}

		var result = OCRResult{
			Text:      file.GetFileName(buttonPath),
			Rectangle: *imgRectangle,
		}
		ocrResult = append(ocrResult, result)
	}
	return ocrResult
}

func (c *cvImpl) ResetKeyboardKeys(
	keyboardDir string,
	tesseractDir string,
	screenshot string,
	lang string,
	upperCase bool,
) []OCRResult {
	var result = []OCRResult{}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	defer img.Close()
	if img.Empty() {
		return result
	}

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

	rectangles, err := c.FindAllRectangles(&gray)
	if err != nil {
		c.logAPI.Error(err.Error())
		return result
	}

	rectMatrix := c.sortRectanglesBySize(rectangles)
	keyboardRects := c.findKeyboardRectangles(rectMatrix, lang)

	var ocrParams = InitOcrParams(
		"",
		lang,
		PsmChars,
		OemChars,
		BlackTheme,
	)

	for _, rect := range keyboardRects {
		cropped := img.Region(rect)
		ocrResult, err := c.FindTextRectangles(&cropped, tesseractDir, ocrParams)
		if err != nil || len(ocrResult) == 0 || ocrResult[0].Text == "" {
			cropped.Close()
			continue
		}

		text := ocrResult[0].Text
		if upperCase {
			text = strings.ToUpper(text)
		} else {
			text = strings.ToLower(text)
		}

		keyPath := filepath.Join(keyboardDir, fmt.Sprintf("%s.png", text))
		if file.Exists(keyPath) {
			continue
		}

		gocv.IMWrite(keyPath, cropped)
		cropped.Close()

		var domainRect = models.ImgRectangleToDomain(&rect)
		var outputOCRResult = OCRResult{Text: text, Rectangle: *domainRect}
		result = append(result, outputOCRResult)
	}

	return result
}

func (c *cvImpl) sortRectanglesBySize(rectangles []image.Rectangle) [][]image.Rectangle {
	var matrix = [][]image.Rectangle{}
mainLoop:
	for _, rect := range rectangles {
		if models.ImageRectIsEmpty(&rect) {
			continue
		}
		for index, line := range matrix {
			if len(line) > 0 {
				if c.ApproximatelyEqualRects(&line[0], &rect) {
					var updatedLine = append(line, rect)
					matrix[index] = updatedLine
					continue mainLoop
				}
			}
		}
		matrix = append(matrix, []image.Rectangle{rect})
	}
	return matrix
}

func (c *cvImpl) findKeyboardRectangles(
	matrix [][]image.Rectangle,
	key string,
) []image.Rectangle {
	var requiredSize = 0
	if result, ok := keyboardSizeMap[key]; ok {
		requiredSize = result
	} else {
		requiredSize = keyboardSizeMap[DefaultOCRLanguage]
	}

	var lines = [][]image.Rectangle{}
	for _, line := range matrix {
		if len(line) >= requiredSize {
			lines = append(lines, line)
		}
	}

	if len(lines) == 1 {
		return lines[0]
	}

	var lineIndexWithLowestY = 0 // keyboard is always in the bottom
	var lowestY = 0
	if len(lines) > 1 {
		for lineIndex, line := range lines {
			for _, rect := range line {
				if rect.Max.Y > lowestY {
					lineIndexWithLowestY = lineIndex
					lowestY = rect.Max.Y
				}
			}
		}
		return lines[lineIndexWithLowestY]
	}

	return []image.Rectangle{}
}

func (c *cvImpl) ApproximatelyEqualRects(rect1 *image.Rectangle, rect2 *image.Rectangle) bool {
	if models.ImageRectIsEmpty(rect1) || models.ImageRectIsEmpty(rect2) {
		return false
	}
	var width1 = rect1.Dx()
	var height1 = rect1.Dy()

	var width2 = rect2.Dx()
	var height2 = rect2.Dy()

	var widthDiff = width1 - width2
	var heightDiff = height1 - height2

	var isApproximateWidth = widthDiff >= SideMinDiff && widthDiff <= SideMaxDiff
	var isApproximateHeight = heightDiff >= SideMinDiff && heightDiff <= SideMaxDiff
	return isApproximateWidth && isApproximateHeight
}
