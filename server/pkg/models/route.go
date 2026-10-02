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

// StepsFromID ...
func (r *Route) StepsFromID(startID int) ([]Step, error) {
	if r == nil {
		return nil, errors.New("route is empty")
	}
	for index, step := range r.Steps {
		if step.ID == startID {
			return r.Steps[index:], nil
		}
	}
	return nil, fmt.Errorf(
		"start_id %d not found: route %s has %d steps",
		startID,
		r.Name,
		len(r.Steps),
	)
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
func (r *Route) StepsWithArgs(args []string) ([]Step, error) {
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
		filled := step.WithArgs(stepArgs)
		steps = append(steps, filled)
		used += count
	}
	return steps, nil
}
