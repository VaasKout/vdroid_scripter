package usecases

import (
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/core/numutils"
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
	"image"
	"strings"
	"time"
	"unicode"

	"gocv.io/x/gocv"
)

// Typing constants
const (
	KeyPressMinDelayMs  = 50
	KeyPressMaxDelayMs  = 150
	KeyboardSettleMs    = 400
	MaxKeyboardSwitches = 3
)

// StepsUseCase ...
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

	device, _ := i.devicesCache.Get(serial)
	if err := i.checkStepAssets(steps, device.ModelOs); err != nil {
		return err
	}
	models.FillStepIDs(steps)

	if i.GetSessionStatus(serial) == models.StatusRecording {
		return ErrRecordingInProgress
	}
	if err := i.ensureSessionIsRunning(serial, basePort); err != nil {
		return err
	}

	added := i.addStepsToQueue(serial, steps)
	if !added {
		return fmt.Errorf("no active session for %s", serial)
	}
	return nil
}

func (i *interactorImpl) checkStepAssets(steps []models.Step, tags ...string) error {
	actionsDir := i.filesDB.CreateActionsDir()
	imagesDir := i.filesDB.CreateImagesDir()
	for _, step := range steps {
		if step.IsCustomEvent() {
			if _, ok := libraryPath(actionsDir, step.Event, file.JSONExt, tags...); !ok {
				return libraryNotFound("event", step.Event, tags...)
			}
		}

		if step.Event == models.TypeTextEvent {
			locale := step.LastLandmark().Locale
			if _, ok := cv.KeyboardLayoutFor(locale); !ok {
				return fmt.Errorf("no keyboard layout for locale: %s", locale)
			}
			continue
		}

		for _, landmark := range step.Landmarks {
			if landmark.Type != models.Image {
				continue
			}
			if _, ok := libraryPath(imagesDir, landmark.Value, file.PngExt, tags...); !ok {
				return libraryNotFound("image", landmark.Value, tags...)
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

	i.logger.Info(fmt.Sprintf("running %s... ⏳", step.ToString()))

	time.Sleep(step.GetDelay())
	err := i.runStepAction(serial, step)
	if err != nil {
		return fmt.Errorf("step %d: %w", step.ID, err)
	}

	i.logger.Info(fmt.Sprintf("%s is COMPLETE ✅", step.ToString()))
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
	var stepZone *image.Rectangle
	if step.ValidLandmarks() {
		var err error
		stepZone, err = i.findRect(serial, step)
		if err != nil {
			return err
		}
	}

	width, height, err := i.scrcpy.GetScreenSize(serial)
	if err != nil {
		return err
	}

	direction, size := models.SwipeParts(step.Event)
	length := models.SwipeSizeLength(direction, size, width, height)
	events := models.GenerateSwipeEvents(direction, length, width, height)
	i.playEvent(serial, stepZone, events)
	return nil
}

func (i *interactorImpl) playGeneratedTap(
	serial string,
	step *models.Step,
	longTap bool,
) error {
	width, height, err := i.scrcpy.GetScreenSize(serial)
	if err != nil {
		return err
	}

	x, y := 0, 0
	var stepZone *image.Rectangle
	point := step.LastLandmark()
	isPoint := point != nil && point.Type == models.Point
	if isPoint {
		_, err = fmt.Sscanf(point.Value, "%d:%d", &x, &y)
		if err != nil || models.OutOfRange(x, y, width, height) {
			return fmt.Errorf(models.StatusError, point.Type, point.Value)
		}
	} else {
		stepZone, err = i.findRect(serial, step)
		if err != nil {
			return err
		}
	}

	events := models.GenerateTapEvents(x, y, width, height)
	if longTap {
		events = models.GenerateLongTapEvents(x, y, width, height)
	}
	i.playEvent(serial, stepZone, events)
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

	tapEvents := models.GenerateTapEvents(0, 0, width, height)
	return i.typeText(serial, step.GetTimeout(), last.Value, last.Locale, tapEvents)
}

func (i *interactorImpl) playCustomEvent(serial string, step *models.Step) error {
	action, err := i.getAction(serial, step.Event)
	if err != nil {
		return err
	}

	var stepZone *image.Rectangle
	if step.ValidLandmarks() {
		stepZone, err = i.findRect(serial, step)
		if err != nil {
			return err
		}
	}

	width, height, err := i.scrcpy.GetScreenSize(serial)
	if err != nil {
		return err
	}

	action.FitScreen(width, height)
	i.playEvent(serial, stepZone, action.Events)
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
	for {
		foundRect, err := i.findRectOnLastFrame(serial, step)
		if err == nil {
			return foundRect, nil
		}
		i.logger.Error(err.Error())
		if !time.Now().Before(deadline) {
			break
		}
	}

	last := step.LastLandmark()
	return nil, fmt.Errorf(models.StatusError, last.Type, last.Value)
}

func (i *interactorImpl) findRectOnLastFrame(
	serial string,
	step *models.Step,
) (*image.Rectangle, error) {
	mat, err := i.scrcpy.GetMatFromLastFrame(serial, true)
	if err != nil {
		return nil, err
	}
	if mat == nil {
		return nil, fmt.Errorf("no video frame received from %s", serial)
	}
	defer mat.Close()
	return i.findLandmarkChain(serial, mat, step)
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
		device, _ := i.devicesCache.Get(serial)
		imagesDir := i.filesDB.CreateImagesDir()
		tmpImage, ok := libraryPath(imagesDir, landmark.Value, file.PngExt, device.ModelOs)
		if !ok {
			return nil, libraryNotFound("image", landmark.Value, device.ModelOs)
		}
		scale := i.imageScale(serial, tmpImage)
		rectangles, err := i.cv.FindImages(mat, tmpImage, scale)
		if err != nil {
			return nil, err
		}
		models.SortRectsReadingOrder(rectangles)
		return rectangles, nil
	}

	if landmark.Type == models.Text {
		var ocrParams = cv.InitOcrParams(
			landmark.Value,
			landmark.Locale,
			cv.PsmText,
			cv.OemText,
		)
		ocrResults, err := i.cv.FindTextRectangles(mat, ocrParams)
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

	deadline := time.Now().Add(timeout)
	runes := []rune(text)
	var keyboard *models.Keyboard
	switches := 0
	index := 0
	for index < len(runes) {
		if keyboard == nil {
			detected, err := i.detectKeyboard(serial, locale, deadline)
			if err != nil {
				return err
			}
			keyboard = detected
		}

		ch := runes[index]
		if keyboard.ShiftNeeded(ch) {
			if !keyboard.HasShift() {
				return fmt.Errorf("shift key not found for %q", ch)
			}
			if switches >= MaxKeyboardSwitches {
				return fmt.Errorf("unable to switch shift for %q", ch)
			}
			i.pressKey(serial, keyboard.Shift, tapEvents)
			switches++
			keyboard = nil
			i.settleKeyboard()
			continue
		}

		key, err := keyboard.KeyToPress(ch)
		if err != nil {
			return err
		}
		i.pressKey(serial, key, tapEvents)
		switches = 0
		index++
		if unicode.IsUpper(ch) {
			keyboard = nil
			i.settleKeyboard()
		}
	}
	return nil
}

func (i *interactorImpl) pressKey(
	serial string,
	key image.Rectangle,
	tapEvents []models.Event,
) {
	i.playEvent(serial, &key, tapEvents)
	time.Sleep(numutils.RandDelay(KeyPressMinDelayMs, KeyPressMaxDelayMs) * time.Millisecond)
}

func (i *interactorImpl) settleKeyboard() {
	time.Sleep(KeyboardSettleMs * time.Millisecond)
}

func (i *interactorImpl) detectKeyboard(
	serial string,
	locale string,
	deadline time.Time,
) (*models.Keyboard, error) {
	for {
		keyboard, err := i.detectKeyboardOnLastFrame(serial, locale)
		if err == nil {
			return keyboard, nil
		}
		i.logger.Error(err.Error())
		if !time.Now().Before(deadline) {
			return nil, err
		}
	}
}

func (i *interactorImpl) detectKeyboardOnLastFrame(
	serial string,
	locale string,
) (*models.Keyboard, error) {
	mat, err := i.scrcpy.GetMatFromLastFrame(serial, true)
	if err != nil {
		return nil, err
	}
	if mat == nil {
		return nil, fmt.Errorf("no video frame received from %s", serial)
	}
	defer mat.Close()
	return i.cv.DetectKeyboard(mat, locale)
}

func (i *interactorImpl) playEvent(
	serial string,
	stepZone *image.Rectangle,
	events []models.Event,
) {
	offsetX, offsetY := touchOffset(stepZone, events)
	var lastTimeStamp int64
	for _, event := range events {
		var data = &event.Data
		data.ApplyOffset(offsetX, offsetY)
		var delay = event.Time - lastTimeStamp
		time.Sleep(time.Duration(delay) * time.Millisecond)
		lastTimeStamp = event.Time

		i.scrcpy.WriteControlData(serial, *data)
	}
}

func touchOffset(stepZone *image.Rectangle, events []models.Event) (int, int) {
	for _, event := range events {
		if event.Data.IsTouch() {
			return event.Data.CountOffset(stepZone)
		}
	}
	return 0, 0
}
