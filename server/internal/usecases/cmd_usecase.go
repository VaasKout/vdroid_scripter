package usecases

import (
	"android_vision_scripter/pkg/models"
	"time"
)

// CmdUseCase ...
type CmdUseCase interface {
	GetDevices() []models.AdbDevice
}

func (i *interactorImpl) FillUpDevicesCache() {
	for {
		devices := i.cmd.GetDevicesList()

	devicesLoop:
		for key := range i.devicesCache.GetMap() {
			for _, serial := range devices {
				if serial == key {
					continue devicesLoop
				}
			}
			i.devicesCache.Delete(key)
		}

		for _, serial := range devices {
			device := i.cmd.GetAdbDevice(serial)
			device.FilterLocales()
			i.devicesCache.Add(serial, *device)
		}

		time.Sleep(1 * time.Minute)
	}
}

func (i *interactorImpl) GetDevices() []models.AdbDevice {
	return i.devicesCache.GetDataArray()
}

func (i *interactorImpl) setScrcpyState(serial string, isRunning bool) {
	var device = &models.AdbDevice{}
	if result, ok := i.devicesCache.Get(serial); ok {
		if result.Serial != "" {
			device = &result
		}
	}

	if device.Serial == "" {
		device = i.cmd.GetAdbDevice(serial)
	}

	device.ScrcpyRunning = isRunning
	i.devicesCache.Add(serial, *device)
}
