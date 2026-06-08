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
	ScriptsUseCase
	ScrcpyUseCase
	KeyboardUseCase
}

type interactorImpl struct {
	cv      cv.API
	cmd     bashcmd.CmdAPI
	filesDB filesdb.FilesDB
	scrcpy  scrcpy.Scrcpy
	yolo    yolo.Yolo
	network network.Client
	logger  *logger.Logger

	devicesCache cache.Cache[models.AdbDevice]
	clientsCache cache.Cache[ClientConnection]
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
	var clientCache = cache.NewSafeCache[ClientConnection]()

	var interactor = &interactorImpl{
		cv:      cv,
		cmd:     cmd,
		filesDB: filesDB,
		logger:  logger,
		scrcpy:  scrcpy,
		yolo:    yolo,
		network: network,

		devicesCache: devicesCache,
		clientsCache: clientCache,
	}

	go interactor.FillUpDevicesCache()
	return interactor
}
