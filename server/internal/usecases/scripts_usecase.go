package usecases

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const scriptRunFile = "run.json"

type ScriptsUseCase interface {
	GetScripts() []string
	GetScript(name string) ([]models.Step, error)
	SaveScript(name string, steps []models.Step) error
	DeleteScript(name string) bool
	RunScript(serial string, name string, basePort int) error
}

func (i *interactorImpl) GetScripts() []string {
	names := []string{}
	scriptsDir := i.filesDB.CreateScriptsDir()
	if scriptsDir == "" {
		return names
	}

	for _, dir := range i.filesDB.GetDirs(scriptsDir, false) {
		names = append(names, filepath.Base(dir))
	}
	sort.Strings(names)
	return names
}

func (i *interactorImpl) GetScript(name string) ([]models.Step, error) {
	name = strings.TrimSpace(name)
	if !file.ValidName(name) {
		return nil, fmt.Errorf("invalid script name: %s", name)
	}

	scriptsDir := i.filesDB.CreateScriptsDir()
	if scriptsDir == "" {
		return nil, errors.New("scripts dir not found")
	}

	bytes, err := os.ReadFile(filepath.Join(scriptsDir, name, scriptRunFile))
	if err != nil {
		return nil, fmt.Errorf("script not found: %s", name)
	}

	steps := []models.Step{}
	err = json.Unmarshal(bytes, &steps)
	if err != nil {
		return nil, err
	}
	return steps, nil
}

func (i *interactorImpl) SaveScript(name string, steps []models.Step) error {
	name = strings.TrimSpace(name)
	if !file.ValidName(name) {
		return fmt.Errorf("invalid script name: %s", name)
	}
	if !models.ValidQueue(steps) {
		return errors.New("invalid steps")
	}
	if err := i.checkStepAssets(steps); err != nil {
		return err
	}

	scriptDir := i.filesDB.CreateScriptsDir(name)
	if scriptDir == "" {
		return errors.New("couldn't create script dir")
	}

	bytes, err := json.MarshalIndent(steps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(scriptDir, scriptRunFile), bytes, 0644)
}

func (i *interactorImpl) DeleteScript(name string) bool {
	name = strings.TrimSpace(name)
	if !file.ValidName(name) {
		return false
	}

	scriptsDir := i.filesDB.CreateScriptsDir()
	if scriptsDir == "" {
		return false
	}
	return i.filesDB.DeleteDirByName(scriptsDir, name)
}

func (i *interactorImpl) RunScript(serial string, name string, basePort int) error {
	steps, err := i.GetScript(name)
	if err != nil {
		return err
	}
	return i.RunSteps(serial, steps, basePort)
}
