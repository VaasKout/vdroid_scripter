package filesdb

import (
	"fmt"
	"os"
	"path/filepath"
)

// Delete ...
type Delete interface {
	DeletePathInKeyboardDir(args ...string) bool
	DeleteFileByName(dir string, name string) bool
	DeleteDirByName(dir string, name string) bool
}

// DeletePathInDBDir ...
func (f *filesDBImpl) DeletePathInKeyboardDir(args ...string) bool {
	var filePath = filepath.Join(f.filesProps.Keyboards, filepath.Join(args...))
	err := os.RemoveAll(filePath)
	if err != nil {
		fmt.Printf("Couldn't delete file %s - %s\n", filePath, err.Error())
		return false
	}
	return true
}

// DeleteFileByName ...
func (f *filesDBImpl) DeleteFileByName(dir string, name string) bool {
	path := f.FindFileByName(dir, name)
	if path == "" {
		fmt.Printf("File %s not found in %s\n", name, dir)
		return false
	}
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Couldn't delete file %s - %s\n", path, err.Error())
		return false
	}
	return true
}

// DeleteDirByName ...
func (f *filesDBImpl) DeleteDirByName(dir string, name string) bool {
	path := f.FindDirByName(dir, name)
	if path == "" {
		fmt.Printf("Dir %s not found in %s\n", name, dir)
		return false
	}
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Couldn't delete dir %s - %s\n", path, err.Error())
		return false
	}
	return true
}
