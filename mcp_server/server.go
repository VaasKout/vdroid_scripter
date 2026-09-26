package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName     = "vdroid-scripter"
	serverVersion  = "0.1.0"
	runningPrefix  = "running"
	pollInterval   = time.Second
	defaultWaitSec = 60

	defaultTimeoutMs = 5000
	defaultDelayMs   = 1000
	firstStepDelayMs = 0

	scanMinConfidence = 40
)

const serverInstructions = `vdroid-scripter drives Android devices with CV-located steps composed from a human-curated library.

Workflow: ping first (it starts the vdroid server when needed and waits for it), then list_devices for a serial, get_routes for saved flows, then queue steps and wait for the outcome. Call get_library only when image targets or recorded gestures might be needed — it lists the available images (template crops) and actions (recorded gestures); library names carry their context as <app>_<screen>_<what>[_variant] — e.g. shop_catalog_swipe_1 is a swipe recorded on a shop app's catalog screen, first variant. stop_server shuts the whole vdroid server process down (every device session closes) — only when the user explicitly asks to stop or restart it; the next tool call starts it again on demand.

Text is free: text landmarks and the generated events (tap, long_tap, the swipes, type_text) need NOTHING from the library. An instruction phrased in words visible on screen ("open Settings", "enter wifi connections") is just tap steps with text landmarks — tap the matching words, drilling through the obvious screens (e.g. Settings -> Network & internet -> Wi-Fi). Only reach for the library when the target has no readable text (an icon = image landmark) or needs a recorded gesture.

Batching: when given a sequence of steps ("tap text1, tap yolo class home, swipe, type hi..."), translate the WHOLE sequence into ONE queue_steps call with the steps in the given order. Never queue one step at a time and never poll get_session_status between steps — the server executes the queue sequentially on its own. After the single call, call wait_for_session once: 'idle' means every step succeeded.

Steps: a step is an event applied to a chain of landmarks. Each landmark is a CV target (type = image | text | yolo, value = library image name / text to find / yolo class; text landmarks also carry locale). The chain resolves on ONE video frame: the first landmark picks its best match on screen, every following landmark picks the candidate of its value NEAREST to the previous landmark, and the event applies to the LAST landmark. One landmark is the normal case ("tap the cart icon" = one image landmark). Put a nearby unique element first to disambiguate duplicates: "tap the toggle next to 'Show refresh rate'" = landmarks [{type text, value Show refresh rate}, {type image, value toggle}]. Events tap and long_tap touch the last landmark's region. swipe_up, swipe_down, swipe_left and swipe_right generate a human-like fixed-length swipe (direction = the finger's movement, so swipe_up reveals content below): without landmarks it starts at a random point on screen, with landmarks it starts inside the last landmark's region — use these for plain scrolling; recorded library swipes remain for app-specific gestures. type_text types the LAST landmark's value on the on-screen keyboard, which must already be open (tap the field first, in an earlier step); that landmark's locale is REQUIRED and names the keyboard language, and the keyboard has to be showing that language. Letters, space and digits (when the keyboard has a number row) are typed, capitals through Shift; punctuation and symbols are not supported and fail the step. For a number or phone field use locale numeric: the server then expects the numeric keypad and types digits only. An EMPTY event is a pure visibility check of the landmark chain. Any other event name replays that recorded library event: offset into the last landmark's region when landmarks are given (e.g. a drag starting from an icon), verbatim without landmarks.

Timing: every step carries two millisecond knobs, delay and timeout, and the server takes them literally — this MCP fills in whatever you omit. delay is slept BEFORE the step acts — it paces the flow and gives the previous action's screen change time to settle; omitted, it is 0 on the first step of a batch (nothing preceded it) and 1000 on every later step; 0 means no delay. timeout is how long the server keeps re-locating the target on live frames before failing — retries run back to back, and the step proceeds the moment the target shows up; omitted, it is 5000; 0 means one look at the current frame with no waiting. For a target that appears late (app launch, a navigation tap, network loading) raise timeout (10000-15000) rather than delay; use a bigger delay only when the target is visible early but not yet safe to touch (mid-animation). Keep probe/visibility checks at the default timeout so a negative answer comes back fast. Session startup never eats the step timeout — the first video frame gets its own grace period.

Locale: for text landmarks and type_text, always set the landmark's locale to the Tesseract language code of its value's language. This holds for every language Tesseract supports; eng is the default. Pass the text exactly as the user wrote it, never transliterate or translate it.

Perception: scan is the ONLY way to observe the screen — there are no screenshots and never will be. Call scan when a step failed, when the user's instruction is conditional ("if X is not visible, ..."), or when the user explicitly asks what is on screen. Never scan habitually between steps — the happy path is one queue_steps call and one wait_for_session. Pass in images the library image names plausibly related to the current app so scan reports which of them are visible. The result is a compact table — a header with the landmark count and the resolved text locale, then one line per landmark in reading order, "type left,top,right,bottom value" — and its type/value pairs are exactly what step landmarks consume: build follow-up steps from them, with the header's locale on text landmarks, and use the coordinates only to judge which elements sit next to each other. Text the OCR read with confidence below 40 is left out and the header counts the dropped entries — when a word you expected is missing, it was misread or is in another language, so scan again with the matching locale before concluding it is not on screen.

Routes: a route is a saved flow — a name, the user's dictation as its prompt, and the exact steps that ran to success. When the user asks to save or remember a flow as <name>, call save_route with the name, the user's dictation VERBATIM as the prompt (conditions included), and the steps that actually succeeded in order; a duplicate name overwrites. The prompt may be absent on routes saved elsewhere (the Android client saves routes without one) — treat such a route as a plain script with no recorded intent. To run a saved route: run_route, then wait_for_session once — 'idle' means the whole route succeeded. Saving stamps every step with an id (1..N, its position) and stores delay and timeout exactly as sent (this MCP fills omitted ones the same way as for queue_steps, so a route's first step gets delay 0); run_route accepts an optional start_id to start mid-route from that step id — use it when the user asks to run a route from a specific point, or to rerun the unchanged remainder after a recovered failure; an id the route does not contain is an error and nothing runs, and whichever step a run starts from gets delay 0 regardless of the stored value. To extend a route: get_route, append the new steps, call save_route with the full list — nothing executes. If a route step fails, the error status names the failed step's id (queue_steps batches get 1-based position ids the same way); recover from that point guided by the route's prompt: scan, decide, then either queue adjusted steps with queue_steps or, when the remaining steps need no changes, run_route with start_id of the failed step — and after a recovered run ask the user whether to update the route with the steps that worked. Never create or modify routes without being asked.

Curation: library images are created by the human with the Android client — there is no tool here to create them. Library actions can be recorded from here: when the user asks to record a gesture as <name>, call record_action and tell them to perform the gesture on the device immediately — the device listens for 5 seconds from the moment of the call and the call blocks until the window ends; 'nothing recorded' means no touch happened, ask them to try again. Never call record_action on your own initiative. Ask the user to add a library item ONLY when the target truly cannot be reached any other way — no readable text for a text landmark, no yolo class, no generated swipe that gets there. Never request curation for something written on the screen.

Rules: NEVER drive the device with adb directly — no adb shell input tap, input swipe, input text, keyevent, or any other adb command, no matter what. Every interaction is a step executed through queue_steps: tap/long_tap to touch a target, a library event to gesture, type_text to type, and an EMPTY event to find or verify an element. There is no separate lookup tool — finding an element and acting on it are both steps; scan exists only for the failure, conditional and what-is-on-screen cases described above.

Literal execution: when the user names a concrete action, queue exactly that action and nothing else — no extra visibility checks, no probing, no added, substituted or reordered steps, no "better" alternatives. When the user asks for the same thing repeatedly, execute it again every time, exactly as many times as asked — never skip a repeat because it was already done and never deduplicate. Never argue, never ask for confirmation — just execute. Improvise only when a step fails (see recovery below).

Duplicates: when a landmark value matches several places on screen, disambiguate with a chain — put a unique nearby landmark first; the following landmark resolves nearest to it. When no unique neighbor exists, a bare landmark deterministically takes the FIRST candidate in reading order (top to bottom, left to right).

Failure and recovery: a failed step clears the remaining queue and stores the error as the session status, so the final status names the target that could not be found. Recover from the failure point: scan the screen (with the relevant library images in the images param), apply the user's instruction or the route's prompt to what the scan shows — tap the alternative the user named, scroll with a generated swipe (swipe_up to reveal content below) or the screen's recorded swipe action (variants _1, _2, ...) when the target should be below, or report honestly when the scan shows an unexpected screen — then re-queue the remaining steps from the failed one onward, again in one call. Conditional dictations split at the condition: queue the unconditional prefix, give the probe step its own short timeout, and resolve the condition with a scan after the wait.`

// Server ...
type Server struct {
	api *apiClient
	mcp *mcp.Server
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
	Serial string `json:"serial" jsonschema:"device serial number, get it from list_devices"`
}

type waitInput struct {
	Serial         string `json:"serial" jsonschema:"device serial number"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"max seconds to wait, default 60"`
}

type landmarkInput struct {
	Type   string `json:"type" jsonschema:"image | text | yolo"`
	Value  string `json:"value" jsonschema:"library image name, text to find, or yolo class; for type_text the text to type"`
	Locale string `json:"locale,omitempty" jsonschema:"Tesseract lang code of value for text landmarks and type_text, default eng"`
}

type stepInput struct {
	Event     string          `json:"event,omitempty" jsonschema:"tap, long_tap, type_text, swipe_up/down/left/right, a library action name, or EMPTY for a visibility check"`
	Landmarks []landmarkInput `json:"landmarks,omitempty" jsonschema:"target chain: each landmark is located nearest to the previous one and the event applies to the LAST; empty to replay a library action verbatim"`
	Timeout   *int            `json:"timeout,omitempty" jsonschema:"ms to keep locating the target; default 5000, raise for targets that appear after a launch or load, 0 = one look"`
	Delay     *int            `json:"delay,omitempty" jsonschema:"ms slept before the step; default 0 on the first step of a batch, 1000 after"`
}

type queueStepsInput struct {
	Serial string      `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Steps  []stepInput `json:"steps" jsonschema:"steps to queue, executed in the given order"`
}

type scanInput struct {
	Serial string   `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Images []string `json:"images,omitempty" jsonschema:"library image names to search for on the screen; pass the images plausibly related to the current app; omit for text+yolo only"`
	Locale string   `json:"locale,omitempty" jsonschema:"Tesseract lang code for the OCR pass, default eng"`
}

type saveRouteInput struct {
	Name   string      `json:"name" jsonschema:"route name, <app>_<flow> convention; a duplicate name overwrites"`
	Prompt string      `json:"prompt,omitempty" jsonschema:"the user's original dictation of the flow, VERBATIM, conditions included; empty only when there was no dictation"`
	Steps  []stepInput `json:"steps" jsonschema:"the steps that actually ran to success, in order"`
}

type routeNameInput struct {
	Name string `json:"name" jsonschema:"route name from get_routes"`
}

type runRouteInput struct {
	Serial  string `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Name    string `json:"name" jsonschema:"route name from get_routes"`
	StartID int    `json:"start_id,omitempty" jsonschema:"id of the step to start from (route steps carry ids 1..N, see get_route); omit to run the whole route; an id the route does not contain is an error and nothing runs"`
}

type recordActionInput struct {
	Serial string `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Name   string `json:"name" jsonschema:"library action name to save the gesture under, <app>_<screen>_<what>[_variant] convention; a duplicate name overwrites"`
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "ping",
		Description: "Check that the vdroid server is reachable, starting it when it is " +
			"not (waits up to 15 seconds for it to come up). Call this first in a " +
			"session so the startup wait is visible; every other tool also starts " +
			"the server on demand.",
	}, s.handlePing)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "list_devices",
		Description: "List connected Android devices with their serial numbers. " +
			"Call this after ping to find the serial required by every other tool.",
	}, s.handleListDevices)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_library",
		Description: "List the library: 'images' are template crops usable as image " +
			"targets, 'actions' are recorded gestures usable as a step's event; names " +
			"encode their context as <app>_<screen>_<what>[_variant]. Only when an " +
			"image target or a recorded gesture might be needed.",
	}, s.handleGetLibrary)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "scan",
		Description: "The ONLY way to observe the screen (no screenshots): OCR text in " +
			"locale, yolo detections, and matches for the library images listed in " +
			"images. Only after a step failure, for a conditional instruction, or " +
			"when the user asks what is on screen — never between steps. Returns a " +
			"header (count, resolved text locale, how many text entries under OCR " +
			"confidence 40 were dropped) then one `type left,top,right,bottom value` " +
			"line per landmark in reading order; type/value are what step landmarks " +
			"consume, with the header's locale on text. A missing expected word was " +
			"misread or is in another language — rescan with the right locale. Opens " +
			"a session automatically.",
	}, s.handleScan)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "queue_steps",
		Description: "Queue steps to run in order on the device (a session opens " +
			"automatically). Put the whole sequence into ONE call and call " +
			"wait_for_session once afterwards — never one step at a time, never " +
			"status checks in between. Step and event semantics are in the server " +
			"instructions. A failed step clears the remaining queue and sets the " +
			"error status.",
	}, s.handleQueueSteps)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "close_session",
		Description: "Close the device session and stop screen capture. " +
			"Call when the flow is finished.",
	}, s.handleCloseSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "stop_server",
		Description: "Stop the local vdroid server process (SIGTERM to vdroid-scripter " +
			"on this machine; it closes every device session first, so screen " +
			"capture stops on every phone). Only when the user explicitly asks to " +
			"stop or restart the server — the next tool call starts it again on demand.",
	}, s.handleStopServer)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_session_status",
		Description: "Get the session status for a device. Values: 'closed' (no session), " +
			"'idle' (session open, step queue empty — previous steps all succeeded), " +
			"'running <step>' (a step is executing), or an error text like " +
			"'unable to find ...' meaning the last step failed and the remaining " +
			"queue was cleared.",
	}, s.handleGetSessionStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "wait_for_session",
		Description: "Block until the device session stops running queued steps, then " +
			"return the final status: 'idle' means everything completed successfully, an error " +
			"text means a step failed (remaining queue was cleared), 'closed' means the " +
			"session ended. Call this after queueing steps instead of polling manually.",
	}, s.handleWaitForSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_routes",
		Description: "List saved route names. A route is a remembered flow: the exact " +
			"steps that succeeded, with the user's original dictation as its prompt " +
			"when it was saved from one.",
	}, s.handleGetRoutes)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_route",
		Description: "Get one saved route: its steps and, when present, its prompt " +
			"(the flow's intent, conditions included). Use it to extend a route or " +
			"to recover a failed run guided by the prompt.",
	}, s.handleGetRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "save_route",
		Description: "Save or overwrite a route. Call ONLY when the user asks to " +
			"save or remember a flow. steps = the steps that actually ran to " +
			"success, in order; prompt = the user's dictation VERBATIM. Saving " +
			"executes nothing.",
	}, s.handleSaveRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "delete_route",
		Description: "Delete a saved route by name.",
	}, s.handleDeleteRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "run_route",
		Description: "Queue a saved route's steps on the device (a session opens " +
			"automatically); start_id starts mid-route from that step id (ids " +
			"1..N). Then call wait_for_session once — an error status names the " +
			"failed step id.",
	}, s.handleRunRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "record_action",
		Description: "Record a gesture the HUMAN performs on the device and save it " +
			"as library action <name>. Call ONLY when the user asks to record a " +
			"gesture. The device listens for 5 seconds from the moment of the " +
			"call, so tell the user to perform the gesture on the device right " +
			"away; the call blocks for the whole window and replies 'saved' or " +
			"'nothing recorded' (ask the user to try again). Only the first " +
			"finger is kept. Fails with 409 while the device is running steps.",
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
		if landmark.Type == "text" && landmark.Confidence < scanMinConfidence {
			dropped++
			continue
		}
		kept = append(kept, landmark)
	}
	return kept, dropped
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

func (s *Server) handleQueueSteps(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in queueStepsInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" || len(in.Steps) == 0 {
		return nil, nil, fmt.Errorf("serial and steps are required")
	}
	for _, step := range in.Steps {
		if step.Event == "" && len(step.Landmarks) == 0 {
			return nil, nil, fmt.Errorf("every step needs an event or a target")
		}
	}

	err := s.api.queueSteps(in.Serial, in.Steps)
	if err != nil {
		return nil, nil, err
	}

	return textResult(fmt.Sprintf("queued %d steps", len(in.Steps))), nil, nil
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
	var deadline = time.Now().Add(time.Duration(timeoutSec) * time.Second)

	if err := sleepCtx(ctx, pollInterval); err != nil {
		return nil, nil, err
	}

	var status string
	for time.Now().Before(deadline) {
		var err error
		status, err = s.api.getSessionStatus(in.Serial)
		if err != nil {
			return nil, nil, err
		}
		if !strings.HasPrefix(status, runningPrefix) {
			return textResult(status), nil, nil
		}
		if err := sleepCtx(ctx, pollInterval); err != nil {
			return nil, nil, err
		}
	}

	var text = fmt.Sprintf("timeout after %ds, last status: %s", timeoutSec, status)
	return textResult(text), nil, nil
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
		"route %s, %d steps; columns: id event timeout/delay landmarks (type \"value\" locale, chain joined by >; check = empty event)",
		route.Name, len(route.Steps),
	)
	if route.Prompt != "" {
		fmt.Fprintf(&lines, "\nprompt: %s", route.Prompt)
	}
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
	var text = fmt.Sprintf(
		"route %s queued, call wait_for_session for the outcome",
		in.Name,
	)
	return textResult(text), nil, nil
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
