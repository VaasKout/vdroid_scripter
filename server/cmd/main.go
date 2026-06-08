// Package main ...
package main

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/internal/scrcpy"
	"android_vision_scripter/internal/server"
	"android_vision_scripter/internal/usecases"
	"android_vision_scripter/internal/yolo"
	"android_vision_scripter/pkg/core/network"
	"android_vision_scripter/pkg/logger"
	"fmt"
	"runtime"
	"time"

	// _ "net/http/pprof" uncomment for profiling

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
		godotenv.Load("~/.env")
	}

	// Uncomment for pprof profiling
	// go func() {
	//     log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	cfg := config.New()
	logAPI := logger.New(logger.INFO, true)

	filesDB := filesdb.New(cfg.FilesProps)
	cmdRunner := bashcmd.New(filesDB, logAPI)
	cvAPI := cv.New(cmdRunner, logAPI)
	scrcpy := scrcpy.New(cmdRunner, cvAPI, filesDB, cfg.ScrcpyProps, logAPI)
	yoloAPI := yolo.New(filesDB, logAPI)
	network := network.New(logAPI)

	interactor := usecases.New(cvAPI, cmdRunner, filesDB, scrcpy, yoloAPI, network, logAPI)
	serverAPI := server.New(interactor, cfg.ServerProps, logAPI)

	// Uncomment to watch alloc space in real time
	// go logHeap()
	serverAPI.ListenAndServe()
}

func logHeap() {
	var m runtime.MemStats

	for {
		time.Sleep(1 * time.Second)
		runtime.ReadMemStats(&m)
		fmt.Println("-----------------------------------")
		fmt.Printf("Heap Alloc = %v KB\n", m.HeapAlloc/1024)
		fmt.Printf("Heap Sys   = %v KB\n", m.HeapSys/1024)
		fmt.Printf("Heap Idle  = %v KB\n", m.HeapIdle/1024)
		fmt.Printf("Heap Inuse = %v KB\n", m.HeapInuse/1024)
		fmt.Printf("Heap Released = %v KB\n", m.HeapReleased/1024)
		fmt.Printf("Heap Objects = %v\n", m.HeapObjects)
	}
}
