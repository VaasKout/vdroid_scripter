// Package cv ...
package cv

import (
	"android_vision_scripter/internal/bashcmd"
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
	cmdRunner bashcmd.CmdAPI
	logAPI    *logger.Logger
}

// New instance of CvAPI
func New(
	cmdRunner bashcmd.CmdAPI,
	logAPI *logger.Logger,
) API {
	return &cvImpl{
		cmdRunner: cmdRunner,
		logAPI:    logAPI,
	}
}
