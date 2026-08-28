package usecases

import (
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"gocv.io/x/gocv"
)

// ScanUseCase ...
type ScanUseCase interface {
	Scan(
		serial string,
		images []string,
		locale string,
		basePort int,
	) ([]FoundLandmark, error)
}

// FoundLandmark ...
type FoundLandmark struct {
	Type      string           `json:"type"`
	Value     string           `json:"value"`
	Locale    string           `json:"locale,omitempty"`
	Rectangle models.Rectangle `json:"rectangle"`
}

func (i *interactorImpl) Scan(
	serial string,
	images []string,
	locale string,
	basePort int,
) ([]FoundLandmark, error) {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return nil, errors.New(SerialIsEmptyError)
	}

	imagePaths, err := i.libraryImagePaths(images)
	if err != nil {
		return nil, err
	}

	if err := i.ensureHeadlessSession(serial, basePort); err != nil {
		return nil, err
	}

	mat, err := i.scrcpy.GetMatFromLastFrame(serial, true)
	if err != nil {
		return nil, err
	}
	if mat == nil {
		return nil, fmt.Errorf("no video frame received from %s", serial)
	}
	defer mat.Close()

	landmarks := i.scanYolo(mat)

	textLandmarks, err := i.scanText(mat, locale)
	if err != nil {
		return nil, err
	}
	landmarks = append(landmarks, textLandmarks...)

	imageLandmarks, err := i.scanImages(mat, imagePaths)
	if err != nil {
		return nil, err
	}
	landmarks = append(landmarks, imageLandmarks...)

	sortLandmarksReadingOrder(landmarks)
	return landmarks, nil
}

func (i *interactorImpl) libraryImagePaths(images []string) (map[string]string, error) {
	paths := map[string]string{}
	if len(images) == 0 {
		return paths, nil
	}

	imagesDir := i.filesDB.CreateImagesDir()
	if imagesDir == "" {
		return nil, errors.New("images dir not found")
	}

	for _, name := range images {
		name = strings.TrimSpace(name)
		if !file.ValidName(name) {
			return nil, fmt.Errorf("invalid image name: %s", name)
		}

		imagePath := filepath.Join(imagesDir, name+file.PngExt)
		if !file.Exists(imagePath) {
			return nil, fmt.Errorf("image not found in library: %s", name)
		}
		paths[name] = imagePath
	}
	return paths, nil
}

func (i *interactorImpl) scanYolo(mat *gocv.Mat) []FoundLandmark {
	landmarks := []FoundLandmark{}
	for _, detection := range i.yolo.DetectLabels(*mat) {
		rect := detection
		rect.Label = ""
		landmarks = append(landmarks, FoundLandmark{
			Type:      models.Yolo,
			Value:     detection.Label,
			Rectangle: rect,
		})
	}
	return landmarks
}

func (i *interactorImpl) scanText(
	mat *gocv.Mat,
	locale string,
) ([]FoundLandmark, error) {
	ocrParams := cv.InitOcrParams("", locale, cv.PsmText, cv.OemText)
	results, err := i.cv.FindTextRectangles(mat, ocrParams)
	if err != nil {
		return nil, err
	}

	landmarks := []FoundLandmark{}
	for _, result := range results {
		if result.IsEmpty() {
			continue
		}
		landmarks = append(landmarks, FoundLandmark{
			Type:      models.Text,
			Value:     result.Text,
			Locale:    ocrParams.Lang,
			Rectangle: result.Rectangle,
		})
	}
	return landmarks, nil
}

func (i *interactorImpl) scanImages(
	mat *gocv.Mat,
	imagePaths map[string]string,
) ([]FoundLandmark, error) {
	names := make([]string, 0, len(imagePaths))
	for name := range imagePaths {
		names = append(names, name)
	}
	sort.Strings(names)

	landmarks := []FoundLandmark{}
	for _, name := range names {
		rects, err := i.cv.FindImages(mat, imagePaths[name])
		if err != nil {
			return nil, err
		}
		for _, rect := range rects {
			landmarks = append(landmarks, FoundLandmark{
				Type:      models.Image,
				Value:     name,
				Rectangle: *models.ImgRectangleToDomain(&rect),
			})
		}
	}
	return landmarks, nil
}

func sortLandmarksReadingOrder(landmarks []FoundLandmark) {
	sort.SliceStable(landmarks, func(a, b int) bool {
		if landmarks[a].Rectangle.TopY != landmarks[b].Rectangle.TopY {
			return landmarks[a].Rectangle.TopY < landmarks[b].Rectangle.TopY
		}
		return landmarks[a].Rectangle.LeftX < landmarks[b].Rectangle.LeftX
	})
}
