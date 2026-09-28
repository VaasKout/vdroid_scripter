// Package file ...
package file

import (
	"os"
	"path/filepath"
	"strings"
)

// CreateDirIfNotExist ...
func CreateDirIfNotExist(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		return err == nil
	}
	return true
}

// GetFilesInDirectory ...
func GetFilesInDirectory(dir string, recursive bool) ([]string, error) {
	if recursive {
		return getAllInDirectoryRecursively(dir, false, true)
	}
	return getAllInDirectory(dir, false, true)
}

func getAllInDirectory(path string, includeDirs bool, includeFiles bool) ([]string, error) {
	if path == "" || !IsDirectory(path) {
		return []string{}, nil
	}
	var files []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return files, err
	}
	for _, entry := range entries {
		if entry.IsDir() && includeDirs {
			files = append(files, filepath.Join(path, entry.Name()))
		}
		if !entry.IsDir() && includeFiles {
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}
	return files, nil
}

func getAllInDirectoryRecursively(dir string, includeDirs bool, includeFiles bool) ([]string, error) {
	if dir == "" || !IsDirectory(dir) {
		return []string{}, nil
	}

	var files []string
	err := filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dir == path {
			return nil
		}
		if info.IsDir() && includeDirs {
			files = append(files, path)
		}

		if !info.IsDir() && includeFiles {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return files, nil
}

// GetFileName ...
func GetFileName(path string) string {
	filenameWithExt := filepath.Base(path)
	filename := strings.TrimSuffix(filenameWithExt, filepath.Ext(filenameWithExt))
	return filename
}

// Common file extensions
const (
	PngExt  = ".png"
	JSONExt = ".json"
)

// ValidName ...
func ValidName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return false
	}
	return !strings.ContainsAny(name, `/\`)
}

// Exists ...
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory ...
func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}
