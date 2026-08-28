// Package file ...
package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateFileIfNotExist ...
func CreateFileIfNotExist(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		file, err := os.Create(file)
		if err != nil {
			return false
		}
		defer file.Close()
	}
	return true
}

// CreateDirIfNotExist ...
func CreateDirIfNotExist(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		return err == nil
	}
	return true
}

// FindDirectoryOfFile ...
func FindDirectoryOfFile(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	// Extract directory
	dir := filepath.Dir(absPath)
	return dir, nil
}

// GetFilesInDirectory ...
func GetFilesInDirectory(dir string, recursive bool) ([]string, error) {
	if recursive {
		return getAllInDirectoryRecursively(dir, false, true)
	}
	return getAllInDirectory(dir, false, true)
}

// GetDirsInDirectory ...
func GetDirsInDirectory(dir string, recursive bool) ([]string, error) {
	if recursive {
		return getAllInDirectoryRecursively(dir, true, false)
	}
	return getAllInDirectory(dir, true, false)
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

// FindFileInDirectory ...
func FindFileInDirectory(directoryPath string, fileKey string) (string, error) {
	var foundPath string

	err := filepath.WalkDir(directoryPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		var baseFileName = filepath.Base(path)
		var appendix = ""
		if !strings.Contains(fileKey, ".") {
			appendix = "."
		}
		if !d.IsDir() && strings.HasPrefix(baseFileName, fileKey+appendix) {
			foundPath = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	if foundPath == "" {
		return "", fmt.Errorf("file %s not found", filepath.Join(directoryPath, fileKey))
	}
	return foundPath, nil
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

// CountFilesInDirectory ...
func CountFilesInDirectory(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Error reading directory: "+path, err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() { // only count files
			count++
		}
	}

	return count
}
