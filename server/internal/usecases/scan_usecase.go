package usecases

import (
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gocv.io/x/gocv"
)

type ScanUseCase interface {
	Scan(serial string, images []string, locale string, basePort int) ([]FoundLandmark, error)
}

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

	if _, ok := i.sessionsCache.Get(serial); !ok {
		started := i.StartSession(serial, basePort)
		if !started {
			return nil, fmt.Errorf("couldn't start scrcpy server for %s", serial)
		}
		go i.scrcpy.ReadVideoStream(serial, nil)
	}

	mat, err := i.waitForFrame(serial)
	if err != nil {
		return nil, err
	}
	defer mat.Close()

	landmarks := i.scanYolo(mat)

	textLandmarks, err := i.scanText(serial, mat, locale)
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

func (i *interactorImpl) waitForFrame(serial string) (*gocv.Mat, error) {
	deadline := time.Now().Add(time.Duration(models.DefaultTimeout) * time.Second)
	for time.Now().Before(deadline) {
		mat, err := i.scrcpy.GetMatFromLastFrame(serial, true)
		if err != nil {
			i.logger.Error(err.Error())
			sleepUntilNextSecond()
			continue
		}
		if mat == nil {
			sleepUntilNextSecond()
			continue
		}
		return mat, nil
	}
	return nil, fmt.Errorf("no video frame received from %s", serial)
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
	serial string,
	mat *gocv.Mat,
	locale string,
) ([]FoundLandmark, error) {
	tesseractDir := i.filesDB.CreateLogsDir(serial, filesdb.TesseractDir)
	if tesseractDir == "" {
		return nil, errors.New("tesseract dir was not found")
	}

	ocrParams := cv.InitOcrParams("", locale, cv.PsmText, cv.OemText)
	results, err := i.cv.FindTextRectangles(mat, tesseractDir, ocrParams)
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
