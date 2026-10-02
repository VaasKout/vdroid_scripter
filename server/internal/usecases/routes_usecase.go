package usecases

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RunRouteParams ...
type RunRouteParams struct {
	Serial   string
	Name     string
	StartID  int
	EndID    int
	Args     []string
	BasePort int
}

// RouteUseCase ...
type RouteUseCase interface {
	GetRoutes() []string
	GetRoute(name string) (*models.Route, error)
	SaveRoute(route *models.Route) error
	DeleteRoute(name string) bool
	RunRoute(params *RunRouteParams) error
}

func (i *interactorImpl) GetRoutes() []string {
	return i.libraryNames(i.filesDB.CreateRoutesDir(), file.JSONExt)
}

func (i *interactorImpl) GetRoute(name string) (*models.Route, error) {
	name = strings.TrimSpace(name)
	if !file.ValidName(name) {
		return nil, fmt.Errorf("invalid route name: %s", name)
	}

	routesDir := i.filesDB.CreateRoutesDir()
	if routesDir == "" {
		return nil, errors.New("routes dir not found")
	}

	data, err := os.ReadFile(filepath.Join(routesDir, name+file.JSONExt))
	if err != nil {
		return nil, fmt.Errorf("route not found: %s", name)
	}

	var route = &models.Route{}
	err = json.Unmarshal(data, route)
	if err != nil {
		return nil, err
	}
	return route, nil
}

func (i *interactorImpl) SaveRoute(route *models.Route) error {
	if route == nil {
		return errors.New("route is empty")
	}
	route.Name = strings.TrimSpace(route.Name)
	if !route.Valid() {
		return errors.New("invalid route")
	}
	if err := i.checkStepAssets(route.Steps); err != nil {
		return err
	}
	route.StampStepIDs()

	routesDir := i.filesDB.CreateRoutesDir()
	if routesDir == "" {
		return errors.New("routes dir not found")
	}

	data, err := json.MarshalIndent(route, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(routesDir, route.Name+file.JSONExt), data, 0644)
}

func (i *interactorImpl) DeleteRoute(name string) bool {
	if !file.ValidName(name) {
		return false
	}

	routesDir := i.filesDB.CreateRoutesDir()
	if routesDir == "" {
		return false
	}
	return i.filesDB.DeleteFileByName(routesDir, strings.TrimSpace(name)+file.JSONExt)
}

func (i *interactorImpl) RunRoute(params *RunRouteParams) error {
	if params == nil {
		return errors.New("params are empty")
	}

	route, err := i.GetRoute(params.Name)
	if err != nil {
		return err
	}

	steps, err := route.StepsWithArgs(params.Args)
	if err != nil {
		return err
	}
	route.Steps = steps
	steps, err = route.StepsBetween(params.StartID, params.EndID)
	if err != nil {
		return err
	}
	models.ClearStartDelay(steps)
	return i.RunSteps(params.Serial, steps, params.BasePort)
}
