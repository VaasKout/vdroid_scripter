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
)

const serverInstructions = `vdroid-scripter drives Android devices with CV-located steps composed from a human-curated library.

Workflow: list_devices for a serial, get_routes for saved flows, then queue steps and wait for the outcome. Call get_library only when image targets or recorded gestures might be needed — it lists the available images (template crops) and actions (recorded gestures); library names carry their context as <app>_<screen>_<what>[_variant] — e.g. swipe_x5_catalog_1 is a swipe recorded on the Pyaterochka catalog screen, first variant.

Text is free: text landmarks and the generated events (tap, long_tap, the swipes, type_text) need NOTHING from the library. An instruction phrased in words visible on screen ("open Settings", "enter wifi connections") is just tap steps with text landmarks — tap the matching words, drilling through the obvious screens (e.g. Settings -> Network & internet -> Wi-Fi). Only reach for the library when the target has no readable text (an icon = image landmark) or needs a recorded gesture.

Batching: when given a sequence of steps ("tap text1, tap yolo class home, swipe, type hi..."), translate the WHOLE sequence into ONE queue_steps call with the steps in the given order. Never queue one step at a time and never poll get_session_status between steps — the server executes the queue sequentially on its own. After the single call, call wait_for_session once: 'idle' means every step succeeded.

Steps: a step is an event applied to a chain of landmarks. Each landmark is a CV target (type = image | text | yolo, value = library image name / text to find / yolo class; text landmarks also carry locale). The chain resolves on ONE video frame: the first landmark picks its best match on screen, every following landmark picks the candidate of its value NEAREST to the previous landmark, and the event applies to the LAST landmark. One landmark is the normal case ("tap the cart icon" = one image landmark). Put a nearby unique element first to disambiguate duplicates: "tap the toggle next to 'Show refresh rate'" = landmarks [{type text, value Show refresh rate}, {type image, value toggle}]. Events tap and long_tap touch the last landmark's region. swipe_up, swipe_down, swipe_left and swipe_right generate a human-like fixed-length swipe (direction = the finger's movement, so swipe_up reveals content below): without landmarks it starts at a random point on screen, with landmarks it starts inside the last landmark's region — use these for plain scrolling; recorded library swipes remain for app-specific gestures. type_text types the LAST landmark's value on the CV-detected keyboard (that landmark's locale = keyboard locale; nothing is located on screen). An EMPTY event is a pure visibility check of the landmark chain. Any other event name replays that recorded library event: offset into the last landmark's region when landmarks are given (e.g. a drag starting from an icon), verbatim without landmarks.

Timing: every step has two knobs. delay (milliseconds, default 1000) is slept BEFORE the step acts — it paces the flow and gives the previous action's screen change time to settle. timeout (seconds, default 3) is how long the server keeps re-locating the target on live frames before failing — retries run back to back, and the step proceeds the moment the target shows up. For a target that appears late (app launch, a navigation tap, network loading) raise timeout (10-15) rather than delay; use a bigger delay only when the target is visible early but not yet safe to touch (mid-animation). Keep probe/visibility checks at the short default timeout so a negative answer comes back fast. Session startup never eats the step timeout — the first video frame gets its own grace period.

Locale: for text landmarks and type_text, always set the landmark's locale to the Tesseract language code of its value's language. This holds for every language Tesseract supports (rus, deu, fra, jpn, ...); eng is the default. Pass the text exactly as the user wrote it, never transliterate or translate it.

Perception: scan is the ONLY way to observe the screen — there are no screenshots and never will be. Call scan when a step failed, when the user's instruction is conditional ("if X is not visible, ..."), or when the user explicitly asks what is on screen. Never scan habitually between steps — the happy path is one queue_steps call and one wait_for_session. Pass in images the library image names plausibly related to the current app so scan reports which of them are visible; scan's landmarks are exactly what step landmarks consume — build follow-up steps from the returned type/value pairs.

Routes: a route is a saved flow — a name, the user's dictation as its prompt, and the exact steps that ran to success. When the user asks to save or remember a flow as <name>, call save_route with the name, the user's dictation VERBATIM as the prompt (conditions included), and the steps that actually succeeded in order; a duplicate name overwrites. The prompt may be absent on routes saved elsewhere (the Android client saves routes without one) — treat such a route as a plain script with no recorded intent. To run a saved route: run_route, then wait_for_session once — 'idle' means the whole route succeeded. To extend a route: get_route, append the new steps, call save_route with the full list — nothing executes. If a route step fails, recover from the failure point guided by the route's prompt: scan, decide, queue the remaining steps with queue_steps — and after a recovered run ask the user whether to update the route with the steps that worked. Never create or modify routes without being asked.

Curation: library images and actions are created by the human with the Android client. There are no tools here to create them. Ask the user to add a library item ONLY when the target truly cannot be reached any other way — no readable text for a text landmark, no yolo class, no generated swipe that gets there. Never request curation for something written on the screen.

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
	Type   string `json:"type" jsonschema:"landmark type: image (template match of a library image), text (OCR), or yolo (detected class)"`
	Value  string `json:"value" jsonschema:"library image name, text to find on screen, or yolo class name; for type_text the text to type"`
	Locale string `json:"locale,omitempty" jsonschema:"for text landmarks and type_text: the Tesseract lang code matching the language of value (rus, deu, jpn, ...), default eng"`
}

type stepInput struct {
	Event     string          `json:"event,omitempty" jsonschema:"tap, long_tap, type_text, a generated swipe (swipe_up, swipe_down, swipe_left, swipe_right), the name of a library event to replay, or EMPTY for a pure visibility check of the target"`
	Landmarks []landmarkInput `json:"landmarks,omitempty" jsonschema:"target chain resolved on one video frame: each landmark is located NEAREST to the previous one and the event applies to the LAST landmark; one landmark for a plain target, a preceding unique landmark to disambiguate duplicates; leave empty to replay a library event verbatim"`
	Timeout   int             `json:"timeout,omitempty" jsonschema:"seconds to keep locating the target before failing, default 3 — enough for elements already on screen; raise it (10-15) for targets that appear after an app launch, navigation or loading"`
	Delay     int             `json:"delay,omitempty" jsonschema:"milliseconds slept BEFORE the step acts, default 1000 — paces the flow and lets the previous action's screen change settle; raise it only when the target is visible but mid-animation"`
}

type queueStepsInput struct {
	Serial string      `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Steps  []stepInput `json:"steps" jsonschema:"steps to queue, executed in the given order"`
}

type scanInput struct {
	Serial string   `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Images []string `json:"images,omitempty" jsonschema:"library image names to search for on the screen; pass the images plausibly related to the current app; omit for text+yolo only"`
	Locale string   `json:"locale,omitempty" jsonschema:"Tesseract lang code for the OCR pass (rus, deu, jpn, ...), default eng"`
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
	Serial string `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Name   string `json:"name" jsonschema:"route name from get_routes"`
}

func (s *stepInput) describe() string {
	var target string
	if len(s.Landmarks) > 0 {
		var last = s.Landmarks[len(s.Landmarks)-1]
		target = fmt.Sprintf("%s %s", last.Type, last.Value)
	}

	if s.Event == "" {
		return "check on " + target
	}
	if s.Event == "type_text" {
		if len(s.Landmarks) == 0 {
			return s.Event
		}
		return fmt.Sprintf("%s %q", s.Event, s.Landmarks[len(s.Landmarks)-1].Value)
	}
	if target != "" {
		return fmt.Sprintf("%s on %s", s.Event, target)
	}
	return s.Event
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "list_devices",
		Description: "List connected Android devices with their serial numbers. " +
			"Call this first to find the serial required by every other tool.",
	}, s.handleListDevices)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_library",
		Description: "List the automation library: 'images' are template crops the human " +
			"saved from device screens (usable as image targets), 'actions' are recorded " +
			"gestures (usable as a step's event). " +
			"Names encode their context as <app>_<screen>_<what>[_variant]. " +
			"Call this only when an image target or a recorded gesture might be needed — " +
			"text landmarks and the generated tap/long_tap/swipe/type_text events use " +
			"nothing from the library.",
	}, s.handleGetLibrary)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "scan",
		Description: "Scan the device's current screen and return the landmarks found: " +
			"every text with its rectangle (OCR, using locale), every yolo detection, " +
			"and matches for the library images passed in images. This is the ONLY way " +
			"to observe the screen — there are no screenshots. Call it when a step " +
			"failed, when the user's instruction is conditional, or when the user asks " +
			"what is on screen — never habitually between steps. What it returns is " +
			"exactly what step landmarks consume. Opens a session automatically if " +
			"none exists.",
	}, s.handleScan)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "queue_steps",
		Description: "Queue steps to run in order on the device; a session opens " +
			"automatically if none exists. Put a whole multi-step sequence into ONE call " +
			"— the server executes the queue sequentially; do not queue steps one at a " +
			"time or check status in between, just call wait_for_session once afterwards. " +
			"Each step applies an event to the LAST landmark of its chain; landmarks resolve " +
			"on one frame, each located nearest to the previous one. Events: " +
			"'tap'/'long_tap' (landmarks required; the generated touch lands at a random " +
			"point inside the found region), 'swipe_up'/'swipe_down'/'swipe_left'/" +
			"'swipe_right' (generated human-like fixed-length swipe, named by the " +
			"finger's direction; random start point on screen, or inside the last " +
			"landmark's region when landmarks are given — the default way to scroll), " +
			"'type_text' (types the last landmark's value " +
			"on the CV keyboard), an EMPTY event (visibility check of the chain, no " +
			"touch), or a library event name (replays the gesture offset into the found " +
			"region when landmarks are given, verbatim otherwise). " +
			"A step failure clears the remaining queue " +
			"and sets the error status.",
	}, s.handleQueueSteps)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "close_session",
		Description: "Close the device session and stop screen capture. " +
			"Call when the flow is finished.",
	}, s.handleCloseSession)

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
		Description: "Queue a saved route's steps on the device; a session opens " +
			"automatically if none exists. Call wait_for_session once afterwards: " +
			"'idle' means the whole route succeeded; an error status names the step " +
			"that failed — recover from that point guided by the route's prompt.",
	}, s.handleRunRoute)
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
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
	return textResult(devices), nil, nil
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
	return textResult(library), nil, nil
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
	return textResult(landmarks), nil, nil
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

	var names = make([]string, 0, len(in.Steps))
	for _, step := range in.Steps {
		names = append(names, step.describe())
	}
	var text = "queued " + strings.Join(names, ", ")
	return textResult(text), nil, nil
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
	return textResult(routes), nil, nil
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
	return textResult(route), nil, nil
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
	err := s.api.runRoute(in.Serial, in.Name)
	if err != nil {
		return nil, nil, err
	}
	var text = fmt.Sprintf(
		"route %s queued, call wait_for_session for the outcome",
		in.Name,
	)
	return textResult(text), nil, nil
}

func sleepCtx(ctx context.Context, duration time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}
