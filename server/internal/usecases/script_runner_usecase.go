package usecases

import (
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
	"image"
	"path/filepath"
	"time"
)

// Attempts to find template zone or text
const (
	Attempts = 3
)

func (i *interactorImpl) RunScript(serial string, scriptName string, socketPort int) error {
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
	started := i.StartScrcpyServer(serial, socketPort)
	if !started {
		var errMsg = fmt.Sprintf("couldn't start scrcpy server for %s", serial)
		return errors.New(errMsg)
	}

	go i.startVideoListener(serial, script)
	return nil
}

func (i *interactorImpl) startVideoListener(
	serial string,
	script *models.Script,
) {
	var doneCh = make(chan struct{})
	defer i.logger.Info(
		fmt.Sprintf(
			"closing scrcpy connection for %s... 🛑",
			serial,
		),
	)

	go func() {
		defer close(doneCh)
		i.executeScript(serial, script)
	}()

	<-doneCh
	i.CloseConnection(serial)
}

func (i *interactorImpl) executeScript(
	serial string,
	script *models.Script,
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

	defer i.logger.Info("closing script videostream... 🛑")
	go i.scrcpy.ReadVideoStream(serial, nil)

	for _, step := range script.Steps {
		if step.Flags == 0 {
			i.playEvent(serial, nil, &step)
		}

		if step.Flags&models.EventOnTemplate != 0 {
			err := i.playEventOnTemplate(serial, &step, scriptDir)
			if err != nil {
				i.logger.Error(err.Error())
				return
			}
		}

		if step.Flags&models.EventOnText != 0 {
			err := i.playEventOnText(serial, &step)
			if err != nil {
				i.logger.Error(err.Error())
				return
			}
		}

		time.Sleep(500 * time.Millisecond) //animation delay
	}
}

func (i *interactorImpl) playEventOnTemplate(
	serial string,
	step *models.ScriptStep,
	scriptDir string,
) error {
	var imgRect = &image.Rectangle{}
	tmpImage := filepath.Join(scriptDir, fmt.Sprintf("%d.png", step.ID))
	if !file.Exists(tmpImage) {
		return fmt.Errorf("template file not found: %s", tmpImage)
	}

	for range Attempts {
		mat, err := i.scrcpy.GetMatFromLastFrame(serial, true)
		if err != nil {
			i.logger.Error(err.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		if mat == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		imgRect, err = i.cv.FindImage(mat, tmpImage)
		if err != nil {
			i.logger.Error(err.Error())
			time.Sleep(1 * time.Second)
			continue
		}

		if models.ImageRectIsEmpty(imgRect) {
			continue
		}

		break
	}

	if models.ImageRectIsEmpty(imgRect) {
		return fmt.Errorf("template not found")
	}

	i.playEvent(serial, imgRect, step)
	return nil
}

func (i *interactorImpl) playEventOnText(
	serial string,
	step *models.ScriptStep,
) error {
	var textRect *image.Rectangle
	tesseractDir := i.filesDB.CreateLogsDir(serial, filesdb.TesseractDir)
	if tesseractDir == "" {
		return fmt.Errorf("tesseract dir was not found")
	}
	var device = i.GetDevice(serial)
	var ocrParams = cv.InitOcrParams(
		step.Text,
		device.Locale,
		cv.PsmText,
		cv.OemText,
		cv.WhiteTheme,
	)

	for range Attempts {
		mat, err := i.scrcpy.GetMatFromLastFrame(serial, true)
		if err != nil {
			i.logger.Error(err.Error())
			time.Sleep(1 * time.Second)
			continue
		}
		if mat == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		rectangles, err := i.cv.FindTextRectangles(
			mat,
			tesseractDir,
			ocrParams,
		)

		if err != nil {
			i.logger.Error(err.Error())
			time.Sleep(1 * time.Second)
			continue
		}

		if err == nil && len(rectangles) > 0 && !rectangles[0].IsEmpty() {
			textRect = rectangles[0].Rectangle.ToImageRectangle()
			break
		}
	}

	if models.ImageRectIsEmpty(textRect) {
		return fmt.Errorf("text %s not found", step.Text)
	}

	i.playEvent(serial, textRect, step)
	return nil
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
			offsetX, offsetY = i.countOffset(data, rect)
		}

		data.ApplyOffset(offsetX, offsetY)
		var delay = event.Time - lastTimeStamp
		time.Sleep(time.Duration(delay) * time.Millisecond)
		lastTimeStamp = event.Time

		i.scrcpy.WriteControlData(serial, *data)
	}
}

func (i *interactorImpl) countOffset(
	data *models.ControlBytes,
	stepZone *image.Rectangle,
) (int, int) {
	if models.ImageRectIsEmpty(stepZone) || data == nil {
		return 0, 0
	}

	x, y := data.GetXY()
	randX, randY := models.GetRandomXY(stepZone)
	return randX - x, randY - y
}
