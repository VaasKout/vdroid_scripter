package usecases

import (
	"android_vision_scripter/internal/cv"
	"strings"

	"gocv.io/x/gocv"
)

// OcrUseCase ...
type OcrUseCase interface {
	FindText(serial string, text string, locale string) []cv.OCRResult
}

func (i *interactorImpl) FindText(
	serial string,
	text string,
	locale string,
) []cv.OCRResult {
	text = strings.TrimSpace(text)
	if text == "" {
		return []cv.OCRResult{}
	}

	screenshot := i.cmd.ScreenShot(serial)
	if screenshot == "" {
		return []cv.OCRResult{}
	}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	defer img.Close()
	if img.Empty() {
		return []cv.OCRResult{}
	}

	var ocrParams = cv.InitOcrParams(
		text,
		locale,
		cv.PsmText,
		cv.OemText,
	)

	result, err := i.cv.FindTextRectangles(&img, ocrParams)
	if err != nil {
		i.logger.Error(err.Error())
		return []cv.OCRResult{}
	}
	return result
}
