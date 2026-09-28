package bashcmd

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/logger"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	TestSerial = "xxx" //serial number of the device
)

func TestGetDeviceList(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs/",
	}
	logAPI := logger.New(logger.INFO, true)
	var filesDB = filesdb.New(fileProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)

	devices := cmdRunner.GetDevicesList()
	t.Log(devices)
	t.Log(len(devices))
}

func TestScreenshot(t *testing.T) {
	var fileProps = &config.FilesProps{
		Logs: "./logs/",
	}

	screenshot := takeScreenshot(fileProps.Logs, TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}

	t.Log(screenshot)
}

func createLogsDir(logsDir string, args ...string) string {
	var dirName = filepath.Join(logsDir, filepath.Join(args...))
	if ok := file.CreateDirIfNotExist(dirName); !ok {
		fmt.Printf("Couldn't create dir %s\n", dirName)
		return ""
	}
	return dirName
}

func takeScreenshot(logsDir string, serial string) string {
	dir := createLogsDir(logsDir, serial, "screenshot")
	if dir == "" {
		return ""
	}
	data, err := exec.Command("adb", "-s", serial, "exec-out", "screencap", "-p").Output()
	if err != nil {
		return ""
	}
	path := filepath.Join(dir, "screenshot.png")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return ""
	}
	return path
}
