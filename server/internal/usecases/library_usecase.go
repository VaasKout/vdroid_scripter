package usecases

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gocv.io/x/gocv"
)

// LibraryResponse ...
type LibraryResponse struct {
	Images  []string `json:"images"`
	Actions []string `json:"actions"`
}

// LibraryUseCase ...
type LibraryUseCase interface {
	GetLibrary() *LibraryResponse
	SaveImage(serial string, rectangle *models.Rectangle, basePort int) error
	DeleteImage(name string) bool
	SaveAction(action *models.Action) bool
	DeleteAction(name string) bool
}

func (i *interactorImpl) GetLibrary() *LibraryResponse {
	return &LibraryResponse{
		Images:  i.libraryNames(i.filesDB.CreateImagesDir(), file.PngExt),
		Actions: i.libraryNames(i.filesDB.CreateActionsDir(), file.JSONExt),
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

func (i *interactorImpl) SaveImage(
	serial string,
	rectangle *models.Rectangle,
	basePort int,
) error {
	serial = strings.TrimSpace(serial)
	if serial == "" || rectangle.IsEmpty() {
		return errors.New("serial and rectangle are required")
	}

	name := strings.TrimSpace(rectangle.Label)
	if !file.ValidName(name) {
		return errors.New("invalid image name")
	}

	imagesDir := i.filesDB.CreateImagesDir()
	if imagesDir == "" {
		return errors.New("images dir not found")
	}

	if err := i.ensureSessionIsRunning(serial, basePort); err != nil {
		return err
	}

	frame, err := i.latestFrame(serial, true)
	if err != nil {
		return err
	}
	defer frame.Close()

	bounds := image.Rect(0, 0, frame.Cols(), frame.Rows())
	zone := rectangle.ToImageRectangle().Intersect(bounds)
	if zone.Empty() {
		return fmt.Errorf("rectangle lies outside the %dx%d frame", bounds.Dx(), bounds.Dy())
	}

	cropped := frame.Region(zone)
	defer cropped.Close()

	imageName := name + file.PngExt
	imgPath := filepath.Join(imagesDir, imageName)
	if !gocv.IMWrite(imgPath, cropped) {
		return fmt.Errorf("couldn't write %s", imgPath)
	}
	return i.saveImageInfo(serial, imgPath)
}

func (i *interactorImpl) saveImageInfo(serial string, imgPath string) error {
	info := &models.ImageInfo{Density: i.deviceDensity(serial)}
	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(imageInfoPath(imgPath), bytes, 0644)
}

func (i *interactorImpl) imageScale(serial string, imgPath string) float64 {
	bytes, err := os.ReadFile(imageInfoPath(imgPath))
	if err != nil {
		return 1
	}

	info := &models.ImageInfo{}
	err = json.Unmarshal(bytes, info)
	if err != nil {
		return 1
	}
	return info.ScaleFor(i.deviceDensity(serial))
}

func (i *interactorImpl) deviceDensity(serial string) int {
	device, ok := i.devicesCache.Get(serial)
	if !ok {
		return 0
	}
	return device.Density
}

func imageInfoPath(imgPath string) string {
	return strings.TrimSuffix(imgPath, file.PngExt) + file.JSONExt
}

func (i *interactorImpl) DeleteImage(name string) bool {
	if !file.ValidName(name) {
		return false
	}

	imagesDir := i.filesDB.CreateImagesDir()
	if imagesDir == "" {
		return false
	}
	name = strings.TrimSpace(name)
	infoName := name + file.JSONExt
	infoPath := filepath.Join(imagesDir, infoName)
	_ = os.Remove(infoPath)

	imageName := name + file.PngExt
	return i.filesDB.DeleteFileByName(imagesDir, imageName)
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

	actionPath := filepath.Join(actionsDir, action.Name+file.JSONExt)
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

	actionPath := filepath.Join(actionsDir, strings.TrimSpace(name)+file.JSONExt)
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
	return i.filesDB.DeleteFileByName(actionsDir, strings.TrimSpace(name)+file.JSONExt)
}
