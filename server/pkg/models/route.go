package models

import (
	"android_vision_scripter/pkg/core/file"
	"errors"
	"fmt"
)

// RouteArgPlaceholder ...
const RouteArgPlaceholder = "%s"

// ErrRouteArgs ...
var ErrRouteArgs = errors.New("wrong route args")

// Route ...
type Route struct {
	Name  string `json:"name"`
	Steps []Step `json:"steps"`
}

// Valid ...
func (r *Route) Valid() bool {
	if r == nil || !file.ValidName(r.Name) {
		return false
	}
	return ValidQueue(r.Steps)
}

// StampStepIDs ...
func (r *Route) StampStepIDs() {
	if r == nil {
		return
	}
	for index := range r.Steps {
		r.Steps[index].ID = index + 1
	}
}

// StepsBetween ...
func (r *Route) StepsBetween(startID int, endID int) ([]Step, error) {
	if r == nil {
		return nil, errors.New("route is empty")
	}
	if startID > endID {
		endID = 0
	}

	start := 0
	if startID > 0 {
		index, err := r.stepIndex(startID)
		if err != nil {
			return nil, fmt.Errorf("start_id %w", err)
		}
		start = index
	}

	end := len(r.Steps)
	if endID > 0 {
		index, err := r.stepIndex(endID)
		if err != nil {
			return nil, fmt.Errorf("end_id %w", err)
		}
		end = index + 1
	}
	return r.Steps[start:end], nil
}

func (r *Route) stepIndex(id int) (int, error) {
	for index, step := range r.Steps {
		if step.ID == id {
			return index, nil
		}
	}
	return 0, fmt.Errorf("%d not found: route %s has %d steps", id, r.Name, len(r.Steps))
}

// ClearArgLocales ...
func (r *Route) ClearArgLocales() {
	if r == nil {
		return
	}
	for stepIndex := range r.Steps {
		landmarks := r.Steps[stepIndex].Landmarks
		for landmarkIndex := range landmarks {
			if landmarks[landmarkIndex].ArgsCount() == 0 {
				continue
			}
			landmarks[landmarkIndex].Locale = ""
		}
	}
}

// ArgsCount ...
func (r *Route) ArgsCount() int {
	if r == nil {
		return 0
	}

	count := 0
	for _, step := range r.Steps {
		count += step.ArgsCount()
	}
	return count
}

// StepsWithArgs ...
func (r *Route) StepsWithArgs(args []string, argsLocale string) ([]Step, error) {
	if r == nil {
		return nil, errors.New("route is empty")
	}

	needed := r.ArgsCount()
	if len(args) != needed {
		return nil, fmt.Errorf(
			"%w: route %s needs %d, got %d",
			ErrRouteArgs,
			r.Name,
			needed,
			len(args),
		)
	}

	steps := make([]Step, 0, len(r.Steps))
	used := 0
	for _, step := range r.Steps {
		count := step.ArgsCount()
		stepArgs := args[used : used+count]
		filled := step.WithArgs(stepArgs, argsLocale)
		steps = append(steps, filled)
		used += count
	}
	return steps, nil
}
