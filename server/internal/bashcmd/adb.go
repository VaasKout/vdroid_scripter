package bashcmd

import (
	"android_vision_scripter/pkg/models"

	"fmt"
	"strconv"
	"strings"
	"time"
)

// Device properties contants
const (
	BrandProp     = "ro.product.brand"
	ModelProp     = "ro.product.model"
	DeviceProp    = "ro.product.device"
	OsVersionProp = "ro.build.version.release"
	LocaleSetting = "system_locales"

	Manufacturer    = "ro.product.manufacturer"
	MarketingName   = "ro.config.marketing_name"
	LineageCodeName = "ro.lineage.device"
)

// Other constants
const (
	OverrideDensityPrefix = "Override density"
)

// AdbAPI ...
type AdbAPI interface {
	GetDevicesList() []string
	GetAdbDevice(serial string) *models.AdbDevice

	PushFile(serial string, path string, dest string) error
	ForwardTCPPort(serial string, port int, tag string) error

	RecordTouches(
		serial string,
		idle time.Duration,
		limit time.Duration,
		screenWidth int,
		screenHeight int,
	) (*models.Action, error)
}

func (c *cmdImpl) GetDevicesList() []string {
	result, err := c.ExecuteCommand("adb devices | tail -n +2 | awk '{print $1}'")
	if err != nil {
		return []string{}
	}

	result = strings.TrimSpace(result)
	if result == "" {
		return []string{}
	}
	return strings.Split(result, "\n")
}

func (c *cmdImpl) GetAdbDevice(serial string) *models.AdbDevice {
	var adbDevice = &models.AdbDevice{}
	if serial == "" || !c.IsAdbConnected(serial) {
		return adbDevice
	}
	adbDevice.Serial = serial
	adbDevice.Model = c.GetProp(serial, ModelProp)
	adbDevice.Brand = c.GetProp(serial, BrandProp)
	adbDevice.Device = c.GetProp(serial, DeviceProp)
	adbDevice.Locale = c.GetSystemSetting(serial, LocaleSetting)
	adbDevice.OsVersion = c.GetProp(serial, OsVersionProp)
	adbDevice.Manufacturer = c.GetProp(serial, Manufacturer)
	adbDevice.MarketingName = c.GetProp(serial, MarketingName)
	adbDevice.Density = c.GetDensity(serial)
	return adbDevice
}

func (c *cmdImpl) GetDensity(serial string) int {
	command := fmt.Sprintf("adb -s %s shell wm density", serial)
	result, err := c.ExecuteCommand(command)
	if err != nil {
		return 0
	}
	return parseDensity(result)
}

func parseDensity(output string) int {
	density := 0
	for _, line := range strings.Split(output, "\n") {
		name, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(name), OverrideDensityPrefix) {
			return parsed
		}
		density = parsed
	}
	return density
}

func (c *cmdImpl) GetProp(serial string, prop string) string {
	var command = fmt.Sprintf("adb -s %s shell getprop %s", serial, prop)
	result, err := c.ExecuteCommand(command)
	if err != nil {
		return ""
	}
	result = strings.TrimSpace(result)
	return result
}

func (c *cmdImpl) GetSystemSetting(serial string, prop string) string {
	var command = fmt.Sprintf("adb -s %s shell settings get system %s", serial, prop)
	result, err := c.ExecuteCommand(command)
	if err != nil {
		return ""
	}
	result = strings.TrimSpace(result)
	return result
}

func (c *cmdImpl) IsAdbConnected(serial string) bool {
	var command = fmt.Sprintf("adb devices | grep -w '%s'", serial)
	result, err := c.ExecuteCommand(command)
	if err != nil {
		return false
	}
	return result != ""
}

func (c *cmdImpl) PushFile(serial string, path string, dest string) error {
	var command = fmt.Sprintf("adb -s %s push %s %s", serial, path, dest)
	_, err := c.ExecuteCommand(command)
	return err
}

func (c *cmdImpl) ForwardTCPPort(serial string, port int, tag string) error {
	var command = fmt.Sprintf("adb -s %s forward tcp:%d localabstract:%s", serial, port, tag)
	_, err := c.ExecuteCommand(command)
	return err
}

func (c *cmdImpl) RecordTouches(
	serial string,
	idle time.Duration,
	limit time.Duration,
	screenWidth int,
	screenHeight int,
) (*models.Action, error) {
	axesByDevice, err := c.touchAxes(serial)
	if err != nil {
		return nil, err
	}

	var command = fmt.Sprintf("adb -s %s shell getevent -lt", serial)
	output, err := c.executeUntilIdle(command, idle, limit, axesByDevice)
	if err != nil {
		return nil, err
	}
	return &models.Action{
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		Events:       parseTouches(output, axesByDevice, screenWidth, screenHeight),
	}, nil
}

func (c *cmdImpl) touchAxes(serial string) (map[string]touchAxes, error) {
	output, err := c.ExecuteCommand(fmt.Sprintf("adb -s %s shell getevent -p", serial))
	if err != nil {
		return nil, err
	}
	rangesByDevice := parseAxisRanges(output)
	axes := pickAxes(rangesByDevice, absMtPositionX, absMtPositionY)
	if len(axes) == 0 {
		axes = pickAxes(rangesByDevice, absX, absY)
	}
	if len(axes) == 0 {
		return nil, fmt.Errorf("no touch input device found on %s", serial)
	}
	return axes, nil
}
