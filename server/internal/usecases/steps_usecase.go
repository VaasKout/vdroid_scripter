package usecases

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const frameWaitTimeout = 5 * time.Second

type RunStepsDto struct {
	Serial string        `json:"serial"`
	Steps  []models.Step `json:"steps"`
}

func (r *RunStepsDto) Valid() bool {
	if r == nil || strings.TrimSpace(r.Serial) == "" || len(r.Steps) == 0 {
		return false
	}
	for index := range r.Steps {
		if !r.Steps[index].Valid() {
			return false
		}
	}
	return true
}

type StepsUseCase interface {
	RunSteps(serial string, steps []models.Step, basePort int) error
}

func (i *interactorImpl) RunSteps(
	serial string,
	steps []models.Step,
	basePort int,
) error {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return errors.New(SerialIsEmptyError)
	}
	if len(steps) == 0 {
		return errors.New("steps are empty")
	}

	for index := range steps {
		step := &steps[index]
		if !step.Valid() {
			return fmt.Errorf("invalid step: %s", step.ToString())
		}
		if err := i.checkStepAssets(step); err != nil {
			return err
		}
	}

	if _, ok := i.sessionsCache.Get(serial); !ok {
		started := i.StartSession(serial, basePort)
		if !started {
			return fmt.Errorf("couldn't start scrcpy server for %s", serial)
		}
		go i.scrcpy.ReadVideoStream(serial, nil)
	}

	i.addStepsToQueue(serial, steps)
	return nil
}

func (i *interactorImpl) checkStepAssets(step *models.Step) error {
	if step.IsCustomEvent() {
		if !ValidLibraryName(step.Event) {
			return fmt.Errorf("invalid event name: %s", step.Event)
		}
		actionPath := filepath.Join(
			i.filesDB.CreateActionsDir(),
			strings.TrimSpace(step.Event)+ActionExt,
		)
		if !file.Exists(actionPath) {
			return fmt.Errorf("event not found in library: %s", step.Event)
		}
	}

	if step.Type != models.Image || !step.HasType() {
		return nil
	}

	if !ValidLibraryName(step.Value) {
		return fmt.Errorf("invalid image name: %s", step.Value)
	}
	imagePath := filepath.Join(
		i.filesDB.CreateImagesDir(),
		strings.TrimSpace(step.Value)+ImageExt,
	)
	if !file.Exists(imagePath) {
		return fmt.Errorf("image not found in library: %s", step.Value)
	}
	return nil
}

func (i *interactorImpl) waitForFrameSizes(serial string) (int, int, error) {
	deadline := time.Now().Add(frameWaitTimeout)
	for time.Now().Before(deadline) {
		mat, err := i.scrcpy.GetMatFromLastFrame(serial, false)
		if err != nil || mat == nil {
			time.Sleep(250 * time.Millisecond)
			continue
		}

		width := mat.Cols()
		height := mat.Rows()
		mat.Close()
		if width > 0 && height > 0 {
			return width, height, nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return 0, 0, fmt.Errorf("no video frame received for %s", serial)
}
