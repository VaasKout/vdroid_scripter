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

type StepsUseCase interface {
	RunSteps(serial string, steps []models.Step, basePort int) error
}

func (i *interactorImpl) RunSteps(
	serial string,
	steps []models.Step,
	basePort int,
) error {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return errors.New(SerialIsEmptyError)
	}
	if !models.ValidQueue(steps) {
		return errors.New("invalid steps")
	}

	if err := i.checkStepAssets(steps); err != nil {
		return err
	}

	if _, ok := i.sessionsCache.Get(serial); !ok {
		started := i.StartSession(serial, basePort)
		if !started {
			return fmt.Errorf("couldn't start scrcpy server for %s", serial)
		}
		go i.scrcpy.ReadVideoStream(serial, nil)
	}

	added := i.addStepsToQueue(serial, steps)
	if !added {
		return fmt.Errorf("no active session for %s", serial)
	}
	return nil
}

func (i *interactorImpl) checkStepAssets(steps []models.Step) error {
	for _, step := range steps {
		if step.IsCustomEvent() {
			actionPath := filepath.Join(
				i.filesDB.CreateActionsDir(),
				strings.TrimSpace(step.Event)+file.JsonExt,
			)
			if !file.Exists(actionPath) {
				return fmt.Errorf("event not found in library: %s", step.Event)
			}
		}

		if step.Event == models.TypeTextEvent {
			continue
		}

		for _, landmark := range step.Landmarks {
			if landmark.Type != models.Image {
				continue
			}

			imagePath := filepath.Join(
				i.filesDB.CreateImagesDir(),
				strings.TrimSpace(landmark.Value)+file.PngExt,
			)
			if !file.Exists(imagePath) {
				return fmt.Errorf("image not found in library: %s", landmark.Value)
			}
		}
	}
	return nil
}

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

	i.logger.Info(fmt.Sprintf("running step %s... ⏳", step.ToString()))

	err := i.runStepAction(serial, step)
	if err != nil {
		return err
	}

	i.logger.Info(fmt.Sprintf("step %s is COMPLETE ✅", step.ToString()))
	return nil
}

func (i *interactorImpl) runStepAction(serial string, step *models.Step) error {
	if step.IsCheckEvent() {
		_, err := i.findRect(serial, step)
		return err
	}

	switch step.Event {
	case models.TapEvent:
		return i.playGeneratedTap(serial, step, false)
	case models.LongTapEvent:
		return i.playGeneratedTap(serial, step, true)
	case models.TypeTextEvent:
		return i.typeTextStep(serial, step)
	}
	if step.IsSwipeEvent() {
		return i.playGeneratedSwipe(serial, step)
	}
	return i.playCustomEvent(serial, step)
}

func (i *interactorImpl) playGeneratedSwipe(serial string, step *models.Step) error {
	var foundRect *image.Rectangle
	if step.ValidLandmarks() {
		var err error
		foundRect, err = i.findRect(serial, step)
		if err != nil {
			return err
		}
	}

	width, height, err := i.scrcpy.GetScreenSize(serial)
	if err != nil {
		return err
	}

	events := models.GenerateSwipeEvents(step.Event, width, height)
	i.playEvent(serial, foundRect, events)
	time.Sleep(300 * time.Millisecond)
	return nil
}

func (i *interactorImpl) playGeneratedTap(
	serial string,
	step *models.Step,
	longTap bool,
) error {
	foundRect, err := i.findRect(serial, step)
	if err != nil {
		return err
	}

	width, height, err := i.scrcpy.GetScreenSize(serial)
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
	last := step.LastLandmark()
	if last == nil {
		return fmt.Errorf("nothing to type")
	}

	width, height, err := i.scrcpy.GetScreenSize(serial)
	if err != nil {
		return err
	}

	tapEvents := models.GenerateTapEvents(width, height)
	return i.typeText(serial, step.GetTimeout(), last.Value, last.Locale, tapEvents)
}

func (i *interactorImpl) playCustomEvent(serial string, step *models.Step) error {
	action, err := i.getAction(step.Event)
	if err != nil {
		return err
	}

	var foundRect *image.Rectangle
	if step.ValidLandmarks() {
		foundRect, err = i.findRect(serial, step)
		if err != nil {
			return err
		}
	}

	i.playEvent(serial, foundRect, action.Events)
	time.Sleep(300 * time.Millisecond) //animation delay
	return nil
}

func (i *interactorImpl) findRect(
	serial string,
	step *models.Step,
) (*image.Rectangle, error) {
	if !step.ValidLandmarks() {
		return nil, errors.New("step target is empty")
	}

	deadline := time.Now().Add(step.GetTimeout())
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

		foundRect, err := i.findLandmarkChain(serial, mat, step)
		mat.Close()

		if err != nil {
			i.logger.Error(err.Error())
			sleepUntilNextSecond()
			continue
		}
		return foundRect, nil
	}

	last := step.LastLandmark()
	return nil, fmt.Errorf(models.StatusError, last.Type, last.Value)
}

func (i *interactorImpl) findLandmarkChain(
	serial string,
	mat *gocv.Mat,
	step *models.Step,
) (*image.Rectangle, error) {
	var prevRect *image.Rectangle
	for _, landmark := range step.Landmarks {
		candidates, err := i.findLandmarkCandidates(serial, mat, &landmark)
		if err != nil {
			return nil, err
		}

		foundRect := models.ClosestRect(candidates, prevRect)
		if models.ImageRectIsEmpty(foundRect) {
			return nil, fmt.Errorf(models.StatusError, landmark.Type, landmark.Value)
		}
		prevRect = foundRect
	}
	return prevRect, nil
}

func (i *interactorImpl) findLandmarkCandidates(
	serial string,
	mat *gocv.Mat,
	landmark *models.Landmark,
) ([]image.Rectangle, error) {
	if landmark.Type == models.Image {
		imagesDir := i.filesDB.CreateImagesDir()
		if imagesDir == "" {
			return nil, errors.New("images dir not found")
		}
		tmpImage := filepath.Join(imagesDir, strings.TrimSpace(landmark.Value)+file.PngExt)
		if !file.Exists(tmpImage) {
			return nil, fmt.Errorf("image not found in library: %s", landmark.Value)
		}
		rectangles, err := i.cv.FindImages(mat, tmpImage)
		if err != nil {
			return nil, err
		}
		models.SortRectsReadingOrder(rectangles)
		return rectangles, nil
	}

	if landmark.Type == models.Text {
		tesseractDir := i.filesDB.CreateLogsDir(serial, filesdb.TesseractDir)
		if tesseractDir == "" {
			return nil, fmt.Errorf("tesseract dir was not found")
		}
		var ocrParams = cv.InitOcrParams(
			landmark.Value,
			landmark.Locale,
			cv.PsmText,
			cv.OemText,
		)
		ocrResults, err := i.cv.FindTextRectangles(mat, tesseractDir, ocrParams)
		if err != nil {
			return nil, err
		}

		rectangles := []image.Rectangle{}
		for _, result := range ocrResults {
			if result.IsEmpty() {
				continue
			}
			rectangles = append(rectangles, *result.Rectangle.ToImageRectangle())
		}
		models.SortRectsReadingOrder(rectangles)
		return rectangles, nil
	}

	if landmark.Type == models.Yolo {
		rectangles := []image.Rectangle{}
		for _, detection := range i.yolo.DetectLabels(*mat) {
			if strings.EqualFold(detection.Label, landmark.Value) {
				rectangles = append(rectangles, *detection.ToImageRectangle())
			}
		}
		models.SortRectsReadingOrder(rectangles)
		return rectangles, nil
	}

	return nil, fmt.Errorf("unknown target type %s", landmark.Type)
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
