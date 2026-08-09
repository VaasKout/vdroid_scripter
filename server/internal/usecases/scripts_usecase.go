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
	Node      string           `json:"node"`
	Name      string           `json:"name"`
	Value     string           `json:"value"`
	Rectangle models.Rectangle `json:"rectangle"`
}

func (s *SaveZoneDto) Valid() bool {
	return s != nil && strings.TrimSpace(s.Serial) != "" &&
		strings.TrimSpace(s.Node) != "" && strings.TrimSpace(s.Name) != "" &&
		strings.TrimSpace(s.Value) != "" && s.Rectangle.IsNotEmpty()
}

// ScriptsUseCase ...
type ScriptsUseCase interface {
	GetNodes() ([]string, error)
	GetScriptNames(node string) ([]string, error)
	GetScript(node string, scriptName string) (*models.Script, error)
	DeleteScript(node string, scriptName string) error
	RunScript(serial string, node string, scriptName string, basePort int) error

	SaveZone(saveZone *SaveZoneDto) bool
	SaveScript(data *models.Script) bool
	FindText(serial string, text string, locale string) []cv.OCRResult
}

func (i *interactorImpl) SaveZone(saveZone *SaveZoneDto) bool {
	if !saveZone.Valid() {
		return false
	}
	var scriptDir = i.filesDB.CreateScriptDir(saveZone.Node, saveZone.Name)
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
	data.Node = strings.TrimSpace(data.Node)
	if data.Name == "" || data.Node == "" {
		return false
	}

	runnerPath := i.getScriptRunner(data.Node, data.Name)
	if runnerPath == "" {
		return false
	}

	return i.saveScriptInFile(data, runnerPath)
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

func (i *interactorImpl) GetNodes() ([]string, error) {
	return i.getFromScriptDirs(), nil
}

func (i *interactorImpl) GetScriptNames(node string) ([]string, error) {
	node = strings.TrimSpace(node)
	if node == "" {
		nodes := i.getFromScriptDirs()
		var allScripts = []string{}
		for _, node := range nodes {
			names := i.getFromScriptDirs(node)
			allScripts = append(allScripts, names...)
		}
		return allScripts, nil
	}

	return i.getFromScriptDirs(node), nil
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

func (i *interactorImpl) GetScript(node string, scriptName string) (*models.Script, error) {
	scriptName = strings.TrimSpace(scriptName)
	if scriptName == "" {
		return &models.Script{}, errors.New(ScriptNameIsEmpty)
	}

	runnerPath := i.getScriptRunner(node, scriptName)
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

func (i *interactorImpl) DeleteScript(node string, scriptName string) error {
	scriptName = strings.TrimSpace(scriptName)
	nodeDir := i.filesDB.CreateScriptDir(node)
	if scriptName == "" {
		return os.RemoveAll(nodeDir)
	}
	i.filesDB.DeleteDirByName(nodeDir, scriptName)
	return nil
}

func (i *interactorImpl) getScriptRunner(
	node string,
	name string,
) string {
	scriptDir := i.filesDB.CreateScriptDir(node, name)
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
