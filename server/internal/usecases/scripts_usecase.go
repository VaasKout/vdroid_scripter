package usecases

import (
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gocv.io/x/gocv"
)

// Common script errors
const (
	ScriptNameIsEmpty = "script name is empty"
)

// Themes
const (
	WhiteTheme = "white"
	BlackTheme = "black"
)

// ScriptsUseCase ...
type ScriptsUseCase interface {
	GetScriptNames(serial string) ([]string, error)
	GetScript(serial string, scriptName string) (*models.Script, error)
	DeleteScript(serial string, scriptName string) error
	RunScript(serial string, scriptName string, basePort int) error

	SaveZone(serial string, zone *models.Rectangle) bool
	SaveStep(
		serial string,
		name string,
		step *models.ScriptStep,
	) bool
	FindText(serial string, text string, locale string) []cv.OCRResult
}

func (i *interactorImpl) SaveZone(serial string, zone *models.Rectangle) bool {
	if zone.IsEmpty() {
		return false
	}
	var dbDir = i.filesDB.CreateDBDir(serial)
	var tmpImg = filepath.Join(dbDir, filesdb.TmpZone)
	created := file.CreateFileIfNotExist(tmpImg)
	if !created {
		return false
	}

	screenShot := i.cmd.ScreenShot(serial)
	if screenShot == "" {
		return false
	}
	imgRect := zone.ToImageRectangle()
	i.cv.CutZone(screenShot, tmpImg, imgRect)
	return true
}

func (i *interactorImpl) SaveStep(
	serial string,
	name string,
	step *models.ScriptStep,
) bool {
	if step == nil || serial == "" || name == "" {
		return false
	}

	runnerPath := i.getScriptRunner(serial, name)
	if runnerPath == "" {
		return false
	}

	script := i.getScriptFromFile(runnerPath)
	script.Name = name

	var id = len(script.Steps) + 1
	step.ID = id

	if step.Flags&models.EventOnTemplate != 0 {
		scriptDir := i.getScriptDir(serial, name)
		tmpImg := i.filesDB.FindFileInDBDir(serial, filesdb.TmpZone)
		if tmpImg != "" {
			newImagePath := filepath.Join(scriptDir, fmt.Sprintf("%d.png", step.ID))
			os.Rename(tmpImg, newImagePath)
		}
	}

	script.Steps = append(script.Steps, *step)
	return i.saveScriptInFile(script, runnerPath)
}
func (i *interactorImpl) FindText(
	serial string,
	text string,
	locale string,
) []cv.OCRResult {
	dir := i.filesDB.CreateLogsDir(serial, filesdb.TesseractDir)
	if dir == "" {
		return []cv.OCRResult{}
	}

	screenshot := i.cmd.ScreenShot(serial)
	if screenshot == "" {
		return []cv.OCRResult{}
	}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		return []cv.OCRResult{}
	}
	defer img.Close()

	var ocrParams = cv.InitOcrParams(
		text,
		locale,
		cv.PsmText,
		cv.OemText,
	)

	result, err := i.cv.FindTextRectangles(&img, dir, ocrParams)
	if err != nil {
		i.logger.Error(err.Error())
		return []cv.OCRResult{}
	}
	return result
}

func (i *interactorImpl) GetScriptNames(serial string) ([]string, error) {
	if serial == "" {
		return []string{}, errors.New(SerialIsEmptyError)
	}
	var device = i.GetDevice(serial)
	var modelOS = device.ToModelOs()
	if modelOS == "" {
		return []string{}, errors.New(DeviceNotFoundError)
	}
	scriptsDir := i.filesDB.CreateDBDir(modelOS, filesdb.ScriptsDir)
	dirs := i.filesDB.GetDirs(scriptsDir)
	names := make([]string, len(dirs))
	for index, script := range dirs {
		names[index] = filepath.Base(script)
	}
	return names, nil
}

func (i *interactorImpl) GetScript(serial string, scriptName string) (*models.Script, error) {
	if serial == "" {
		return &models.Script{}, errors.New(SerialIsEmptyError)
	}
	if scriptName == "" {
		return &models.Script{}, errors.New(ScriptNameIsEmpty)
	}

	runnerPath := i.getScriptRunner(serial, scriptName)
	if runnerPath == "" {
		var errStr = fmt.Sprintf("unable to create json runner %s for %s", scriptName, serial)
		return &models.Script{}, errors.New(errStr)
	}

	bytes, err := os.ReadFile(runnerPath)
	if err != nil {
		return &models.Script{}, err
	}
	var script = &models.Script{}
	_ = json.Unmarshal(bytes, script)
	return script, nil
}

func (i *interactorImpl) DeleteScript(serial string, scriptName string) error {
	if serial == "" {
		return errors.New(SerialIsEmptyError)
	}
	if scriptName == "" {
		return errors.New(ScriptNameIsEmpty)
	}
	var device = i.GetDevice(serial)
	var modelOS = device.ToModelOs()
	if modelOS == "" {
		return errors.New(DeviceNotFoundError)
	}

	scriptsDir := i.filesDB.CreateDBDir(modelOS, filesdb.ScriptsDir)
	i.filesDB.DeleteDirByName(scriptsDir, scriptName)
	return nil
}

func (i *interactorImpl) getScriptRunner(
	serial string,
	scriptName string,
) string {
	scriptDir := i.getScriptDir(serial, scriptName)
	if scriptDir == "" {
		return ""
	}

	runJSON := filepath.Join(scriptDir, filesdb.RunJSON)
	if ok := file.CreateFileIfNotExist(runJSON); !ok {
		return ""
	}
	return runJSON
}

func (i *interactorImpl) getScriptDir(serial string, name string) string {
	if name == "" || serial == "" {
		return ""
	}
	var device = i.GetDevice(serial)
	var modelOS = device.ToModelOs()
	if modelOS == "" {
		return ""
	}

	return i.filesDB.CreateDBDir(modelOS, filesdb.ScriptsDir, name)
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

func (i *interactorImpl) saveScriptInFile(script *models.Script, filePath string) bool {
	if script == nil || filePath == "" {
		return false
	}
	bytes := script.ToJSON()
	if len(bytes) == 0 {
		return false
	}
	err := os.WriteFile(filePath, bytes, 0644)
	return err == nil
}
