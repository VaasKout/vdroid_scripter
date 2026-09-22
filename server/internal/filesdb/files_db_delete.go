package filesdb

import (
	"fmt"
	"os"
)

// Delete ...
type Delete interface {
	DeleteFileByName(dir string, name string) bool
	DeleteDirByName(dir string, name string) bool
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
