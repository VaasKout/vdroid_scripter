package bashcmd

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/logger"
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
	var filesDB = filesdb.New(fileProps)

	screenshot := takeScreenshot(filesDB, TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}

	t.Log(screenshot)
}

func takeScreenshot(filesDB filesdb.FilesDB, serial string) string {
	dir := filesDB.CreateLogsDir(serial, "screenshot")
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
