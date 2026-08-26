// Package usecases ...
package usecases

import (
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/internal/scrcpy"
	"android_vision_scripter/internal/yolo"
	"android_vision_scripter/pkg/core/cache"
	"android_vision_scripter/pkg/core/network"
	"android_vision_scripter/pkg/logger"
	"android_vision_scripter/pkg/models"
)

// Common errors
const (
	SerialIsEmptyError  = "serial is empty"
	DeviceNotFoundError = "device not found"
)

// Interactor ...
type Interactor interface {
	CmdUseCase
	ScrcpyUseCase
	KeyboardUseCase
	SessionUseCase
	LibraryUseCase
	ScanUseCase
	RouteUseCase
	StepsUseCase
	OcrUseCase
}

type interactorImpl struct {
	cv      cv.API
	cmd     bashcmd.CmdAPI
	filesDB filesdb.FilesDB
	scrcpy  scrcpy.Scrcpy
	yolo    yolo.Yolo
	network network.Client
	logger  *logger.Logger

	devicesCache  cache.Cache[models.AdbDevice]
	sessionsCache cache.Cache[models.Session]
}

// New instance of interactor
func New(
	cv cv.API,
	cmd bashcmd.CmdAPI,
	filesDB filesdb.FilesDB,
	scrcpy scrcpy.Scrcpy,
	yolo yolo.Yolo,
	network network.Client,
	logger *logger.Logger,
) Interactor {
	var devicesCache = cache.NewSafeCache[models.AdbDevice]()
	var sessionsCache = cache.NewSafeCache[models.Session]()

	var interactor = &interactorImpl{
		cv:      cv,
		cmd:     cmd,
		filesDB: filesDB,
		logger:  logger,
		scrcpy:  scrcpy,
		yolo:    yolo,
		network: network,

		devicesCache:  devicesCache,
		sessionsCache: sessionsCache,
	}

	go interactor.FillUpDevicesCache()
	return interactor
}
