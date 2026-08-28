package main

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/internal/scrcpy"
	"android_vision_scripter/pkg/logger"
	"sync"
	"testing"
	"time"
)

const (
	TestSerial = "xxx" //serial number of the device
)

func TestConnectToServer(t *testing.T) {
	var config = config.Config{
		FilesProps: &config.FilesProps{
			Logs: "./logs/",
		},
		ScrcpyProps: &config.ScrcpyProps{
			ScrcpyVersion: "3.3.4",
		},
	}
	var fileDB = filesdb.New(config.FilesProps)
	logAPI := logger.New(logger.INFO, true)
	var cmdAPI = bashcmd.New(fileDB, logAPI)
	var cvAPI = cv.New(logAPI)
	var scrcpyAPI = scrcpy.New(cmdAPI, cvAPI, fileDB, config.ScrcpyProps, logAPI)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Go(func() {
		defer wg.Done()
		time.Sleep(10 * time.Second)
	})
	scrcpyAPI.StartScrcpyServer(TestSerial, 1234)
	wg.Wait()
}
