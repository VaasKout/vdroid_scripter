package filesdb

import (
	"android_vision_scripter/pkg/core/file"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Read ...
type Read interface {
	GetFiles(dir string) []string
	GetDirs(dir string, recurcievly bool) []string
	FindFileByName(dir string, name string) string
	FindDirByName(dir string, name string) string
}

func (f *filesDBImpl) GetFiles(dir string) []string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Dir is not exist: %s\n", dir)
		return []string{}
	}
	files, err := file.GetFilesInDirectory(dir, true)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	return files
}

func (f *filesDBImpl) GetDirs(dir string, recurcievly bool) []string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Dir is not exist: %s\n", dir)
		return []string{}
	}
	dirs, err := file.GetDirsInDirectory(dir, recurcievly)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	return dirs
}

func (f *filesDBImpl) FindFileByName(dir string, name string) string {
	files := f.GetFiles(dir)
	for _, path := range files {
		var fileWithExt = filepath.Base(path)
		if strings.EqualFold(fileWithExt, name) {
			return path
		}
		var fileName = file.GetFileName(path)
		if strings.EqualFold(fileName, name) {
			return path
		}
	}
	return ""
}

func (f *filesDBImpl) FindDirByName(dir string, name string) string {
	dirs := f.GetDirs(dir, true)
	for _, path := range dirs {
		var fileName = file.GetFileName(path)
		if strings.EqualFold(fileName, name) {
			return path
		}
	}
	return ""
}
