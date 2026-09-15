package bashcmd

import (
	"android_vision_scripter/pkg/models"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Touch panel axis codes reported by getevent -p
const (
	absMtPositionX = "0035"
	absMtPositionY = "0036"
	absX           = "0000"
	absY           = "0001"

	releasedTrackingID = "ffffffff"
	synReport          = "SYN_REPORT"
)

var (
	inputDeviceRegexp = regexp.MustCompile(`^add device \d+: (/dev/input/event\d+)`)
	absRangeRegexp    = regexp.MustCompile(`([0-9a-f]{4})\s+: value -?\d+, min (-?\d+), max (-?\d+)`)
	eventLineRegexp   = regexp.MustCompile(`^\[\s*(\d+)\.(\d+)\]\s+(/dev/input/event\d+):\s+\S+\s+(\S+)\s+(\S+)`)
)

type axisRange struct {
	min int
	max int
}

type touchAxes struct {
	x axisRange
	y axisRange
}

type touchPoint struct {
	x      int
	y      int
	hasX   bool
	hasY   bool
	active bool
}

type touchDevice struct {
	axes         touchAxes
	width        int
	height       int
	slot         int
	points       map[int]*touchPoint
	trackingSeen bool
	btnTouch     bool
	pressed      bool
	x            int
	y            int
}

type touchClock struct {
	started bool
	startMs int64
}

func parseAxisRanges(output string) map[string]map[string]axisRange {
	rangesByDevice := map[string]map[string]axisRange{}
	var current string
	for _, line := range strings.Split(output, "\n") {
		if match := inputDeviceRegexp.FindStringSubmatch(line); match != nil {
			current = match[1]
			rangesByDevice[current] = map[string]axisRange{}
			continue
		}
		match := absRangeRegexp.FindStringSubmatch(line)
		if match == nil || current == "" {
			continue
		}
		rangesByDevice[current][match[1]] = axisRange{
			min: atoi(match[2]),
			max: atoi(match[3]),
		}
	}
	return rangesByDevice
}

func pickAxes(
	rangesByDevice map[string]map[string]axisRange,
	xCode string,
	yCode string,
) map[string]touchAxes {
	result := map[string]touchAxes{}
	for path, ranges := range rangesByDevice {
		x, hasX := ranges[xCode]
		y, hasY := ranges[yCode]
		if !hasX || !hasY {
			continue
		}
		result[path] = touchAxes{x: x, y: y}
	}
	return result
}

func parseTouches(
	output string,
	axesByDevice map[string]touchAxes,
	width int,
	height int,
) []models.Event {
	devices := map[string]*touchDevice{}
	clock := &touchClock{}
	var events []models.Event
	for _, line := range strings.Split(output, "\n") {
		match := eventLineRegexp.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		path := match[3]
		axes, ok := axesByDevice[path]
		if !ok {
			continue
		}
		device, ok := devices[path]
		if !ok {
			device = newTouchDevice(axes, width, height)
			devices[path] = device
		}

		code, value := match[4], match[5]
		if code != synReport {
			device.apply(code, value)
			continue
		}
		timeMs := clock.relative(parseTimeMs(match[1], match[2]))
		event, ok := device.commit(timeMs)
		if ok {
			events = append(events, event)
		}
	}
	return events
}

func newTouchDevice(axes touchAxes, width int, height int) *touchDevice {
	return &touchDevice{
		axes:   axes,
		width:  width,
		height: height,
		points: map[int]*touchPoint{},
	}
}

func (c *touchClock) relative(timeMs int64) int64 {
	if !c.started {
		c.started = true
		c.startMs = timeMs
	}
	return timeMs - c.startMs
}

func (d *touchDevice) point() *touchPoint {
	point, ok := d.points[d.slot]
	if ok {
		return point
	}
	point = &touchPoint{}
	d.points[d.slot] = point
	return point
}

func (d *touchDevice) apply(code string, value string) {
	switch code {
	case "ABS_MT_SLOT":
		d.slot = int(parseHex(value))
	case "ABS_MT_TRACKING_ID":
		d.trackingSeen = true
		d.point().active = value != releasedTrackingID
	case "ABS_MT_POSITION_X", "ABS_X":
		point := d.point()
		point.x = scaleAxis(int(parseHex(value)), d.axes.x, d.width)
		point.hasX = true
	case "ABS_MT_POSITION_Y", "ABS_Y":
		point := d.point()
		point.y = scaleAxis(int(parseHex(value)), d.axes.y, d.height)
		point.hasY = true
	case "BTN_TOUCH":
		d.btnTouch = value == "DOWN"
	}
}

func (d *touchDevice) commit(timeMs int64) (models.Event, bool) {
	point, ok := d.points[0]
	if !ok {
		return models.Event{}, false
	}
	pressedNow := d.btnTouch
	if d.trackingSeen {
		pressedNow = point.active
	}

	if !d.pressed && pressedNow {
		if !point.hasX || !point.hasY {
			return models.Event{}, false
		}
		d.pressed = true
		d.x, d.y = point.x, point.y
		return d.event(timeMs, models.ActionDown), true
	}
	if d.pressed && !pressedNow {
		d.pressed = false
		return d.event(timeMs, models.ActionUp), true
	}
	if d.pressed && (point.x != d.x || point.y != d.y) {
		d.x, d.y = point.x, point.y
		return d.event(timeMs, models.ActionMove), true
	}
	return models.Event{}, false
}

func (d *touchDevice) event(timeMs int64, action byte) models.Event {
	return models.NewTouchEvent(timeMs, action, d.x, d.y, d.width, d.height)
}

func scaleAxis(raw int, axis axisRange, size int) int {
	span := axis.max - axis.min
	if span <= 0 || size <= 1 {
		return raw
	}
	scaled := float64(raw-axis.min) / float64(span) * float64(size-1)
	clamped := math.Max(0, math.Min(scaled, float64(size-1)))
	return int(math.Round(clamped))
}

func parseTimeMs(seconds string, fraction string) int64 {
	value, err := strconv.ParseFloat(seconds+"."+fraction, 64)
	if err != nil {
		return 0
	}
	return int64(math.Round(value * 1000))
}

func parseHex(value string) int64 {
	parsed, err := strconv.ParseInt(value, 16, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func atoi(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}
