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

const (
	Timeout = 15
)

func (i *interactorImpl) RunScript(serial string, scriptName string, basePort int) error {
	if serial == "" {
		return errors.New(SerialIsEmptyError)
	}
	if scriptName == "" {
		return errors.New(ScriptNameIsEmpty)
	}

	path := i.getScriptRunner(serial, scriptName)
	if path == "" {
		return errors.New("script not found")
	}

	script := i.getScriptFromFile(path)
	if script == nil || len(script.Steps) == 0 {
		return errors.New("script steps are empty")
	}

	var newConnection = false
	if _, ok := i.clientsCache.Get(serial); !ok {
		started := i.StartScrcpyServer(serial, basePort)
		if !started {
			var errMsg = fmt.Sprintf("couldn't start scrcpy server for %s", serial)
			return errors.New(errMsg)
		}
		newConnection = started
	}

	go i.startVideoListener(serial, script, newConnection)
	return nil
}

func (i *interactorImpl) startVideoListener(
	serial string,
	script *models.Script,
	newConnection bool,
) {
	var doneCh = make(chan struct{})

	go func() {
		defer close(doneCh)
		i.executeScript(serial, script, newConnection)
	}()

	<-doneCh
	if newConnection {
		i.CloseConnection(serial)
	}
}

func (i *interactorImpl) executeScript(
	serial string,
	script *models.Script,
	newConnection bool,
) {
	var clientConnection *ClientConnection
	if result, ok := i.clientsCache.Get(serial); ok {
		clientConnection = &result
	}
	if clientConnection == nil || clientConnection.VideoPort == 0 {
		i.logger.Error(fmt.Sprintf("no connection to %s", serial))
		return
	}
	if script == nil || script.Name == "" {
		i.logger.Error("script is empty")
		return
	}

	scriptDir := i.getScriptDir(serial, script.Name)
	if scriptDir == "" {
		i.logger.Error(fmt.Sprintf("scriptDir not found %s", script.Name))
		return
	}

	if newConnection {
		go i.scrcpy.ReadVideoStream(serial, nil)
	}

	for _, step := range script.Steps {
		i.logger.Info(fmt.Sprintf("running step %d... ⏳", step.ID))
		if step.Flags == 0 {
			i.playEvent(serial, nil, &step)
		}

		err := i.playEventOnRect(serial, &step, scriptDir)
		if err != nil {
			i.logger.Error(err.Error())
			return
		}

		if step.HasFlag(models.TypeText) {
			err := i.typeText(serial, &step)
			if err != nil {
				i.logger.Error(err.Error())
				return
			}
		}

		time.Sleep(300 * time.Millisecond) //animation delay
	}

	i.logger.Info(fmt.Sprintf("script %s is COMPLETE ✅", script.Name))
}

func (i *interactorImpl) playEventOnRect(
	serial string,
	step *models.ScriptStep,
	scriptDir string,
) error {
	var foundRect *image.Rectangle
	deadline := time.Now().Add(Timeout * time.Second)
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

		foundRect, err = i.findRectByFlag(serial, mat, step, scriptDir)
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
		return fmt.Errorf("rectangle for step %d not found", step.ID)
	}

	if step.HasEventFlag() {
		i.playEvent(serial, foundRect, step)
	}
	return nil
}

func (i *interactorImpl) findRectByFlag(
	serial string,
	mat *gocv.Mat,
	step *models.ScriptStep,
	scriptDir string,
) (*image.Rectangle, error) {
	if step.HasAnyTemplateFlags() {
		tmpImage := filepath.Join(scriptDir, fmt.Sprintf("%d.png", step.ID))
		if !file.Exists(tmpImage) {
			return nil, fmt.Errorf("template file not found: %s", tmpImage)
		}
		return i.cv.FindImage(mat, tmpImage)
	}

	if step.HasAnyTextFlags() {
		tesseractDir := i.filesDB.CreateLogsDir(serial, filesdb.TesseractDir)
		if tesseractDir == "" {
			return nil, fmt.Errorf("tesseract dir was not found")
		}
		var ocrParams = cv.InitOcrParams(
			step.Text,
			step.Locale,
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

	if step.HasAnyYoloFlags() {
		var labels = i.yolo.DetectLabels(*mat)
		for _, rect := range labels {
			if strings.EqualFold(rect.Label, step.Text) {
				return rect.ToImageRectangle(), nil
			}
		}
		return nil, fmt.Errorf("yolo class not found: %s", step.Text)
	}

	return nil, fmt.Errorf("unknown flag %d", step.Flags)
}

func (i *interactorImpl) typeText(
	serial string,
	step *models.ScriptStep,
) error {
	if step == nil || step.Text == "" {
		return fmt.Errorf("nothing to type")
	}

	deadline := time.Now().Add(Timeout * time.Second)
attemptsLoop:
	for time.Now().Before(deadline) {
		keyboardKeys, err := i.getKeyboardKeys(serial, step)
		if err != nil {
			i.logger.Error(err.Error())
			sleepUntilNextSecond()
			continue attemptsLoop
		}
		chars := []rune(strings.ToLower(step.Text))
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
			i.playEvent(serial, imgRect, step)
			time.Sleep(numutils.RandDelay(100, 300) * time.Millisecond)
		}

		break
	}
	return nil
}

func (i *interactorImpl) getKeyboardKeys(
	serial string,
	step *models.ScriptStep,
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

	modelOs := i.GetDevice(serial).ToModelOs()
	keyboardDir := i.filesDB.CreateDBDir(modelOs, filesdb.Keyboards, step.Locale)
	keyboardButtons := i.filesDB.GetFiles(keyboardDir)

	chars := strutils.GetUniqueChars(step.Text)

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
	step *models.ScriptStep,
) {
	var offsetX = 0
	var offsetY = 0
	var lastTimeStamp int64
	for index, event := range step.Events {
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
