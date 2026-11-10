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
	"slices"
	"sort"
	"strings"

	"gocv.io/x/gocv"
)

const modelOsSeparator = "_"

// LibraryResponse ...
type LibraryResponse struct {
	Images  []string `json:"images"`
	Actions []string `json:"actions"`
}

// LibraryUseCase ...
type LibraryUseCase interface {
	GetLibrary() *LibraryResponse
	SaveImage(serial string, rectangle *models.Rectangle, basePort int) error
	GetDeviceSpecificName(serial string, name string) string
	DeleteImage(name string) bool
	RenameImage(from string, to string) (string, error)
	SaveAction(action *models.Action) bool
	DeleteAction(name string) bool
	RenameAction(from string, to string) (string, error)
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

func (i *interactorImpl) getAction(serial string, name string) (*models.Action, error) {
	device, _ := i.devicesCache.Get(serial)
	actionPath, ok := libraryPath(i.filesDB.CreateActionsDir(), name, file.JSONExt, device.ModelOs)
	if !ok {
		return nil, libraryNotFound("action", name, device.ModelOs)
	}

	bytes, err := os.ReadFile(actionPath)
	if err != nil {
		return nil, err
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

func (i *interactorImpl) RenameImage(from string, to string) (string, error) {
	return renameLibraryItem(i.filesDB.CreateImagesDir(), from, to, i.deviceTags(), file.PngExt)
}

func (i *interactorImpl) RenameAction(from string, to string) (string, error) {
	return renameLibraryItem(i.filesDB.CreateActionsDir(), from, to, i.deviceTags(), file.JSONExt)
}

func (i *interactorImpl) deviceTags() []string {
	var tags []string
	for _, device := range i.devicesCache.GetDataArray() {
		if device.ModelOs != "" && !slices.Contains(tags, device.ModelOs) {
			tags = append(tags, device.ModelOs)
		}
	}
	return tags
}

func renameLibraryItem(
	dir string,
	from string,
	to string,
	tags []string,
	mainExt string,
) (string, error) {
	from, to = strings.TrimSpace(from), strings.TrimSpace(to)
	if dir == "" || !file.ValidName(from) || !file.ValidName(to) {
		return "", errors.New("invalid name")
	}
	path, ok := libraryPath(dir, from, mainExt, tags...)
	if !ok {
		return "", fmt.Errorf("not found in library: %s (tried %s)", from, strings.Join(libraryCandidates(from, tags...), ", "))
	}
	source := file.GetFileName(path)
	target := variantName(to, source, tags)
	if target == source {
		return source, nil
	}
	if file.Exists(filepath.Join(dir, target+mainExt)) {
		return "", fmt.Errorf("exists in library: %s", target)
	}
	if err := os.Rename(filepath.Join(dir, source+mainExt), filepath.Join(dir, target+mainExt)); err != nil {
		return "", err
	}
	infoPath := filepath.Join(dir, source+file.JSONExt)
	if !file.Exists(infoPath) {
		return target, nil
	}
	if err := os.Rename(infoPath, filepath.Join(dir, target+file.JSONExt)); err != nil {
		return target, fmt.Errorf("renamed to %s, but its info file was not: %w", target, err)
	}
	return target, nil
}

func variantName(to string, source string, tags []string) string {
	for _, tag := range tags {
		suffix := modelOsSeparator + tag
		if strings.HasSuffix(source, suffix) && !strings.HasSuffix(to, suffix) {
			return to + suffix
		}
	}
	return to
}

func (i *interactorImpl) GetDeviceSpecificName(serial string, name string) string {
	name = strings.TrimSpace(name)
	device, _ := i.devicesCache.Get(serial)
	if name == "" || device.ModelOs == "" {
		return ""
	}
	suffix := modelOsSeparator + device.ModelOs
	if strings.HasSuffix(name, suffix) {
		return name
	}
	return name + suffix
}

func libraryPath(dir string, name string, ext string, tags ...string) (string, bool) {
	if dir == "" {
		return "", false
	}
	for _, candidate := range libraryCandidates(name, tags...) {
		path := filepath.Join(dir, candidate+ext)
		if file.Exists(path) {
			return path, true
		}
	}
	return "", false
}

func libraryCandidates(name string, tags ...string) []string {
	name = strings.TrimSpace(name)
	candidates := make([]string, 0, len(tags)+1)
	for _, tag := range slices.Compact(slices.Sorted(slices.Values(tags))) {
		if tag != "" {
			candidates = append(candidates, name+modelOsSeparator+tag)
		}
	}
	return append(candidates, name)
}

func libraryNotFound(kind string, name string, tags ...string) error {
	candidates := libraryCandidates(name, tags...)
	if len(candidates) == 1 {
		return fmt.Errorf("%s not found in library: %s", kind, name)
	}
	return fmt.Errorf(
		"%s not found in library: %s (tried %s)",
		kind, name, strings.Join(candidates, ", "),
	)
}
