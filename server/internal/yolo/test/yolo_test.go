package test

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/internal/yolo"
	"android_vision_scripter/pkg/logger"
	"testing"
	"time"

	"gocv.io/x/gocv"
)

const (
	TestSerial = "emulator-5554"
)

func TestDetectLabels(t *testing.T) {
	cfg := config.New()
	logAPI := logger.New(logger.INFO, true)
	filesDB := filesdb.New(cfg.FilesProps)
	cmdRunner := bashcmd.New(filesDB, logAPI)
	yoloAPI := yolo.New(filesDB, logAPI)

	screenshot := cmdRunner.ScreenShot(TestSerial)
	if screenshot == "" {
		t.Fatal("screenshot is empty")
	}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("could not read screenshot image")
	}
	defer img.Close()

	start := time.Now()
	detections := yoloAPI.DetectLabels(img)
	elapsed := time.Since(start)

	t.Logf("elapsed: %dms", elapsed.Milliseconds())
	t.Logf("found %d detections", len(detections))
	for _, detection := range detections {
		t.Logf(
			"label=%s confidence=%.2f rectangle=%+v",
			detection.Label,
			detection.Confidence,
			detection.Rectangle,
		)
	}
}
