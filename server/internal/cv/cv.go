// Package cv ...
package cv

import (
	"android_vision_scripter/pkg/logger"
)

// CV rectangles contants
const (
	MinBorderDistance  = 20
	MatchCoefficient   = 0.9
	MaxTemplateMatches = 50
)

// API ...
type API interface {
	ImageHandler
	KeyboardHandler
	TextHandler
}

type cvImpl struct {
	logAPI *logger.Logger
}

// New instance of CvAPI
func New(logAPI *logger.Logger) API {
	return &cvImpl{
		logAPI: logAPI,
	}
}
