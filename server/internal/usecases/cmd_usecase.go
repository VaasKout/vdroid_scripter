package usecases

import (
	"android_vision_scripter/pkg/models"
	"fmt"
	"time"

	"gocv.io/x/gocv"
)

// CmdUseCase ...
type CmdUseCase interface {
	FillUpDevicesCache()
	GetDevices() []models.AdbDevice
	GetDevice(serial string) *models.AdbDevice
	TakeScreenshot(serial string, withRectangles bool) (string, []models.Rectangle, error)
}

func (i *interactorImpl) TakeScreenshot(
	serial string,
	withRectangles bool,
) (string, []models.Rectangle, error) {
	screenshot := i.cmd.ScreenShot(serial)
	if screenshot == "" {
		return "", nil, fmt.Errorf("couldn't take screenshot for %s", serial)
	}
	if !withRectangles {
		return screenshot, nil, nil
	}

	img := gocv.IMRead(screenshot, gocv.IMReadColor)
	if img.Empty() {
		return "", nil, fmt.Errorf("couldn't read screenshot for %s", serial)
	}
	defer img.Close()

	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)

	imgRectangles, err := i.cv.FindAllRectangles(&gray)
	if err != nil {
		return "", nil, err
	}
	return screenshot, models.ImgRectanglesToDomain(imgRectangles), nil
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

func (i *interactorImpl) GetDevice(serial string) *models.AdbDevice {
	var device = &models.AdbDevice{}
	if result, ok := i.devicesCache.Get(serial); ok {
		if result.Serial != "" {
			device = &result
		}
	}

	if device.Serial == "" {
		i.cmd.GetAdbDevice(serial)
	}

	i.devicesCache.Add(serial, *device)
	return device
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
