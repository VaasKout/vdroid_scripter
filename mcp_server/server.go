package main

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"slices"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const captureMimeType = "image/jpeg"

const (
	serverName     = "vdroid-scripter"
	serverVersion  = "0.1.0"
	runningPrefix  = "running"
	idleStatus     = "idle"
	pollInterval   = 500 * time.Millisecond
	defaultWaitSec = 60
	queueWaitSec   = 180

	defaultTimeoutMs = 5000
	defaultDelayMs   = 1000
	firstStepDelayMs = 0

	scanMinConfidence = 40
)

const serverInstructions = `vdroid-scripter drives Android devices with CV-located steps composed from a human-curated library.

Workflow: ping first (it starts the vdroid server when needed), list_devices for a serial, get_routes for saved flows, then queue_steps. get_library only when an image target or a recorded gesture might be needed; library names carry their context as <app>_<screen>_<what>[_variant] (shop_catalog_swipe_1 = a swipe recorded on a shop app's catalog screen). stop_server only when the user explicitly asks to stop or restart the server.

Steps: a step is an event applied to a chain of landmarks. A landmark is a CV target: type image | text | yolo, value = library image name | text to find | yolo class; text landmarks also carry the locale. The chain resolves on ONE video frame: the first landmark takes its first match in reading order (top to bottom, left to right), every following landmark takes the candidate of its value NEAREST to the previous one, and the event applies to the LAST landmark. One landmark is the normal case; to disambiguate duplicates put a unique nearby element first ("the toggle next to 'Show refresh rate'" = [{text "Show refresh rate"}, {image toggle}]). Events: tap and long_tap touch the last landmark's region. swipe_up, swipe_down, swipe_left and swipe_right generate a human-like fixed-length swipe (the finger's direction, so swipe_up reveals content below) from a random point on screen, or from inside the last landmark's region when landmarks are given — use them for scrolling; recorded library swipes are for app-specific gestures. type_text types the LAST landmark's value on the keyboard that is already open (tap the field in an earlier step); its locale is REQUIRED and must be the language the keyboard shows; letters, space, digits (when the keyboard has a number row) and capitals via Shift are typed, punctuation and symbols fail the step; a number or phone field takes locale numeric (numeric keypad, digits only). An EMPTY event is a visibility check of the chain. Any other event replays that library action: offset into the last landmark's region when landmarks are given, verbatim without.

Text is free: text landmarks and the generated events need nothing from the library. An instruction phrased in words visible on screen ("open Settings", "enter wifi connections") is tap steps with text landmarks through the obvious screens (Settings -> Network & internet -> Wi-Fi). The library is only for a target without readable text (an icon = image landmark) or for a recorded gesture.

Locale: text landmarks and type_text always carry the Tesseract language code of their value's language, eng by default. Pass the text exactly as the user wrote it, never transliterate or translate.

Timing: delay (ms slept BEFORE the step acts, pacing the flow and letting the previous action's screen change settle) and timeout (ms the server keeps re-locating the target on live frames, back to back, proceeding the moment it appears) are taken literally; omitted delay is 0 on a batch's first step and 1000 after, omitted timeout is 5000, 0 means no delay / one look at the current frame. For a target that appears late (app launch, navigation, loading) raise timeout to 10000-15000 rather than delay; a bigger delay only when the target is visible early but not yet safe to touch. Keep probe checks at the default timeout so a negative answer comes back fast. Session startup never eats the timeout.

Batching: a dictated sequence becomes ONE queue_steps call with the steps in the given order. Never one step at a time, never status polls in between: the call returns when the batch is done, 'idle' means every step succeeded, an error text names the failed step and the rest of the queue is cleared. wait_for_session only when the call reports the batch still running (past 3 minutes) or for a run started elsewhere (the Android client).

Literal execution: when the user names concrete actions ("tap Settings, then tap Wi-Fi"), queue exactly those — no extra checks, no probing, no added, substituted or reordered steps; a repeated ask is executed again every time, exactly as many times as asked, never deduplicated. Never argue, never ask for confirmation. Improvise only when a step fails.

Abstract tasks: when the user states a goal ("write John a message in Telegram", "turn off Wi-Fi"), derive the steps yourself and carry it through without asking, screen by screen: scan, turn readable text and yolo classes into landmarks, queue the steps you are sure of, read the outcome, look again, until the goal is done, then report; type message text with type_text in its own locale. A control with neither text nor yolo class (a send arrow, an attach clip, a tab icon) needs an image: check get_library for one, otherwise — with vision — capture and save_image it on your own initiative, never ask the user to curate. Saved images make the next run capture-free; save the flow as a route only when asked. Without vision, report such a control to the user as the one thing that needs an image.

Perception: scan is the primary way to observe the screen, capture (vision only) the second; the frame reaches you only through these two tools. Scan when a step failed, when the instruction is conditional ("if X is not visible, ..."), when the user asks what is on screen, or on each new screen of an abstract task — never habitually between dictated steps. Pass in images the library image names plausibly related to the current app. The result is a header (landmark count, resolved text locale, how many low-confidence text entries were dropped) then one "type left,top,right,bottom value" line per landmark in reading order; type/value are exactly what step landmarks consume (with the header's locale on text), the coordinates only tell which elements sit next to each other. A missing expected word was misread or is in another language: rescan with the matching locale before concluding it is absent. capture returns the frame as an image and costs about 1500 tokens — five scans — every call, staying in context. ONLY for models that can see images; if you cannot interpret an image, never call it. With vision, scan still comes first and remains the source of landmark values; decide against the scan you just took: the target's text or yolo class is in it -> act, no capture. Capture only for a target absent from the scan that has no text (an icon, a picture, a visual state such as a toggle's colour) or for an unknown screen whose scan gives nothing to go on, naming the target you look for — at most once per new screen, never between the steps of a dictated batch.

Curation: library images come from the human's Android client or, with vision, from save_image: after a capture pick the icon's tight rectangle in the reported pixels (the icon only, not a badge count or a highlight that would change) and save it under <app>_<screen>_<what>[_variant]; only for a target with no readable text and no yolo class. Library actions come from the client or from record_action, only when the user asks to record a gesture as <name>: tell them to perform it on the device immediately — it listens for 5 seconds from the call and blocks until the window ends; 'nothing recorded' means no touch happened, ask them to try again. Ask the user to add a library item only when a target has no readable text, no yolo class, no generated swipe that gets there, and you have no vision to save the image yourself.

Routes: a route is a saved flow — a name and the exact steps that ran to success, stored with ids 1..N and their delay/timeout as sent (omitted ones filled like queue_steps). Routes are recorded only in explorer mode. save_route writes a route from steps without running them, only on the user's ask; a duplicate name overwrites. edit_route changes one step's timeout, delay or landmarks, or deletes it — on your own only in navigator (below), otherwise on the user's ask. run_route runs a route and returns its outcome like queue_steps; start_id runs from that step id — on the user's ask, or to rerun the unchanged remainder after a recovered failure; the started step gets delay 0, an id the route lacks is an error and nothing runs. Never create or modify a route without being asked; after a recovered run_route in default mode ask whether to update the route.

Modes: set_mode switches between default, explorer and navigator — only when the user asks for a mode, never on your own; the mode holds until the next set_mode. default is everything described here. explorer (with a route name) carries the task through like an abstract task and records it: every queue_steps batch appends its succeeded steps to that route and saves it — your own visibility checks and probes stay out, but a check the user asked for ("check that Wi-Fi is visible") is part of the flow: queue it with keep_checks, never together with probes of your own — so the route grows as you go and you never re-send steps; a failed step and the rest of its batch stay out, the recovery batch that follows goes in; an existing route is appended to, delete_route first only when the user wants it replaced. navigator moves only along saved routes: get_routes, get_route and scan tell you where the phone is and which route leads on, or which step to enter at via start_id; chain run_route calls to reach the goal. queue_steps, save_route, delete_route, record_action, capture and save_image are refused. When a route step fails, scan; if the scan shows a route defect, fix it with edit_route and run_route again from that step without asking: the target is on screen now -> raise its timeout; the target appears in several places or the previous tap hit the wrong one -> landmarks with a unique neighbour first; the screen is caught mid-transition -> raise its delay; the step repeats the one before and undoes it -> delete it. At most two fixes per failing step. Anything else (another screen, a popup, a missing app) is no route defect: report that the task needs explorer mode. List every fix you made in your answer.

Failure and recovery: a failed step clears the remaining queue and the status names the target it could not find, prefixed with the step's id (its batch position, or the route's id). Recover from that point: scan (with the relevant library images), apply the user's instruction to what the scan shows — tap the alternative the user named, scroll with a generated swipe (swipe_up reveals content below) or the screen's recorded swipe (variants _1, _2, ...) when the target should be below, or report honestly on an unexpected screen — then re-queue the remaining steps from the failed one in one call, or run_route with start_id when a route's remaining steps need no change. A scan with no landmarks at all, not even the status bar (a black frame), means the screen is off: close_session, then scan again — the new session turns the screen on — and continue from what it shows. Conditional dictations split at the condition: queue the unconditional prefix, give the probe step a short timeout, resolve the condition with a scan once the call returns.

Rules: NEVER touch the device with adb directly — no input tap, swipe, text or keyevent, no screencap, no other adb command. Every interaction is a step through queue_steps or run_route, every look is scan or capture.`

// Server ...
type Server struct {
	api  *apiClient
	mcp  *mcp.Server
	mode agentMode
}

// New ...
func New(baseURL string) *Server {
	var server = &Server{
		api: newAPIClient(baseURL),
	}
	server.mcp = mcp.NewServer(
		&mcp.Implementation{Name: serverName, Version: serverVersion},
		&mcp.ServerOptions{Instructions: serverInstructions},
	)
	server.registerTools()
	return server
}

// Run ...
func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

type emptyInput struct{}

type serialInput struct {
	Serial string `json:"serial" jsonschema:"device serial from list_devices"`
}

type waitInput struct {
	Serial         string `json:"serial" jsonschema:"device serial from list_devices"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"max seconds to wait, default 60"`
}

type landmarkInput struct {
	Type   string `json:"type" jsonschema:"image | text | yolo"`
	Value  string `json:"value" jsonschema:"library image name, text to find, or yolo class; the text to type for type_text"`
	Locale string `json:"locale,omitempty" jsonschema:"Tesseract lang code of value, default eng"`
}

type stepInput struct {
	Event     string          `json:"event,omitempty" jsonschema:"tap, long_tap, type_text, swipe_up/down/left/right, a library action name, or empty for a visibility check"`
	Landmarks []landmarkInput `json:"landmarks,omitempty" jsonschema:"target chain, the event applies to the last; empty to replay an action verbatim"`
	Timeout   *int            `json:"timeout,omitempty" jsonschema:"ms, default 5000"`
	Delay     *int            `json:"delay,omitempty" jsonschema:"ms, default 0 on the first step, 1000 after"`
}

type queueStepsInput struct {
	Serial     string      `json:"serial" jsonschema:"device serial from list_devices"`
	Steps      []stepInput `json:"steps" jsonschema:"steps in execution order"`
	KeepChecks bool        `json:"keep_checks,omitempty" jsonschema:"explorer: record this batch's visibility checks too, only checks the user asked for"`
}

type scanInput struct {
	Serial string   `json:"serial" jsonschema:"device serial from list_devices"`
	Images []string `json:"images,omitempty" jsonschema:"library image names to look for; omit for text and yolo only"`
	Locale string   `json:"locale,omitempty" jsonschema:"Tesseract lang code for the OCR, default eng"`
}

type saveImageInput struct {
	Serial string `json:"serial" jsonschema:"device serial from list_devices"`
	Name   string `json:"name" jsonschema:"<app>_<screen>_<what>[_variant]; overwrites"`
	Left   int    `json:"left" jsonschema:"rectangle edges in the pixels reported by capture"`
	Top    int    `json:"top"`
	Right  int    `json:"right"`
	Bottom int    `json:"bottom"`
}

type saveRouteInput struct {
	Name  string      `json:"name" jsonschema:"<app>_<flow>; overwrites"`
	Steps []stepInput `json:"steps" jsonschema:"the steps that ran to success, in order"`
}

type routeNameInput struct {
	Name string `json:"name" jsonschema:"route name from get_routes"`
}

type runRouteInput struct {
	Serial  string `json:"serial" jsonschema:"device serial from list_devices"`
	Name    string `json:"name" jsonschema:"route name from get_routes"`
	StartID int    `json:"start_id,omitempty" jsonschema:"step id to start from (1..N, see get_route); omit for the whole route"`
}

type recordActionInput struct {
	Serial string `json:"serial" jsonschema:"device serial from list_devices"`
	Name   string `json:"name" jsonschema:"<app>_<screen>_<what>[_variant]; overwrites"`
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "ping",
		Description: "Check that the vdroid server is reachable, starting it when needed " +
			"(up to 15 seconds). Call first in a session.",
	}, s.handlePing)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_devices",
		Description: "List connected Android devices with their serials.",
	}, s.handleListDevices)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_library",
		Description: "List library images (template crops for image landmarks) and " +
			"actions (recorded gestures usable as a step's event).",
	}, s.handleGetLibrary)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "scan",
		Description: "The primary way to observe the screen: OCR text in locale, yolo " +
			"detections and matches for the library images in images, as a table of " +
			"landmarks step chains consume. Opens a session automatically.",
	}, s.handleScan)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "capture",
		Description: "The current frame as a JPEG (about 1500 tokens). Vision only: " +
			"never call it if you cannot interpret images. Scan first; capture only " +
			"for a target the scan could not name. Opens a session automatically.",
	}, s.handleCapture)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "save_image",
		Description: "Crop a rectangle of the current frame (pixels as reported by " +
			"capture) into library image name, usable at once as an image landmark. " +
			"Only for a target with no readable text and no yolo class.",
	}, s.handleSaveImage)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "queue_steps",
		Description: "Run steps in order (a session opens automatically) and block " +
			"until the batch is done: 'idle' or the failed step's error. The whole " +
			"sequence goes into ONE call.",
	}, s.handleQueueSteps)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "close_session",
		Description: "Close the device session and stop screen capture when the flow is finished.",
	}, s.handleCloseSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "stop_server",
		Description: "Stop the local vdroid server process, closing every device " +
			"session. Only on the user's explicit ask; the next tool call starts it again.",
	}, s.handleStopServer)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_session_status",
		Description: "Session status: 'closed', 'idle', 'running <step>', or the " +
			"failed step's error text.",
	}, s.handleGetSessionStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "wait_for_session",
		Description: "Block until the session stops running steps and return the " +
			"final status. Only after queue_steps or run_route reported still " +
			"running, or for a run started by the Android client; 'closed' means the " +
			"video stream ended, just queue again.",
	}, s.handleWaitForSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_routes",
		Description: "List saved route names.",
	}, s.handleGetRoutes)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_route",
		Description: "Get a route's steps with their ids, to extend it or rerun part of it.",
	}, s.handleGetRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "save_route",
		Description: "Write or overwrite a route from steps without running them. " +
			"Only when the user asks; explorer mode records a route while running.",
	}, s.handleSaveRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "edit_route",
		Description: "Change one step of a saved route (timeout, delay, landmarks) or " +
			"delete it; later ids shift down after a delete. Saves at once.",
	}, s.handleEditRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "set_mode",
		Description: "Switch the agent mode: default, explorer (records every batch into " +
			"route) or navigator (saved routes only). Only when the user asks for a " +
			"mode; it holds until the next set_mode.",
	}, s.handleSetMode)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "delete_route",
		Description: "Delete a saved route by name.",
	}, s.handleDeleteRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "run_route",
		Description: "Run a saved route (a session opens automatically), optionally " +
			"from start_id, and block until it finishes: 'idle' or the failed " +
			"step's error.",
	}, s.handleRunRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "record_action",
		Description: "Record a gesture the HUMAN performs on the device during the 5 " +
			"seconds after the call and save it as library action name. Only when " +
			"the user asks; tell them to perform it right away. Fails while the " +
			"device is running steps.",
	}, s.handleRecordAction)
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func (s *Server) handlePing(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in emptyInput,
) (*mcp.CallToolResult, any, error) {
	err := s.api.pingServer()
	if err != nil {
		return nil, nil, err
	}
	return textResult("vdroid server is up at " + s.api.baseURL), nil, nil
}

func (s *Server) handleStopServer(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in emptyInput,
) (*mcp.CallToolResult, any, error) {
	stopped, err := s.api.stopServer()
	if err != nil {
		return nil, nil, err
	}
	if !stopped {
		return textResult("vdroid server is not running"), nil, nil
	}
	return textResult("vdroid server stopped"), nil, nil
}

func (s *Server) handleListDevices(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in emptyInput,
) (*mcp.CallToolResult, any, error) {
	devices, err := s.api.getDevices()
	if err != nil {
		return nil, nil, err
	}
	return textResult(formatDevices(devices)), nil, nil
}

func formatDevices(devices []deviceInfo) string {
	if len(devices) == 0 {
		return "no devices connected"
	}

	var lines strings.Builder
	fmt.Fprintf(&lines, "%d devices; columns: serial model, brand device, Android version, locale", len(devices))
	for _, device := range devices {
		fmt.Fprintf(
			&lines,
			"\n%s %s, %s %s, Android %s, locale %s",
			device.Serial, deviceName(device), device.Brand, device.Device, device.OsVersion, device.Locale,
		)
		if device.ScrcpyRunning {
			lines.WriteString(", session open")
		}
	}
	return lines.String()
}

func deviceName(device deviceInfo) string {
	if device.MarketingName != "" {
		return device.MarketingName
	}
	return device.Model
}

func (s *Server) handleGetLibrary(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in emptyInput,
) (*mcp.CallToolResult, any, error) {
	library, err := s.api.getLibrary()
	if err != nil {
		return nil, nil, err
	}
	var text = formatNames("images", library.Images) + "\n" + formatNames("actions", library.Actions)
	return textResult(text), nil, nil
}

func formatNames(kind string, names []string) string {
	if len(names) == 0 {
		return kind + ": none"
	}
	return fmt.Sprintf("%s (%d): %s", kind, len(names), strings.Join(names, ", "))
}

func (s *Server) handleScan(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in scanInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" {
		return nil, nil, fmt.Errorf("serial is required")
	}
	landmarks, err := s.api.scan(in.Serial, in.Images, in.Locale)
	if err != nil {
		return nil, nil, err
	}
	return textResult(formatLandmarks(landmarks, in.Locale)), nil, nil
}

func formatLandmarks(landmarks []scanLandmark, requestedLocale string) string {
	locale := textLocale(landmarks, requestedLocale)
	kept, dropped := confidentLandmarks(landmarks)
	if len(kept) == 0 {
		return fmt.Sprintf("no landmarks found (text locale %s%s)", locale, droppedNote(dropped))
	}

	var lines strings.Builder
	fmt.Fprintf(
		&lines,
		"%d landmarks in reading order, text locale %s%s; columns: type left,top,right,bottom value",
		len(kept), locale, droppedNote(dropped),
	)
	for _, landmark := range kept {
		rect := landmark.Rectangle
		fmt.Fprintf(
			&lines,
			"\n%s %d,%d,%d,%d %s",
			landmark.Type, rect.LeftX, rect.TopY, rect.RightX, rect.BottomY, landmark.Value,
		)
	}
	return lines.String()
}

func confidentLandmarks(landmarks []scanLandmark) ([]scanLandmark, int) {
	kept := make([]scanLandmark, 0, len(landmarks))
	dropped := 0
	for _, landmark := range landmarks {
		if lowConfidence(landmark) {
			dropped++
			continue
		}
		kept = append(kept, landmark)
	}
	return kept, dropped
}

func lowConfidence(landmark scanLandmark) bool {
	if landmark.Type != "text" || landmark.Confidence == nil {
		return false
	}
	return *landmark.Confidence < scanMinConfidence
}

func droppedNote(dropped int) string {
	if dropped == 0 {
		return ""
	}
	return fmt.Sprintf(", %d text entries below confidence %d dropped", dropped, scanMinConfidence)
}

func textLocale(landmarks []scanLandmark, requestedLocale string) string {
	for _, landmark := range landmarks {
		if landmark.Locale != "" {
			return landmark.Locale
		}
	}
	if requestedLocale != "" {
		return requestedLocale
	}
	return "eng"
}

func (s *Server) handleCapture(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in serialInput,
) (*mcp.CallToolResult, any, error) {
	if err := s.refuseInNavigator("capture"); err != nil {
		return nil, nil, err
	}
	if in.Serial == "" {
		return nil, nil, fmt.Errorf("serial is required")
	}
	data, err := s.api.capture(in.Serial)
	if err != nil {
		return nil, nil, err
	}
	config, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	var text = fmt.Sprintf("frame %dx%d pixels, same coordinates as scan", config.Width, config.Height)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
			&mcp.ImageContent{Data: data, MIMEType: captureMimeType},
		},
	}, nil, nil
}

func (s *Server) handleSaveImage(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in saveImageInput,
) (*mcp.CallToolResult, any, error) {
	if err := s.refuseInNavigator("save_image"); err != nil {
		return nil, nil, err
	}
	if in.Serial == "" || in.Name == "" {
		return nil, nil, fmt.Errorf("serial and name are required")
	}
	if in.Right <= in.Left || in.Bottom <= in.Top {
		return nil, nil, fmt.Errorf("rectangle must have right > left and bottom > top")
	}
	rectangle := saveImageRectangle{
		LeftX:   in.Left,
		RightX:  in.Right,
		TopY:    in.Top,
		BottomY: in.Bottom,
		Label:   in.Name,
	}
	err := s.api.saveImage(in.Serial, rectangle)
	if err != nil {
		return nil, nil, err
	}
	return textResult("saved image " + in.Name), nil, nil
}

func (s *Server) handleQueueSteps(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in queueStepsInput,
) (*mcp.CallToolResult, any, error) {
	mode, route := s.mode.get()
	if mode == modeNavigator {
		return nil, nil, navigatorRefusal("queue_steps")
	}
	if in.Serial == "" || len(in.Steps) == 0 {
		return nil, nil, fmt.Errorf("serial and steps are required")
	}
	for _, step := range in.Steps {
		if step.Event == "" && len(step.Landmarks) == 0 {
			return nil, nil, fmt.Errorf("every step needs an event or a target")
		}
	}

	recorded := slices.Clone(in.Steps)
	err := s.api.queueSteps(in.Serial, in.Steps)
	if err != nil {
		return nil, nil, err
	}

	status, finished, err := s.waitUntilNotRunning(ctx, in.Serial, queueWaitSec*time.Second)
	if err != nil {
		return nil, nil, err
	}
	var outcome = batchOutcome(fmt.Sprintf("queued %d steps", len(in.Steps)), status, finished)
	if mode != modeExplorer {
		return textResult(outcome), nil, nil
	}
	var succeeded = succeededSteps(recorded, status, finished)
	if !in.KeepChecks {
		succeeded = withoutChecks(succeeded)
	}
	return textResult(outcome + "; " + s.recordRoute(route, succeeded)), nil, nil
}

func succeededSteps(steps []stepInput, status string, finished bool) []stepInput {
	if !finished {
		return nil
	}
	if status == idleStatus {
		return steps
	}

	var failedID int
	_, err := fmt.Sscanf(status, "step %d:", &failedID)
	if err != nil || failedID < 1 || failedID > len(steps) {
		return nil
	}
	return steps[:failedID-1]
}

func (s *Server) recordRoute(name string, succeeded []stepInput) string {
	if len(succeeded) == 0 {
		return fmt.Sprintf("route %s unchanged", name)
	}
	total, err := s.api.appendToRoute(name, succeeded)
	if err != nil {
		return fmt.Sprintf("route %s not saved: %v", name, err)
	}
	return fmt.Sprintf("route %s +%d = %d steps", name, len(succeeded), total)
}

func (s *Server) handleCloseSession(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in serialInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" {
		return nil, nil, fmt.Errorf("serial is required")
	}
	err := s.api.closeSession(in.Serial)
	if err != nil {
		return nil, nil, err
	}
	return textResult("session closed"), nil, nil
}

func (s *Server) handleGetSessionStatus(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in serialInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" {
		return nil, nil, fmt.Errorf("serial is required")
	}
	status, err := s.api.getSessionStatus(in.Serial)
	if err != nil {
		return nil, nil, err
	}
	return textResult(status), nil, nil
}

func (s *Server) handleWaitForSession(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in waitInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" {
		return nil, nil, fmt.Errorf("serial is required")
	}

	var timeoutSec = in.TimeoutSeconds
	if timeoutSec <= 0 {
		timeoutSec = defaultWaitSec
	}

	status, finished, err := s.waitUntilNotRunning(ctx, in.Serial, time.Duration(timeoutSec)*time.Second)
	if err != nil {
		return nil, nil, err
	}
	if !finished {
		return textResult(fmt.Sprintf("timeout after %ds, last status: %s", timeoutSec, status)), nil, nil
	}
	return textResult(status), nil, nil
}

func (s *Server) waitUntilNotRunning(
	ctx context.Context,
	serial string,
	timeout time.Duration,
) (string, bool, error) {
	var deadline = time.Now().Add(timeout)
	var status string
	for time.Now().Before(deadline) {
		var err error
		status, err = s.api.getSessionStatus(serial)
		if err != nil {
			return "", false, err
		}
		if !strings.HasPrefix(status, runningPrefix) {
			return status, true, nil
		}
		if err := sleepCtx(ctx, pollInterval); err != nil {
			return "", false, err
		}
	}
	return status, false, nil
}

func batchOutcome(prefix string, status string, finished bool) string {
	if finished {
		return prefix + " -> " + status
	}
	return fmt.Sprintf(
		"%s, still running after %ds (last status: %s) - call wait_for_session",
		prefix, queueWaitSec, status,
	)
}

func (s *Server) handleGetRoutes(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in emptyInput,
) (*mcp.CallToolResult, any, error) {
	routes, err := s.api.getRoutes()
	if err != nil {
		return nil, nil, err
	}
	return textResult(formatNames("routes", routes)), nil, nil
}

func (s *Server) handleGetRoute(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in routeNameInput,
) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	route, err := s.api.getRoute(in.Name)
	if err != nil {
		return nil, nil, err
	}
	return textResult(formatRoute(route)), nil, nil
}

func formatRoute(route routeResponse) string {
	var lines strings.Builder
	fmt.Fprintf(
		&lines,
		"route %s, %d steps: id event timeout/delay landmarks (check = empty event)",
		route.Name, len(route.Steps),
	)
	for _, step := range route.Steps {
		fmt.Fprintf(&lines, "\n%d %s %d/%d", step.ID, stepEvent(step.Event), step.Timeout, step.Delay)
		if len(step.Landmarks) == 0 {
			continue
		}
		lines.WriteString(" " + formatChain(step.Landmarks))
	}
	return lines.String()
}

func stepEvent(event string) string {
	if event == "" {
		return "check"
	}
	return event
}

func formatChain(landmarks []landmarkInput) string {
	parts := make([]string, 0, len(landmarks))
	for _, landmark := range landmarks {
		part := fmt.Sprintf("%s %q", landmark.Type, landmark.Value)
		if landmark.Locale != "" {
			part += " " + landmark.Locale
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, " > ")
}

func (s *Server) handleSaveRoute(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in saveRouteInput,
) (*mcp.CallToolResult, any, error) {
	if err := s.refuseInNavigator("save_route"); err != nil {
		return nil, nil, err
	}
	if in.Name == "" || len(in.Steps) == 0 {
		return nil, nil, fmt.Errorf("name and steps are required")
	}
	err := s.api.saveRoute(&in)
	if err != nil {
		return nil, nil, err
	}
	return textResult("saved route " + in.Name), nil, nil
}

func (s *Server) handleDeleteRoute(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in routeNameInput,
) (*mcp.CallToolResult, any, error) {
	if err := s.refuseInNavigator("delete_route"); err != nil {
		return nil, nil, err
	}
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	err := s.api.deleteRoute(in.Name)
	if err != nil {
		return nil, nil, err
	}
	return textResult("deleted route " + in.Name), nil, nil
}

func (s *Server) handleRunRoute(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in runRouteInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" || in.Name == "" {
		return nil, nil, fmt.Errorf("serial and name are required")
	}
	err := s.api.runRoute(in.Serial, in.Name, in.StartID)
	if err != nil {
		return nil, nil, err
	}
	status, finished, err := s.waitUntilNotRunning(ctx, in.Serial, queueWaitSec*time.Second)
	if err != nil {
		return nil, nil, err
	}
	return textResult(batchOutcome("route "+in.Name, status, finished)), nil, nil
}

func (s *Server) handleRecordAction(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in recordActionInput,
) (*mcp.CallToolResult, any, error) {
	if err := s.refuseInNavigator("record_action"); err != nil {
		return nil, nil, err
	}
	if in.Serial == "" || in.Name == "" {
		return nil, nil, fmt.Errorf("serial and name are required")
	}
	recorded, err := s.api.recordAction(in.Serial, in.Name)
	if err != nil {
		return nil, nil, err
	}
	if !recorded {
		var text = "nothing recorded: no touch happened on the device during the " +
			"5 second window; ask the user to perform the gesture again"
		return textResult(text), nil, nil
	}
	return textResult("saved action " + in.Name), nil, nil
}

func sleepCtx(ctx context.Context, duration time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}
