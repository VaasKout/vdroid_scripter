package models

import (
	"android_vision_scripter/pkg/core/file"
	"errors"
	"fmt"
)

// Route ...
type Route struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt,omitempty"`
	Steps  []Step `json:"steps"`
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
