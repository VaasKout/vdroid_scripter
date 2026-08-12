package filesdb

import (
	"android_vision_scripter/pkg/core/file"
	"fmt"
	"path/filepath"
)

// Create ...
type Create interface {
	CreateLogsDir(args ...string) string
	CreateKeyboardDir(args ...string) string
	CreateScrcpyDir(args ...string) string
	CreateOnnxDir(args ...string) string
	CreateImagesDir(args ...string) string
	CreateActionsDir(args ...string) string
}

// CreateLogsDir ...
func (f *filesDBImpl) CreateLogsDir(args ...string) string {
	var dirName = filepath.Join(f.filesProps.Logs, filepath.Join(args...))
	if ok := file.CreateDirIfNotExist(dirName); !ok {
		fmt.Printf("Couldn't create dir %s\n", dirName)
		return ""
	}
	return dirName
}

// CreateLogsDir ...
func (f *filesDBImpl) CreateScrcpyDir(args ...string) string {
	var dirName = filepath.Join(f.filesProps.Scrcpy, filepath.Join(args...))
	if ok := file.CreateDirIfNotExist(dirName); !ok {
		fmt.Printf("couldn't create dir %s\n", dirName)
		return ""
	}
	return dirName
}

// CreateOnnxDir ...
func (f *filesDBImpl) CreateOnnxDir(args ...string) string {
	var dirName = filepath.Join(f.filesProps.Yolo, filepath.Join(args...))
	if ok := file.CreateDirIfNotExist(dirName); !ok {
		fmt.Printf("couldn't create dir %s\n", dirName)
		return ""
	}
	return dirName
}

// CreateKeyboardDir ...
func (f *filesDBImpl) CreateKeyboardDir(args ...string) string {
	var dirName = filepath.Join(f.filesProps.Keyboards, filepath.Join(args...))
	if ok := file.CreateDirIfNotExist(dirName); !ok {
		fmt.Printf("couldn't create dir %s\n", dirName)
		return ""
	}
	return dirName
}

func (f *filesDBImpl) CreateImagesDir(args ...string) string {
	var dirName = filepath.Join(f.filesProps.Images, filepath.Join(args...))
	if ok := file.CreateDirIfNotExist(dirName); !ok {
		fmt.Printf("couldn't create dir %s\n", dirName)
		return ""
	}
	return dirName
}

func (f *filesDBImpl) CreateActionsDir(args ...string) string {
	var dirName = filepath.Join(f.filesProps.Actions, filepath.Join(args...))
	if ok := file.CreateDirIfNotExist(dirName); !ok {
		fmt.Printf("couldn't create dir %s\n", dirName)
		return ""
	}
	return dirName
}
