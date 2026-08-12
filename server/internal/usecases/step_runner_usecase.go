package usecases

import (
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/core/numutils"
	"android_vision_scripter/pkg/core/strutils"
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
	"image"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"gocv.io/x/gocv"
)

func (i *interactorImpl) executeStep(serial string, step *models.Step) error {
	var session *models.Session
	if result, ok := i.sessionsCache.Get(serial); ok {
		session = &result
	}
	if session == nil || session.VideoPort == 0 {
		return fmt.Errorf("no connection to %s", serial)
	}
	if !step.Valid() {
		return errors.New("step is empty")
	}

	i.logger.Info(fmt.Sprintf("running step %s... ⏳", step.Describe()))

	err := i.runStepAction(serial, step)
	if err != nil {
		return err
	}

	i.logger.Info(fmt.Sprintf("step %s is COMPLETE ✅", step.Describe()))
	return nil
}

func (i *interactorImpl) runStepAction(serial string, step *models.Step) error {
	switch step.Action {
	case models.StepCheck:
		_, err := i.findTargetRect(serial, step.Target, step.GetTimeout())
		return err
	case models.StepTap:
		return i.playGeneratedTap(serial, step, false)
	case models.StepLongTap:
		return i.playGeneratedTap(serial, step, true)
	case models.StepTypeText:
		return i.typeTextStep(serial, step)
	}
	return i.playCustomAction(serial, step)
}

func (i *interactorImpl) playGeneratedTap(
	serial string,
	step *models.Step,
	longTap bool,
) error {
	foundRect, err := i.findTargetRect(serial, step.Target, step.GetTimeout())
	if err != nil {
		return err
	}

	width, height, err := i.waitForFrameSizes(serial)
	if err != nil {
		return err
	}

	events := models.GenerateTapEvents(width, height)
	if longTap {
		events = models.GenerateLongTapEvents(width, height)
	}
	i.playEvent(serial, foundRect, events)
	return nil
}

func (i *interactorImpl) typeTextStep(serial string, step *models.Step) error {
	var timeout = step.GetTimeout()
	if step.Target != nil {
		_, err := i.findTargetRect(serial, step.Target, timeout)
		if err != nil {
			return err
		}
	}

	width, height, err := i.waitForFrameSizes(serial)
	if err != nil {
		return err
	}

	tapEvents := models.GenerateTapEvents(width, height)
	return i.typeText(serial, timeout, step.Text, step.Locale, tapEvents)
}

func (i *interactorImpl) playCustomAction(serial string, step *models.Step) error {
	action, err := i.getAction(step.Action)
	if err != nil {
		return err
	}

	var foundRect *image.Rectangle
	if step.Target != nil {
		foundRect, err = i.findTargetRect(serial, step.Target, step.GetTimeout())
		if err != nil {
			return err
		}
	}

	i.playEvent(serial, foundRect, action.Events)
	time.Sleep(300 * time.Millisecond) //animation delay
	return nil
}

func (i *interactorImpl) findTargetRect(
	serial string,
	target *models.StepTarget,
	timeout time.Duration,
) (*image.Rectangle, error) {
	if !target.Valid() {
		return nil, errors.New("step target is empty")
	}

	var foundRect *image.Rectangle
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		mat, err := i.scrcpy.GetMatFromLastFrame(serial, true)
		if err != nil {
			i.logger.Error(err.Error())
			sleepUntilNextSecond()
			continue
		}
		if mat == nil {
			sleepUntilNextSecond()
			continue
		}

		foundRect, err = i.findRectByTarget(serial, mat, target)
		mat.Close()

		if err != nil {
			i.logger.Error(err.Error())
			sleepUntilNextSecond()
			continue
		}
		if models.ImageRectIsEmpty(foundRect) {
			sleepUntilNextSecond()
			continue
		}
		break
	}

	if models.ImageRectIsEmpty(foundRect) {
		return nil, fmt.Errorf(models.StatusError, target.Type, target.Value)
	}

	return foundRect, nil
}

func (i *interactorImpl) findRectByTarget(
	serial string,
	mat *gocv.Mat,
	target *models.StepTarget,
) (*image.Rectangle, error) {

	if target.Type == models.TargetImage {
		imagesDir := i.filesDB.CreateImagesDir()
		if imagesDir == "" {
			return nil, errors.New("images dir not found")
		}
		tmpImage := filepath.Join(imagesDir, strings.TrimSpace(target.Value)+ImageExt)
		if !file.Exists(tmpImage) {
			return nil, fmt.Errorf("image not found in library: %s", target.Value)
		}
		return i.cv.FindImage(mat, tmpImage)
	}

	if target.Type == models.TargetText {
		tesseractDir := i.filesDB.CreateLogsDir(serial, filesdb.TesseractDir)
		if tesseractDir == "" {
			return nil, fmt.Errorf("tesseract dir was not found")
		}
		var ocrParams = cv.InitOcrParams(
			target.Value,
			target.Locale,
			cv.PsmText,
			cv.OemText,
		)
		rectangles, err := i.cv.FindTextRectangles(mat, tesseractDir, ocrParams)
		if err != nil {
			return nil, err
		}
		if len(rectangles) == 0 || rectangles[0].IsEmpty() {
			return nil, nil
		}
		return rectangles[0].Rectangle.ToImageRectangle(), nil
	}

	if target.Type == models.TargetYolo {
		var labels = i.yolo.DetectLabels(*mat)
		for _, rect := range labels {
			if strings.EqualFold(rect.Label, target.Value) {
				return rect.ToImageRectangle(), nil
			}
		}
		return nil, fmt.Errorf("yolo class not found: %s", target.Value)
	}

	return nil, fmt.Errorf("unknown target type %s", target.Type)
}

func (i *interactorImpl) typeText(
	serial string,
	timeout time.Duration,
	text string,
	locale string,
	tapEvents []models.Event,
) error {
	if text == "" {
		return fmt.Errorf("nothing to type")
	}

	typed := false
	deadline := time.Now().Add(timeout)
attemptsLoop:
	for time.Now().Before(deadline) {
		keyboardKeys, err := i.getKeyboardKeys(serial, text, locale)
		if err != nil {
			i.logger.Error(err.Error())
			sleepUntilNextSecond()
			continue attemptsLoop
		}
		chars := []rune(strings.ToLower(text))
		keysToPress := make([]cv.OCRResult, len(chars))

	charsLoop:
		for index, ch := range chars {
			for _, key := range keyboardKeys {
				if key.Text == "" {
					return fmt.Errorf("empty keyboard key")
				}

				if []rune(key.Text)[0] == ch {
					keysToPress[index] = key
					continue charsLoop
				}
				if key.Text == cv.Space && unicode.IsSpace(ch) {
					keysToPress[index] = key
					continue charsLoop
				}
			}

			i.logger.Error(fmt.Sprintf("char %c not found", ch))
			sleepUntilNextSecond()
			continue attemptsLoop
		}

		for _, key := range keysToPress {
			var imgRect = key.Rectangle.ToImageRectangle()
			i.playEvent(serial, imgRect, tapEvents)
			time.Sleep(numutils.RandDelay(100, 300) * time.Millisecond)
		}

		typed = true
		break
	}

	if !typed {
		return fmt.Errorf("unable to type %q", text)
	}
	return nil
}

func (i *interactorImpl) getKeyboardKeys(
	serial string,
	text string,
	locale string,
) ([]cv.OCRResult, error) {
	keyboardKeys := []cv.OCRResult{}
	mat, err := i.scrcpy.GetMatFromLastFrame(serial, true)
	if err != nil {
		i.logger.Error(err.Error())
		return []cv.OCRResult{}, err
	}
	if mat == nil {
		return []cv.OCRResult{}, fmt.Errorf("mat is nil")
	}
	defer mat.Close()

	modelOS := i.GetDevice(serial).ToModelOs()
	keyboardDir := i.filesDB.CreateKeyboardDir(modelOS, locale)
	keyboardButtons := i.filesDB.GetFiles(keyboardDir)

	chars := strutils.GetUniqueChars(text)

	filteredButtons := []string{}
	for _, button := range keyboardButtons {
		for _, ch := range chars {
			if string(ch) == file.GetFileName(button) {
				filteredButtons = append(filteredButtons, button)
			}
		}
	}

	if len(filteredButtons) == 0 {
		return []cv.OCRResult{}, fmt.Errorf("buttons not found")
	}

	keyboardKeys = i.cv.GetKeyboardKeys(filteredButtons, *mat)
	return keyboardKeys, nil
}

func (i *interactorImpl) playEvent(
	serial string,
	rect *image.Rectangle,
	events []models.Event,
) {
	var offsetX = 0
	var offsetY = 0
	var lastTimeStamp int64
	for index, event := range events {
		var data = &event.Data
		if index == 0 {
			offsetX, offsetY = data.CountOffset(rect)
		}

		data.ApplyOffset(offsetX, offsetY)
		var delay = event.Time - lastTimeStamp
		time.Sleep(time.Duration(delay) * time.Millisecond)
		lastTimeStamp = event.Time

		i.scrcpy.WriteControlData(serial, *data)
	}
}

func sleepUntilNextSecond() {
	now := time.Now()
	next := now.Truncate(time.Second).Add(time.Second)
	time.Sleep(next.Sub(now))
}
