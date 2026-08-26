// Package config ...
package config

import (
	"os"
	"path/filepath"
	"strconv"
)

// Default values
const (
	BasePath       = "vdroid_scripter"
	LogsDir        = "logs"
	KeyboardDir    = "keyboards"
	ScrcpyDir      = "scrcpy"
	YoloDir        = "yolo"
	ImagesDir      = "images"
	ActionsDir     = "actions"
	RoutesDir      = "routes"
	ScrcpyVersion  = "3.3.4"
	ServerPort     = ":8080"
	BaseSocketPort = 3001
)

// Config ...
type Config struct {
	ServerProps *ServerProps
	FilesProps  *FilesProps
	ScrcpyProps *ScrcpyProps
}

// ServerProps ...
type ServerProps struct {
	Port       string
	SocketPort int
}

// FilesProps ...
type FilesProps struct {
	Logs      string
	Keyboards string
	Scrcpy    string
	Yolo      string
	Images    string
	Actions   string
	Routes    string
}

// ScrcpyProps ...
type ScrcpyProps struct {
	ScrcpyVersion string
}

// New ...
func New() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = ServerPort
	}

	var baseSocketPort = BaseSocketPort
	socketPort := os.Getenv("SOCKET_PORT")
	if socketPort != "" {
		socketPortInt, err := strconv.ParseInt(socketPort, 10, 32)
		if err == nil {
			baseSocketPort = int(socketPortInt)
		}
	}

	basePath := os.Getenv("BASE_PATH")
	if basePath == "" {
		basePath = BasePath
	}
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}
	cachePath := filepath.Join(userCacheDir, basePath)

	logsDir := os.Getenv("LOGS")
	if logsDir == "" {
		logsDir = LogsDir
	}

	keyboardsDir := os.Getenv("KEYBOARDS_DIR")
	if keyboardsDir == "" {
		keyboardsDir = KeyboardDir
	}

	scrcpyDir := os.Getenv("SCRCPY_DIR")
	if scrcpyDir == "" {
		scrcpyDir = ScrcpyDir
	}

	yoloDir := os.Getenv("YOLO_DIR")
	if yoloDir == "" {
		yoloDir = YoloDir
	}

	imagesDir := os.Getenv("IMAGES_DIR")
	if imagesDir == "" {
		imagesDir = ImagesDir
	}

	actionsDir := os.Getenv("ACTIONS_DIR")
	if actionsDir == "" {
		actionsDir = ActionsDir
	}

	scrcpyVersion := os.Getenv("SCRCPY_VERSION")
	if scrcpyVersion == "" {
		scrcpyVersion = ScrcpyVersion
	}

	routesDir := os.Getenv("ROUTES_DIR")
	if routesDir == "" {
		routesDir = RoutesDir
	}

	return &Config{
		ServerProps: &ServerProps{
			Port:       port,
			SocketPort: baseSocketPort,
		},
		FilesProps: &FilesProps{
			Logs:      filepath.Join(cachePath, logsDir),
			Keyboards: filepath.Join(cachePath, keyboardsDir),
			Scrcpy:    filepath.Join(cachePath, scrcpyDir),
			Yolo:      filepath.Join(cachePath, yoloDir),
			Images:    filepath.Join(cachePath, imagesDir),
			Actions:   filepath.Join(cachePath, actionsDir),
			Routes:    filepath.Join(cachePath, routesDir),
		},
		ScrcpyProps: &ScrcpyProps{
			ScrcpyVersion: scrcpyVersion,
		},
	}
}
