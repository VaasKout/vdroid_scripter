package test

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/internal/yolo"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/logger"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"gocv.io/x/gocv"
)

const (
	TestSerial  = "emulator-5554"
	TestLogsDir = "./logs"
)

func TestDetectLabels(t *testing.T) {
	cfg := config.New()
	logAPI := logger.New(logger.INFO, true)
	filesDB := filesdb.New(cfg.FilesProps)
	yoloAPI := yolo.New(filesDB, logAPI)

	screenshot := takeScreenshot(TestLogsDir, TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	start := time.Now()
	rects := yoloAPI.DetectLabels(img)
	elapsed := time.Since(start)

	t.Logf("elapsed: %dms", elapsed.Milliseconds())
	t.Logf("found %d detections", len(rects))
	for _, rect := range rects {
		t.Logf("label=%s rectangle=%+v", rect.Label, rect)
	}
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
