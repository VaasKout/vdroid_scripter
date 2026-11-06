package bashcmd

import (
	"android_vision_scripter/pkg/models"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	deviceInfoTimeout = 30
	bootCompleted     = "1"
	nullSetting       = "null"
	densityWord       = "density:"
	overrideWord      = "Override"
)

const deviceInfoScript = `echo boot=$(getprop sys.boot_completed); ` +
	`echo model=$(getprop ro.product.model); ` +
	`echo brand=$(getprop ro.product.brand); ` +
	`echo device=$(getprop ro.product.device); ` +
	`echo os=$(getprop ro.build.version.release); ` +
	`echo manufacturer=$(getprop ro.product.manufacturer); ` +
	`echo marketing=$(getprop ro.config.marketing_name); ` +
	`echo locales=$(settings get system system_locales); ` +
	`echo locale=$(getprop persist.sys.locale); ` +
	`echo density=$(wm density)`

type AdbAPI interface {
	TrackDevices() (<-chan []AdbEntry, error)
	GetAdbDevice(serial string) models.AdbDevice

	PushFile(serial string, path string, dest string) error
	ForwardTCPPort(serial string, port int, tag string) error

	RecordGesture(
		serial string,
		idle time.Duration,
		limit time.Duration,
		screenWidth int,
		screenHeight int,
	) (*models.Action, error)
}

func (c *cmdImpl) GetAdbDevice(serial string) models.AdbDevice {
	if serial == "" {
		return models.AdbDevice{}
	}
	command := fmt.Sprintf("timeout %d adb -s %s shell '%s'", deviceInfoTimeout, serial, deviceInfoScript)
	output, err := c.ExecuteCommand(command)
	if err != nil {
		return models.AdbDevice{}
	}
	return parseDeviceInfo(serial, output)
}

func parseDeviceInfo(serial string, output string) models.AdbDevice {
	values := map[string]string{}
	for _, line := range strings.Split(output, "\n") {
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	device := models.AdbDevice{
		Serial:        serial,
		Model:         values["model"],
		Brand:         values["brand"],
		Device:        values["device"],
		OsVersion:     values["os"],
		Manufacturer:  values["manufacturer"],
		MarketingName: values["marketing"],
	}
	if values["boot"] == bootCompleted {
		device.Locale = deviceLocale(values["locales"], values["locale"])
		device.Density = parseDensity(values["density"])
	}
	device.ModelOs = device.ToModelOs()
	return device
}

func deviceLocale(locales string, fallback string) string {
	if locales == "" || locales == nullSetting {
		locales = fallback
	}
	first, _, _ := strings.Cut(locales, ",")
	return strings.TrimSpace(first)
}

func parseDensity(output string) int {
	density := 0
	fields := strings.Fields(output)
	for index := 1; index+1 < len(fields); index++ {
		if fields[index] != densityWord {
			continue
		}
		value, err := strconv.Atoi(fields[index+1])
		if err != nil {
			continue
		}
		if fields[index-1] == overrideWord {
			return value
		}
		density = value
	}
	return density
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

func (c *cmdImpl) RecordGesture(
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
	events, ignored := parseGesture(output, axesByDevice, screenWidth, screenHeight)
	if len(ignored) > 0 {
		c.logger.Info(fmt.Sprintf("keys ignored while recording on %s: %s", serial, strings.Join(ignored, ", ")))
	}
	return &models.Action{
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
		Events:       events,
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
