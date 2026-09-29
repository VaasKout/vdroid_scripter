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

Workflow: ping first, list_devices for a serial, get_routes for saved flows, then act; sessions open automatically. get_library only when an icon target or a recorded gesture might be needed: images are template crops used as image landmarks, actions are recorded gestures used as events. Library names carry their context as <app>_<screen>_<what>[_variant] (shop_catalog_swipe_1 = a swipe recorded on a shop app's catalog screen); a duplicate name overwrites. Text is free: text landmarks and the generated events need nothing from the library, so an instruction in words visible on screen ("open Settings") is tap steps on text landmarks through the obvious screens (Settings -> Network & internet -> Wi-Fi).

Steps: a step is an event applied to a chain of landmarks. A landmark is a CV target: type image | text | yolo, value = library image name | text to find | yolo class. The chain resolves on ONE video frame: the first landmark takes its first match in reading order (top to bottom, left to right), each following landmark the candidate NEAREST to the previous one, and the event applies to the LAST. One landmark is the normal case; to disambiguate duplicates put a unique nearby element first ("the toggle next to 'Show refresh rate'" = [{text "Show refresh rate"}, {image toggle}]). Events: tap and long_tap touch the last landmark. swipe_up/down/left/right make a human-like fixed-length swipe in the finger's direction (swipe_up reveals content below) from a random point, or from inside the last landmark when given — use them for scrolling; recorded library swipes are for app-specific gestures. type_text types the last landmark's value on the keyboard already open (tap the field in an earlier step); its locale is REQUIRED and must be the keyboard's language; letters, space, digits (number row) and capitals are typed, punctuation and symbols fail the step; a number or phone field takes locale numeric (keypad, digits only). An EMPTY event is a visibility check. Any other event replays that library action, moved into the last landmark when given, verbatim without.

Locale: text landmarks and type_text carry the Tesseract code of their value's language, eng by default; pass the text exactly as the user wrote it, never transliterate or translate.

Timing: delay is ms slept BEFORE the step acts, letting the previous change settle; timeout is ms the server keeps re-locating the target on live frames, proceeding the moment it appears; both are literal, 0 = no delay / one look. Omitted: delay 0 on a batch's first step and 1000 after, timeout 5000. For a target that appears late (app launch, navigation, loading) raise timeout to 10000-15000, not delay; a bigger delay only for a target visible early but not yet safe to touch. A probe never gets a raised timeout, so a negative answer comes back fast. Session startup never eats the timeout.

Batching: a dictated sequence becomes ONE queue_steps call in the given order — never one step at a time, never status polls in between. The call returns when the batch is done: 'idle' = every step succeeded; otherwise the error names the failed step by id (its batch position, or the route's id) and the rest of the queue is cleared. Only a batch reported still running needs wait_for_session.

Literal execution: when the user names concrete actions ("tap Settings, then tap Wi-Fi"), queue exactly those — no extra checks or probes, nothing added, substituted or reordered; a repeated ask runs again every time, as many times as asked. Never argue or ask for confirmation. Improvise only when a step fails.

Perception: scan is the primary way to see the screen, capture the second. Scan when a step failed, for a conditional instruction ("if X is not visible, ..."), when the user asks what is on screen, or on each new screen of an abstract task — never habitually between dictated steps; pass in images the library names related to the current app. The result is a header (count, resolved text locale, dropped low-confidence entries), then one "type left,top,right,bottom value" line per landmark in reading order; type/value are what step landmarks consume (with the header's locale on text), coordinates only show what sits next to what. A missing expected word was misread or is in another language: rescan with the matching locale before calling it absent. capture costs about 1500 tokens — five scans — per call and stays in context; never call it if you cannot interpret images. With vision, scan still comes first and names the landmark values: if the target's text or yolo class is in the scan, act without capturing. Capture only for a target the scan lacks that has no text (an icon, a picture, a toggle's colour) or for an unknown screen the scan says nothing about, naming the target — at most once per new screen, never within a dictated batch.

Curation: a target with no readable text and no yolo class needs a library image: use one from get_library, otherwise with vision capture, pick the icon's tight rectangle in the reported pixels (only the icon, no badge or highlight that changes) and save_image it as <app>_<screen>_<what>[_variant] — on your own initiative, never asking the user to curate; saved images make the next run capture-free. Ask the user for a library item only when you have no vision and no text, yolo class or generated swipe reaches the target. Actions come from the Android client, or from record_action on the user's ask.

Abstract tasks (explorer only): when the user states a goal ("write John a message in Telegram", "turn off Wi-Fi"), derive the steps yourself and carry it through without asking, screen by screen: scan, turn text and yolo classes into landmarks, queue what you are sure of, read the outcome, look again, until done, then report; type message text with type_text in its own locale.

Routes: a route is a saved flow — a name and the steps that ran to success, with ids 1..N and their delay/timeout. run_route returns its outcome like queue_steps; start_id runs from that step — on the user's ask, or to rerun the unchanged remainder after a recovered failure; the start step gets delay 0, an unknown id runs nothing. Never create or modify a route unless the user asks, apart from explorer recording and navigator route repair.

Modes: navigator (the start mode) and explorer; the mode holds until the user changes it (see Rules). In both, dictated steps run as asked (Literal execution), with recovery on failure.
navigator records nothing, and dictated steps never change a route. A goal is reached only across saved routes: get_routes, get_route and scan show where the phone is and which route leads on, or which step to enter via start_id; chain run_route calls; a goal no route reaches needs explorer: say so. When a route run fails, scan; for a route defect, fix the step with edit_route and rerun from it without asking: the target is on screen now -> raise its timeout; it appears in several places or the previous tap hit the wrong one -> landmarks with a unique neighbour first; the screen is mid-transition -> raise its delay; the step repeats the one before and undoes it -> delete it. At most two fixes per failing step; anything else (another screen, a popup, a missing app) is no defect: report it. List every fix in your answer.
explorer (with a route name) carries a goal through as an abstract task and records your own work: each batch of yours appends its succeeded steps, minus your visibility checks and probes, to that route and saves it — no save_route needed; a failed step and the rest of its batch stay out, the recovery batch goes in; an existing route is appended to (delete_route first only when the user wants it replaced). Steps the user dictates go in their own batch with dictated=true: they run as asked and only their visibility checks, the checks the user asked for, are recorded.

Failure and recovery: recover from the failed step: scan (with the relevant library images) and apply the user's instruction to what it shows — tap the alternative the user named, scroll with a generated swipe or the screen's recorded swipe (variants _1, _2, ...) when the target should be further down, or report honestly on an unexpected screen — then re-queue from the failed step in one call, or run_route with start_id when a route's remaining steps need no change. A scan with no landmarks at all, not even the status bar, means the screen is off: close_session, scan again (the new session turns the screen on) and continue. Conditional dictations split at the condition: queue the unconditional prefix ending with the probe step, then resolve the condition with a scan once the call returns.

Rules: NEVER touch the device with adb — no input, keyevent, screencap or any other adb command; every interaction is a step through queue_steps or run_route, every look is scan or capture. NEVER call set_mode unless the user explicitly tells you to change the mode — not because a task seems to need another mode, not to finish a task; when another mode is needed, say so and wait.`

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
	Serial string `json:"serial" jsonschema:"device serial"`
}

type waitInput struct {
	Serial         string `json:"serial" jsonschema:"device serial"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"max seconds, default 60"`
}

type landmarkInput struct {
	Type   string `json:"type" jsonschema:"image | text | yolo"`
	Value  string `json:"value" jsonschema:"image name, text or yolo class; for type_text the text to type"`
	Locale string `json:"locale,omitempty" jsonschema:"Tesseract code of value, default eng"`
}

type stepInput struct {
	Event     string          `json:"event,omitempty" jsonschema:"tap, long_tap, type_text, swipe_up/down/left/right, a library action, or empty = visibility check"`
	Landmarks []landmarkInput `json:"landmarks,omitempty" jsonschema:"target chain, event on the last; empty = replay verbatim"`
	Timeout   *int            `json:"timeout,omitempty" jsonschema:"ms"`
	Delay     *int            `json:"delay,omitempty" jsonschema:"ms"`
}

type queueStepsInput struct {
	Serial   string      `json:"serial" jsonschema:"device serial"`
	Steps    []stepInput `json:"steps" jsonschema:"in execution order"`
	Dictated bool        `json:"dictated,omitempty" jsonschema:"explorer: the batch is exactly the user's dictated steps"`
}

type scanInput struct {
	Serial string   `json:"serial" jsonschema:"device serial"`
	Images []string `json:"images,omitempty" jsonschema:"library image names to look for"`
	Locale string   `json:"locale,omitempty" jsonschema:"Tesseract code for the OCR, default eng"`
}

type saveImageInput struct {
	Serial string `json:"serial" jsonschema:"device serial"`
	Name   string `json:"name" jsonschema:"<app>_<screen>_<what>[_variant]"`
	Left   int    `json:"left" jsonschema:"edges in capture pixels"`
	Top    int    `json:"top"`
	Right  int    `json:"right"`
	Bottom int    `json:"bottom"`
}

type saveRouteInput struct {
	Name  string      `json:"name" jsonschema:"<app>_<flow>"`
	Steps []stepInput `json:"steps" jsonschema:"the steps that ran to success, in order"`
}

type routeNameInput struct {
	Name string `json:"name" jsonschema:"route name"`
}

type runRouteInput struct {
	Serial  string `json:"serial" jsonschema:"device serial"`
	Name    string `json:"name" jsonschema:"route name"`
	StartID int    `json:"start_id,omitempty" jsonschema:"step id to start from; omit for the whole route"`
}

type recordActionInput struct {
	Serial string `json:"serial" jsonschema:"device serial"`
	Name   string `json:"name" jsonschema:"<app>_<screen>_<what>[_variant]"`
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "ping",
		Description: "Check that the vdroid server is reachable, starting it when needed (up to 15 s).",
	}, s.handlePing)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_devices",
		Description: "List connected Android devices with their serials.",
	}, s.handleListDevices)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_library",
		Description: "List library image and action names.",
	}, s.handleGetLibrary)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "scan",
		Description: "Observe the screen: OCR text in locale, yolo detections and matches " +
			"for the library images in images.",
	}, s.handleScan)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "capture",
		Description: "The current frame as a JPEG (about 1500 tokens). Vision only, and " +
			"only for a target the scan could not name.",
	}, s.handleCapture)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "save_image",
		Description: "Crop a rectangle of the current frame (capture pixels) into library " +
			"image name. Only for a target without text or yolo class.",
	}, s.handleSaveImage)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "queue_steps",
		Description: "Run steps in order and block until the batch is done: 'idle' or " +
			"the failed step's error.",
	}, s.handleQueueSteps)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "close_session",
		Description: "Close the device session when the flow is finished.",
	}, s.handleCloseSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "stop_server",
		Description: "Stop the local vdroid server process, closing every session. Only " +
			"on the user's explicit ask; the next call starts it again.",
	}, s.handleStopServer)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_session_status",
		Description: "Session status: 'closed', 'idle', 'running <step>', or the " +
			"failed step's error text.",
	}, s.handleGetSessionStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "wait_for_session",
		Description: "Block until the session stops running steps and return the final " +
			"status; 'closed' means the video stream ended, just queue again. Only for a " +
			"batch reported still running or a run started by the Android client.",
	}, s.handleWaitForSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_routes",
		Description: "List saved route names.",
	}, s.handleGetRoutes)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_route",
		Description: "A route's steps with their ids.",
	}, s.handleGetRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "save_route",
		Description: "Write or overwrite a route from steps without running them. Only " +
			"when the user asks.",
	}, s.handleSaveRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "edit_route",
		Description: "Change one step of a saved route (timeout, delay, landmarks) or " +
			"delete it; later ids shift down after a delete.",
	}, s.handleEditRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "set_mode",
		Description: "Switch the agent mode: navigator (the start mode, nothing recorded) " +
			"or explorer (records your own batches into route). ONLY when the user " +
			"explicitly tells you to change the mode, never on your own.",
	}, s.handleSetMode)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "delete_route",
		Description: "Delete a saved route by name.",
	}, s.handleDeleteRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "run_route",
		Description: "Run a saved route, optionally from start_id, and block until it " +
			"finishes: 'idle' or the failed step's error.",
	}, s.handleRunRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "record_action",
		Description: "Record a gesture the HUMAN performs on the device within 5 seconds " +
			"of the call and save it as library action name; 'nothing recorded' = no " +
			"touch, ask them to retry. Only when the user asks; tell them to perform it " +
			"right away.",
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
	var recordable = filterChecks(succeededSteps(recorded, status, finished), in.Dictated)
	return textResult(outcome + "; " + s.recordRoute(route, recordable)), nil, nil
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
