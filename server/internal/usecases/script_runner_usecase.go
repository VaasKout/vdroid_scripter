package usecases

import (
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/core/numutils"
	"android_vision_scripter/pkg/core/strutils"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"gocv.io/x/gocv"
)

func (i *interactorImpl) RunScripts(
	serial string,
	scripts []ScriptRef,
	basePort int,
) error {
	if serial == "" {
		return errors.New(SerialIsEmptyError)
	}
	if len(scripts) == 0 {
		return errors.New(ScriptNameIsEmpty)
	}

	var entries = make([]string, 0, len(scripts))
	for _, ref := range scripts {
		if ref.Location == "" || ref.Name == "" {
			return errors.New("script location and name are required")
		}

		path := i.getScriptRunner(ref.Location, ref.Name)
		if path == "" {
			return fmt.Errorf("script not found: %s/%s", ref.Location, ref.Name)
		}

		script := i.getScriptFromFile(path)
		if script == nil || script.Name == "" {
			return fmt.Errorf("script is empty: %s/%s", ref.Location, ref.Name)
		}
		entries = append(entries, ref.Location+"/"+ref.Name)
	}

	if _, ok := i.sessionsCache.Get(serial); !ok {
		started := i.StartSession(serial, basePort)
		if !started {
			var errMsg = fmt.Sprintf("couldn't start scrcpy server for %s", serial)
			return errors.New(errMsg)
		}
		go i.scrcpy.ReadVideoStream(serial, nil)
	}

	i.addScriptsToQueue(serial, entries)
	return nil
}

func (i *interactorImpl) executeScript(
	serial string,
	script *models.Script,
) error {
	var session *models.Session
	if result, ok := i.sessionsCache.Get(serial); ok {
		session = &result
	}
	if session == nil || session.VideoPort == 0 {
		return fmt.Errorf("no connection to %s", serial)
	}
	if script == nil || script.Name == "" {
		return errors.New("script is empty")
	}

	scriptDir := i.filesDB.CreateScriptDir(script.Location, script.Name)
	if scriptDir == "" {
		return fmt.Errorf("scriptDir not found %s", script.Name)
	}

	i.logger.Info(fmt.Sprintf("running script %s... ⏳", script.Name))

	if len(script.Params) == 0 && len(script.Events) > 0 {
		i.playEvent(serial, nil, script.Events)
		time.Sleep(300 * time.Millisecond) //animation delay
	}

	var timeout = script.GetTimeout()
	for index, param := range script.Params {
		if param.Type == models.TypeText {
			err := i.typeText(serial, timeout, &param, script.Events)
			if err != nil {
				return paramNotFoundError(&param, script)
			}
			continue
		}

		if index < len(script.Params)-1 || len(script.Events) == 0 {
			foundRect, err := i.findRectangle(serial, &param, timeout, scriptDir)
			if err != nil || foundRect == nil {
				return paramNotFoundError(&param, script)
			}
		}

		if len(script.Events) > 0 && index == len(script.Params)-1 {
			foundRect, err := i.findRectangle(serial, &param, timeout, scriptDir)
			if err != nil {
				return paramNotFoundError(&param, script)
			}
			i.playEvent(serial, foundRect, script.Events)
		}
	}

	i.logger.Info(fmt.Sprintf("script %s is COMPLETE ✅", script.Name))
	return nil
}

func paramNotFoundError(param *models.Parameter, script *models.Script) error {
	var scriptPath = script.Location + "/" + script.Name
	return fmt.Errorf(models.StatusError, param.Type, param.Value, scriptPath)
}

func (i *interactorImpl) findRectangle(
	serial string,
	param *models.Parameter,
	timeout time.Duration,
	scriptDir string,
) (*image.Rectangle, error) {
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

		foundRect, err = i.findRectByFlag(serial, mat, param, scriptDir)
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
		return nil, fmt.Errorf("rectangle for step %s not found", param.Value)
	}

	return foundRect, nil
}

func (i *interactorImpl) findRectByFlag(
	serial string,
	mat *gocv.Mat,
	param *models.Parameter,
	scriptDir string,
) (*image.Rectangle, error) {

	if param.Type == models.Template {
		tmpImage := filepath.Join(scriptDir, fmt.Sprintf("%s.png", param.Value))
		if !file.Exists(tmpImage) {
			return nil, fmt.Errorf("template file not found: %s", tmpImage)
		}
		return i.cv.FindImage(mat, tmpImage)
	}

	if param.Type == models.Text {
		tesseractDir := i.filesDB.CreateLogsDir(serial, filesdb.TesseractDir)
		if tesseractDir == "" {
			return nil, fmt.Errorf("tesseract dir was not found")
		}
		var ocrParams = cv.InitOcrParams(
			param.Value,
			param.Locale,
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

	if param.Type == models.YoloClass {
		var labels = i.yolo.DetectLabels(*mat)
		for _, rect := range labels {
			if strings.EqualFold(rect.Label, param.Value) {
				return rect.ToImageRectangle(), nil
			}
		}
		return nil, fmt.Errorf("yolo class not found: %s", param.Value)
	}

	return nil, fmt.Errorf("unknown type %s", param.Type)
}

func (i *interactorImpl) typeText(
	serial string,
	timeout time.Duration,
	param *models.Parameter,
	events []models.Event,
) error {
	if param == nil || param.Value == "" {
		return fmt.Errorf("nothing to type")
	}

	tapEvents := models.ExtractTapEvents(events)
	deadline := time.Now().Add(timeout)
attemptsLoop:
	for time.Now().Before(deadline) {
		keyboardKeys, err := i.getKeyboardKeys(serial, param)
		if err != nil {
			i.logger.Error(err.Error())
			sleepUntilNextSecond()
			continue attemptsLoop
		}
		chars := []rune(strings.ToLower(param.Value))
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

		break
	}
	return nil
}

func (i *interactorImpl) getKeyboardKeys(
	serial string,
	param *models.Parameter,
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
	keyboardDir := i.filesDB.CreateKeyboardDir(modelOS, param.Locale)
	keyboardButtons := i.filesDB.GetFiles(keyboardDir)

	chars := strutils.GetUniqueChars(param.Value)

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

func (i *interactorImpl) getScriptFromFile(filePath string) *models.Script {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return &models.Script{}
	}
	var script = &models.Script{}
	_ = json.Unmarshal(bytes, script)
	return script
}
