package bashcmd

import (
	"android_vision_scripter/pkg/models"
	"math"
	"regexp"
	"sort"
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
	evKey              = "EV_KEY"
	keyDown            = "DOWN"
	keyUp              = "UP"
)

var (
	inputDeviceRegexp = regexp.MustCompile(`^add device \d+: (/dev/input/event\d+)`)
	absRangeRegexp    = regexp.MustCompile(`([0-9a-f]{4})\s+: value -?\d+, min (-?\d+), max (-?\d+)`)
	eventLineRegexp   = regexp.MustCompile(`^\[\s*(\d+)\.(\d+)\]\s+(/dev/input/event\d+):\s+(\S+)\s+(\S+)\s+(\S+)`)
)

var androidKeycodes = map[string]int{
	"KEY_HOME": 3, "0066": 3,
	"KEY_HOMEPAGE": 3, "00ac": 3,
	"KEY_BACK": 4, "009e": 4,
	"KEY_MENU": 82, "008b": 82,
	"KEY_APPSELECT": 187, "0244": 187, "00fe": 187,
	"KEY_VOLUMEUP": 24, "0073": 24,
	"KEY_VOLUMEDOWN": 25, "0072": 25,
	"KEY_POWER": 26, "0074": 26,
	"KEY_CAMERA": 27, "00d4": 27,
	"KEY_SEARCH": 84, "00d9": 84,
}

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

type gestureClock struct {
	started bool
	startMs int64
}

type eventLine struct {
	timeMs int64
	path   string
	kind   string
	code   string
	value  string
}

func parseEventLine(line string) (eventLine, bool) {
	match := eventLineRegexp.FindStringSubmatch(line)
	if match == nil {
		return eventLine{}, false
	}
	return eventLine{
		timeMs: parseTimeMs(match[1], match[2]),
		path:   match[3],
		kind:   match[4],
		code:   match[5],
		value:  match[6],
	}, true
}

func (l eventLine) keycode() (int, bool) {
	if l.kind != evKey {
		return 0, false
	}
	keycode, ok := androidKeycodes[l.code]
	return keycode, ok
}

func isGestureLine(line string, axesByDevice map[string]touchAxes) bool {
	parsed, ok := parseEventLine(line)
	if !ok {
		return false
	}
	if _, key := parsed.keycode(); key {
		return true
	}
	_, touch := axesByDevice[parsed.path]
	return touch
}

func keyAction(value string) (byte, bool) {
	switch value {
	case keyDown:
		return models.KeyActionDown, true
	case keyUp:
		return models.KeyActionUp, true
	}
	return 0, false
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

type gestureParser struct {
	axesByDevice map[string]touchAxes
	width        int
	height       int
	devices      map[string]*touchDevice
	clock        gestureClock
	ignored      map[string]bool
	events       []models.Event
}

func parseGesture(
	output string,
	axesByDevice map[string]touchAxes,
	width int,
	height int,
) ([]models.Event, []string) {
	parser := &gestureParser{
		axesByDevice: axesByDevice,
		width:        width,
		height:       height,
		devices:      map[string]*touchDevice{},
		ignored:      map[string]bool{},
	}
	for _, line := range strings.Split(output, "\n") {
		parsed, ok := parseEventLine(line)
		if !ok {
			continue
		}
		parser.add(parsed)
	}
	return parser.events, sortedKeys(parser.ignored)
}

func (p *gestureParser) add(line eventLine) {
	if keycode, key := line.keycode(); key {
		p.addKey(line, keycode)
		return
	}
	if axes, touch := p.axesByDevice[line.path]; touch {
		p.addTouch(line, axes)
		return
	}
	if line.kind == evKey {
		p.ignored[line.code] = true
	}
}

func (p *gestureParser) addKey(line eventLine, keycode int) {
	action, pressed := keyAction(line.value)
	if !pressed {
		return
	}
	timeMs := p.clock.relative(line.timeMs)
	p.events = append(p.events, models.NewKeyEvent(timeMs, action, keycode))
}

func (p *gestureParser) addTouch(line eventLine, axes touchAxes) {
	device := p.device(line.path, axes)
	if line.code != synReport {
		device.apply(line.code, line.value)
		return
	}
	event, ok := device.commit(p.clock.relative(line.timeMs))
	if !ok {
		return
	}
	p.events = append(p.events, event)
}

func (p *gestureParser) device(path string, axes touchAxes) *touchDevice {
	device, ok := p.devices[path]
	if ok {
		return device
	}
	device = newTouchDevice(axes, p.width, p.height)
	p.devices[path] = device
	return device
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func newTouchDevice(axes touchAxes, width int, height int) *touchDevice {
	return &touchDevice{
		axes:   axes,
		width:  width,
		height: height,
		points: map[int]*touchPoint{},
	}
}

func (c *gestureClock) relative(timeMs int64) int64 {
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
