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

type LibraryResponse struct {
	Images  []string `json:"images"`
	Actions []string `json:"actions"`
}

type LibraryUseCase interface {
	GetLibrary() *LibraryResponse
	SaveImage(serial string, rectangle *models.Rectangle) bool
	DeleteImage(name string) bool
	SaveAction(action *models.Action) bool
	DeleteAction(name string) bool
}

func (i *interactorImpl) GetLibrary() *LibraryResponse {
	return &LibraryResponse{
		Images:  i.libraryNames(i.filesDB.CreateImagesDir(), file.PngExt),
		Actions: i.libraryNames(i.filesDB.CreateActionsDir(), file.JsonExt),
	}
}

func (i *interactorImpl) libraryNames(dir string, ext string) []string {
	names := []string{}
	if dir == "" {
		return names
	}
	for _, path := range i.filesDB.GetFiles(dir) {
		if !strings.EqualFold(filepath.Ext(path), ext) {
			continue
		}
		names = append(names, file.GetFileName(path))
	}
	sort.Strings(names)
	return names
}

func (i *interactorImpl) SaveImage(serial string, rectangle *models.Rectangle) bool {
	serial = strings.TrimSpace(serial)
	if serial == "" || rectangle.IsEmpty() {
		return false
	}

	name := strings.TrimSpace(rectangle.Label)
	if !file.ValidName(name) {
		return false
	}

	imagesDir := i.filesDB.CreateImagesDir()
	if imagesDir == "" {
		return false
	}

	imgPath := filepath.Join(imagesDir, name+file.PngExt)
	created := file.CreateFileIfNotExist(imgPath)
	if !created {
		return false
	}

	screenShot := i.cmd.ScreenShot(serial)
	if screenShot == "" {
		return false
	}

	imgRect := rectangle.ToImageRectangle()
	i.cv.CutZone(screenShot, imgPath, imgRect)
	return true
}

func (i *interactorImpl) DeleteImage(name string) bool {
	if !file.ValidName(name) {
		return false
	}

	imagesDir := i.filesDB.CreateImagesDir()
	if imagesDir == "" {
		return false
	}
	return i.filesDB.DeleteFileByName(imagesDir, strings.TrimSpace(name)+file.PngExt)
}

func (i *interactorImpl) SaveAction(action *models.Action) bool {
	if action.IsEmpty() || !file.ValidName(action.Name) {
		return false
	}

	actionsDir := i.filesDB.CreateActionsDir()
	if actionsDir == "" {
		return false
	}

	action.Name = strings.TrimSpace(action.Name)
	bytes := action.ToJSON()
	if len(bytes) == 0 {
		return false
	}

	actionPath := filepath.Join(actionsDir, action.Name+file.JsonExt)
	err := os.WriteFile(actionPath, bytes, 0644)
	if err != nil {
		i.logger.Error(err.Error())
	}
	return err == nil
}

func (i *interactorImpl) getAction(name string) (*models.Action, error) {
	actionsDir := i.filesDB.CreateActionsDir()
	if actionsDir == "" {
		return nil, errors.New("actions dir not found")
	}

	actionPath := filepath.Join(actionsDir, strings.TrimSpace(name)+file.JsonExt)
	bytes, err := os.ReadFile(actionPath)
	if err != nil {
		return nil, fmt.Errorf("action not found in library: %s", name)
	}

	var action = &models.Action{}
	err = json.Unmarshal(bytes, action)
	if err != nil {
		return nil, err
	}
	if action.IsEmpty() {
		return nil, fmt.Errorf("action is empty: %s", name)
	}
	return action, nil
}

func (i *interactorImpl) DeleteAction(name string) bool {
	if !file.ValidName(name) {
		return false
	}

	actionsDir := i.filesDB.CreateActionsDir()
	if actionsDir == "" {
		return false
	}
	return i.filesDB.DeleteFileByName(actionsDir, strings.TrimSpace(name)+file.JsonExt)
}
