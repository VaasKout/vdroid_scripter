// Package filesdb ...
package filesdb

import (
	"android_vision_scripter/config"
)

// Basic directories for phone data
const (
	ScreenshotDir = "screenshot"
	TesseractDir  = "tesseract"
)

// FilesDB ...
type FilesDB interface {
	Create
	Read
	Delete
}

type filesDBImpl struct {
	filesProps *config.FilesProps
}

// New instance of FilesDB
func New(
	filesProps *config.FilesProps,
) FilesDB {
	return &filesDBImpl{
		filesProps: filesProps,
	}
}
