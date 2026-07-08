package filesdb

import (
	"android_vision_scripter/pkg/core/file"
	"fmt"
	"path/filepath"
)

// Create ...
type Create interface {
	CreateLogsDir(args ...string) string
	CreateScriptDir(args ...string) string
	CreateKeyboardDir(args ...string) string
	CreateScrcpyDir(args ...string) string
	CreateOnnxDir(args ...string) string
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

// CreateScriptDir ...
func (f *filesDBImpl) CreateScriptDir(args ...string) string {
	var dirName = filepath.Join(f.filesProps.Scripts, filepath.Join(args...))
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
