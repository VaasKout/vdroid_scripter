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
	"strings"

	"gocv.io/x/gocv"
)

// Common script errors
const (
	ScriptNameIsEmpty = "script name is empty"
)

type SaveZoneDto struct {
	Serial    string           `json:"serial"`
	Location  string           `json:"location"`
	Name      string           `json:"name"`
	Value     string           `json:"value"`
	Rectangle models.Rectangle `json:"rectangle"`
}

func (s *SaveZoneDto) Valid() bool {
	return s != nil && strings.TrimSpace(s.Serial) != "" &&
		strings.TrimSpace(s.Location) != "" && strings.TrimSpace(s.Name) != "" &&
		strings.TrimSpace(s.Value) != "" && s.Rectangle.IsNotEmpty()
}

type RunScriptsDto struct {
	Serial  string      `json:"serial"`
	Scripts []ScriptRef `json:"scripts"`
}

type ScriptRef struct {
	Location string `json:"location"`
	Name     string `json:"name"`
}

func (r *RunScriptsDto) Valid() bool {
	return r != nil && strings.TrimSpace(r.Serial) != "" && len(r.Scripts) > 0
}

type EditScriptDto struct {
	PrevLocation string         `json:"prev_location"`
	Script       *models.Script `json:"script"`
}

func (e *EditScriptDto) Valid() bool {
	return e != nil && e.Script != nil &&
		strings.TrimSpace(e.Script.Name) != "" &&
		strings.TrimSpace(e.Script.Location) != ""
}

// ScriptsUseCase ...
type ScriptsUseCase interface {
	GetLocations() ([]string, error)
	GetScriptNames(location string) ([]string, error)
	GetScript(location string, scriptName string) (*models.Script, error)
	DeleteScript(location string, scriptName string) error
	RunScripts(serial string, scripts []ScriptRef, basePort int) error

	SaveZone(saveZone *SaveZoneDto) bool
	SaveScript(data *models.Script) bool
	EditScript(data *EditScriptDto) bool
	FindText(serial string, text string, locale string) []cv.OCRResult
}

func (i *interactorImpl) SaveZone(saveZone *SaveZoneDto) bool {
	if !saveZone.Valid() {
		return false
	}
	var scriptDir = i.filesDB.CreateScriptDir(saveZone.Location, saveZone.Name)
	var tmpImg = filepath.Join(scriptDir, saveZone.Value+".png")
	created := file.CreateFileIfNotExist(tmpImg)
	if !created {
		return false
	}

	screenShot := i.cmd.ScreenShot(saveZone.Serial)
	if screenShot == "" {
		return false
	}
	imgRect := saveZone.Rectangle.ToImageRectangle()
	i.cv.CutZone(screenShot, tmpImg, imgRect)
	return true
}

func (i *interactorImpl) SaveScript(data *models.Script) bool {
	if data == nil {
		return false
	}

	data.Name = strings.TrimSpace(data.Name)
	data.Location = strings.TrimSpace(data.Location)
	if data.Name == "" || data.Location == "" {
		return false
	}

	runnerPath := i.getScriptRunner(data.Location, data.Name)
	if runnerPath == "" {
		return false
	}

	return i.saveScriptInFile(data, runnerPath)
}

func (i *interactorImpl) EditScript(data *EditScriptDto) bool {
	if !data.Valid() {
		return false
	}

	script := data.Script
	name := strings.TrimSpace(script.Name)
	newLocation := strings.TrimSpace(script.Location)
	prevLocation := strings.TrimSpace(data.PrevLocation)

	if prevLocation != "" && prevLocation != newLocation {
		moved := i.moveScriptFiles(prevLocation, newLocation, name)
		if !moved {
			return false
		}
	}

	return i.SaveScript(script)
}

func (i *interactorImpl) moveScriptFiles(
	prevLocation string,
	location string,
	name string,
) bool {
	prevScriptDir := filepath.Join(i.filesDB.CreateScriptDir(), prevLocation, name)
	if _, err := os.Stat(prevScriptDir); err != nil {
		return true
	}

	newLocationDir := i.filesDB.CreateScriptDir(location)
	if newLocationDir == "" {
		return false
	}

	newScriptDir := filepath.Join(newLocationDir, name)
	err := os.Rename(prevScriptDir, newScriptDir)
	if err != nil {
		i.logger.Error(err.Error())
	}
	return err == nil
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

	dir := i.filesDB.CreateLogsDir(serial, filesdb.TesseractDir)
	if dir == "" {
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

	result, err := i.cv.FindTextRectangles(&img, dir, ocrParams)
	if err != nil {
		i.logger.Error(err.Error())
		return []cv.OCRResult{}
	}
	return result
}

func (i *interactorImpl) GetLocations() ([]string, error) {
	return i.getFromScriptDirs(), nil
}

func (i *interactorImpl) GetScriptNames(location string) ([]string, error) {
	location = strings.TrimSpace(location)
	if location == "" {
		locations := i.getFromScriptDirs()
		var allScripts = []string{}
		for _, location := range locations {
			names := i.getFromScriptDirs(location)
			allScripts = append(allScripts, names...)
		}
		return allScripts, nil
	}

	return i.getFromScriptDirs(location), nil
}

func (i *interactorImpl) getFromScriptDirs(args ...string) []string {
	scriptsDir := i.filesDB.CreateScriptDir(args...)
	dirs := i.filesDB.GetDirs(scriptsDir, false)
	names := make([]string, 0, len(dirs))
	for _, script := range dirs {
		name := filepath.Base(script)
		names = append(names, name)
	}
	return names
}

func (i *interactorImpl) GetScript(location string, scriptName string) (*models.Script, error) {
	scriptName = strings.TrimSpace(scriptName)
	if scriptName == "" {
		return &models.Script{}, errors.New(ScriptNameIsEmpty)
	}

	runnerPath := i.getScriptRunner(location, scriptName)
	if runnerPath == "" {
		var errStr = fmt.Sprintf("unable to create json runner for %s", scriptName)
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

func (i *interactorImpl) DeleteScript(location string, scriptName string) error {
	scriptName = strings.TrimSpace(scriptName)
	locationDir := i.filesDB.CreateScriptDir(location)
	if scriptName == "" {
		return os.RemoveAll(locationDir)
	}
	i.filesDB.DeleteDirByName(locationDir, scriptName)
	return nil
}

func (i *interactorImpl) getScriptRunner(
	location string,
	name string,
) string {
	scriptDir := i.filesDB.CreateScriptDir(location, name)
	if scriptDir == "" {
		return ""
	}

	runJSON := filepath.Join(scriptDir, filesdb.RunJSON)
	if ok := file.CreateFileIfNotExist(runJSON); !ok {
		return ""
	}
	return runJSON
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
