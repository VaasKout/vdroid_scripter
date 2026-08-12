package usecases

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	ImageExt  = ".png"
	ActionExt = ".json"
)

type Library struct {
	Images  []string `json:"images"`
	Actions []string `json:"actions"`
}

type SaveImageDto struct {
	Serial    string           `json:"serial"`
	Name      string           `json:"name"`
	Rectangle models.Rectangle `json:"rectangle"`
}

func (s *SaveImageDto) Valid() bool {
	return s != nil && strings.TrimSpace(s.Serial) != "" &&
		ValidLibraryName(s.Name) && s.Rectangle.IsNotEmpty()
}

func ValidLibraryName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return false
	}
	return !strings.ContainsAny(name, `/\`)
}

type LibraryUseCase interface {
	GetLibrary() *Library
	SaveImage(data *SaveImageDto) bool
	DeleteImage(name string) bool
	SaveAction(action *models.Action) bool
	DeleteAction(name string) bool
}

func (i *interactorImpl) GetLibrary() *Library {
	return &Library{
		Images:  i.libraryNames(i.filesDB.CreateImagesDir(), ImageExt),
		Actions: i.libraryNames(i.filesDB.CreateActionsDir(), ActionExt),
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

func (i *interactorImpl) SaveImage(data *SaveImageDto) bool {
	if !data.Valid() {
		return false
	}

	imagesDir := i.filesDB.CreateImagesDir()
	if imagesDir == "" {
		return false
	}

	imgPath := filepath.Join(imagesDir, strings.TrimSpace(data.Name)+ImageExt)
	created := file.CreateFileIfNotExist(imgPath)
	if !created {
		return false
	}

	screenShot := i.cmd.ScreenShot(data.Serial)
	if screenShot == "" {
		return false
	}

	imgRect := data.Rectangle.ToImageRectangle()
	i.cv.CutZone(screenShot, imgPath, imgRect)
	return true
}

func (i *interactorImpl) DeleteImage(name string) bool {
	if !ValidLibraryName(name) {
		return false
	}

	imagesDir := i.filesDB.CreateImagesDir()
	if imagesDir == "" {
		return false
	}
	return i.filesDB.DeleteFileByName(imagesDir, strings.TrimSpace(name)+ImageExt)
}

func (i *interactorImpl) SaveAction(action *models.Action) bool {
	if action.IsEmpty() || !ValidLibraryName(action.Name) {
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

	actionPath := filepath.Join(actionsDir, action.Name+ActionExt)
	err := os.WriteFile(actionPath, bytes, 0644)
	if err != nil {
		i.logger.Error(err.Error())
	}
	return err == nil
}

func (i *interactorImpl) DeleteAction(name string) bool {
	if !ValidLibraryName(name) {
		return false
	}

	actionsDir := i.filesDB.CreateActionsDir()
	if actionsDir == "" {
		return false
	}
	return i.filesDB.DeleteFileByName(actionsDir, strings.TrimSpace(name)+ActionExt)
}
