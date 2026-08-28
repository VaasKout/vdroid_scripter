package models

import (
	"android_vision_scripter/pkg/core/file"
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
