package usecases

import (
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/pkg/models"
	"fmt"
	"sync"
	"time"
)

const (
	parallelDeviceFetches = 8
	trackerRestartDelay   = 2 * time.Second
	deviceRetryInterval   = 10 * time.Second
)

type deviceFetches struct {
	mu     sync.Mutex
	done   *sync.Cond
	active map[string]bool
	slots  chan struct{}
}

func newDeviceFetches() *deviceFetches {
	instance := &deviceFetches{
		active: map[string]bool{},
		slots:  make(chan struct{}, parallelDeviceFetches),
	}
	instance.done = sync.NewCond(&instance.mu)
	return instance
}

// DevicesUseCase ...
type DevicesUseCase interface {
	GetDevices() []models.AdbDevice
}

func (i *interactorImpl) GetDevices() []models.AdbDevice {
	i.waitDeviceFetches()
	devices := i.devicesCache.GetDataArray()
	for index := range devices {
		_, running := i.sessionsCache.Get(devices[index].Serial)
		devices[index].ScrcpyRunning = running
	}
	return devices
}

func (i *interactorImpl) watchDevices() {
	for {
		updates, err := i.cmd.TrackDevices()
		if err != nil {
			i.logger.Error(fmt.Sprintf("device tracker failed to start: %s, retrying in %s", err.Error(), trackerRestartDelay))
			time.Sleep(trackerRestartDelay)
			continue
		}
		for listing := range updates {
			i.applyDeviceListing(listing)
		}
		i.logger.Error(fmt.Sprintf("device tracker ended, restarting in %s", trackerRestartDelay))
		time.Sleep(trackerRestartDelay)
	}
}

func (i *interactorImpl) applyDeviceListing(listing []bashcmd.AdbEntry) {
	i.deviceFetches.mu.Lock()
	defer i.deviceFetches.mu.Unlock()
	present := map[string]bool{}
	for _, item := range listing {
		if item.State != bashcmd.AdbStateDevice {
			continue
		}
		present[item.Serial] = true
		if _, known := i.devicesCache.Get(item.Serial); known {
			continue
		}
		i.devicesCache.Add(item.Serial, models.AdbDevice{Serial: item.Serial, Model: item.Model, Device: item.Device})
		i.startDeviceFetch(item.Serial)
	}
	for _, device := range i.devicesCache.GetDataArray() {
		if !present[device.Serial] {
			i.devicesCache.Delete(device.Serial)
		}
	}
}

func (i *interactorImpl) retryDeviceFetches() {
	for {
		time.Sleep(deviceRetryInterval)
		i.deviceFetches.mu.Lock()
		for _, device := range i.devicesCache.GetDataArray() {
			if !device.Ready() {
				i.startDeviceFetch(device.Serial)
			}
		}
		i.deviceFetches.mu.Unlock()
	}
}

func (i *interactorImpl) startDeviceFetch(serial string) {
	if i.deviceFetches.active[serial] {
		return
	}
	i.deviceFetches.active[serial] = true
	go i.fetchDevice(serial)
}

func (i *interactorImpl) fetchDevice(serial string) {
	i.deviceFetches.slots <- struct{}{}
	i.refreshDevice(serial)
	<-i.deviceFetches.slots
	i.deviceFetches.mu.Lock()
	delete(i.deviceFetches.active, serial)
	i.deviceFetches.done.Broadcast()
	i.deviceFetches.mu.Unlock()
}

func (i *interactorImpl) refreshDevice(serial string) {
	fresh := i.cmd.GetAdbDevice(serial)
	i.deviceFetches.mu.Lock()
	defer i.deviceFetches.mu.Unlock()
	device, known := i.devicesCache.Get(serial)
	if !known {
		return
	}
	device.Update(fresh)
	i.devicesCache.Add(serial, device)
	if device.Ready() {
		i.logger.Info(fmt.Sprintf(
			"device %s ready: %s, locale %s, density %d ✅", serial, device.ModelOs, device.Locale, device.Density,
		))
		return
	}
	i.logger.Info(fmt.Sprintf("device %s not ready yet, next look in %s ⏳", serial, deviceRetryInterval))
}

func (i *interactorImpl) waitDeviceFetches() {
	i.deviceFetches.mu.Lock()
	defer i.deviceFetches.mu.Unlock()
	for len(i.deviceFetches.active) > 0 {
		i.deviceFetches.done.Wait()
	}
}
