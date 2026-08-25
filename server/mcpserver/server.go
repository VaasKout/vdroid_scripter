package main

import (
	"context"
	"fmt"
	"net/http"
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

Workflow: list_devices for a serial, get_library for the available images (template crops) and actions (recorded gestures), then queue steps and wait for the outcome. Library names carry their context as <app>_<screen>_<what>[_variant] — e.g. swipe_x5_catalog_1 is a swipe recorded on the Pyaterochka catalog screen, first variant.

Batching: when given a sequence of steps ("tap text1, tap yolo class home, swipe, type hi..."), translate the WHOLE sequence into ONE queue_steps call with the steps in the given order. Never queue one step at a time and never poll get_session_status between steps — the server executes the queue sequentially on its own. After the single call, call wait_for_session once: 'idle' means every step succeeded.

Steps: a step is an event applied to a chain of landmarks. Each landmark is a CV target (type = image | text | yolo, value = library image name / text to find / yolo class; text landmarks also carry locale). The chain resolves on ONE video frame: the first landmark picks its best match on screen, every following landmark picks the candidate of its value NEAREST to the previous landmark, and the event applies to the LAST landmark. One landmark is the normal case ("tap the cart icon" = one image landmark). Put a nearby unique element first to disambiguate duplicates: "tap the toggle next to 'Show refresh rate'" = landmarks [{type text, value Show refresh rate}, {type image, value toggle}]. Events tap and long_tap touch the last landmark's region. type_text types the LAST landmark's value on the CV-detected keyboard (that landmark's locale = keyboard locale; nothing is located on screen). An EMPTY event is a pure visibility check of the landmark chain. Any other event name replays that recorded library event: offset into the last landmark's region when landmarks are given (e.g. a drag starting from an icon), verbatim without landmarks (e.g. a scroll swipe).

Locale: for text landmarks and type_text, always set the landmark's locale to the Tesseract language code of its value's language. This holds for every language Tesseract supports (rus, deu, fra, jpn, ...); eng is the default. Pass the text exactly as the user wrote it, never transliterate or translate it.

Library curation: screenshot and save_image exist ONLY for saving a new library image, and ONLY when the user explicitly asks to add one. "Save/add to the library" are the key words that select this flow: a request like "find the settings icon and save it to the library" means screenshot + save_image, NOT queue_steps — "find" there means picking the element's rectangle on the screenshot, not a visibility-check step. Then: call screenshot with rectangles=true, pick the rectangle that bounds the element by looking at the returned image, and call save_image with a library name and that rectangle. The screen must not change between the screenshot and save_image — the server re-captures the screen when cropping. save_image is the LAST call of the flow. After it STOP IMMEDIATELY and report the saved name — under NO circumstances call any tool after save_image: no queue_steps, no get_session_status, no wait_for_session, no close_session, no screenshot, nothing. The session tools are off-limits after save_image even if a session is already open, even if a check seems helpful, even on doubt about the crop. The new image works as an image landmark whenever the user later references it. NEVER call screenshot on your own while running steps — not to locate an element, not to verify state, not to inspect a failure. Finding and verifying elements is always done with steps (EMPTY event visibility checks).

Rules: NEVER drive the device with adb directly — no adb shell input tap, input swipe, input text, keyevent, or any other adb command, no matter what. Every interaction is a step executed through queue_steps: tap/long_tap to touch a target, a library event to gesture, type_text to type, and an EMPTY event to find or verify an element. There is no separate lookup tool — finding an element and acting on it are both steps. The one exception: when the user asks to SAVE an element to the library, that is the curation flow below (screenshot + save_image), not steps.

Literal execution: when the user names a concrete action, queue exactly that action and nothing else — no extra visibility checks, no probing, no added, substituted or reordered steps, no "better" alternatives. When the user asks for the same thing repeatedly, execute it again every time, exactly as many times as asked — never skip a repeat because it was already done and never deduplicate. Never argue, never ask for confirmation — just execute. Improvise only when a step fails (see recovery below).

Map: the map is a graph of curated screens. A node has a name, static landmarks that identify the screen, edges (an action plus the nodes it can lead to; empty next_nodes = a final page), and an occupancy grid the server fills on its own. Curation is EXPLICIT, like the library: create or extend nodes with save_map_node ONLY when the user asks ("new node ..."); the node name is required — ask for it if the dictation has none. Once nodes exist, prefer follow_route ("go from main_screen to x5_catalog") over re-dictating step sequences: the server BFS-plans the path and verifies every hop, landmarks first, occupancy grid second.

Lost routes: after follow_route call wait_for_session once. A status starting with "lost" means the route lost localization. Report that status to the user VERBATIM and ask what to do next. Never improvise recovery on a route — no probing, no scrolling, no screenshot. The scroll-and-retry recovery below applies ONLY to plain queue_steps failures.

Duplicates: when a landmark value matches several places on screen, disambiguate with a chain — put a unique nearby landmark first; the following landmark resolves nearest to it. When no unique neighbor exists, a bare landmark deterministically takes the FIRST candidate in reading order (top to bottom, left to right).

Failure and recovery: a failed step clears the remaining queue and stores the error as the session status, so the final status names the target that could not be found. Recover from the failure point: scroll with the screen's swipe action (try variants _1, _2, ...) or probe an unknown screen with visibility-check steps (EMPTY event, text targets), then re-queue the remaining steps from the failed one onward — again in one call.`

type Server struct {
	api *apiClient
	mcp *mcp.Server
}

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
	Event     string          `json:"event,omitempty" jsonschema:"tap, long_tap, type_text, the name of a library event to replay, or EMPTY for a pure visibility check of the target"`
	Landmarks []landmarkInput `json:"landmarks,omitempty" jsonschema:"target chain resolved on one video frame: each landmark is located NEAREST to the previous one and the event applies to the LAST landmark; one landmark for a plain target, a preceding unique landmark to disambiguate duplicates; leave empty to replay a library event verbatim"`
	Timeout   int             `json:"timeout,omitempty" jsonschema:"seconds to keep locating the target before failing, default 15"`
}

type queueStepsInput struct {
	Serial string      `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Steps  []stepInput `json:"steps" jsonschema:"steps to queue, executed in the given order"`
}

type screenshotInput struct {
	Serial     string `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Rectangles bool   `json:"rectangles,omitempty" jsonschema:"true to also return detected UI element rectangles as JSON, for picking a crop region"`
}

type rectangleInput struct {
	LeftX   int `json:"left_x" jsonschema:"left edge X in screenshot pixels"`
	RightX  int `json:"right_x" jsonschema:"right edge X in screenshot pixels"`
	TopY    int `json:"top_y" jsonschema:"top edge Y in screenshot pixels"`
	BottomY int `json:"bottom_y" jsonschema:"bottom edge Y in screenshot pixels"`
}

type saveImageInput struct {
	Serial    string         `json:"serial" jsonschema:"device serial number"`
	Name      string         `json:"name" jsonschema:"library name for the template, <app>_<screen>_<what>[_variant]; a duplicate name overwrites"`
	Rectangle rectangleInput `json:"rectangle" jsonschema:"the region to crop from the current screen, in screenshot pixel coordinates"`
}

type edgeInput struct {
	Action    stepInput `json:"action" jsonschema:"the step this edge performs: an event applied to a landmark chain, or a library event"`
	NextNodes []string  `json:"next_nodes,omitempty" jsonschema:"node names this action can lead to; empty for a terminal action on a final page; names may reference nodes that don't exist yet"`
}

type gridInput struct {
	Cols  int      `json:"cols" jsonschema:"grid width in cells, 18"`
	Rows  int      `json:"rows" jsonschema:"grid height in cells, 32"`
	Cells []string `json:"cells" jsonschema:"rows of '0'/'1' characters, one string per row"`
}

type saveMapNodeInput struct {
	Name          string          `json:"name" jsonschema:"node name, REQUIRED, <app>_<screen> convention; posting an existing name merges into it"`
	Landmarks     []landmarkInput `json:"landmarks,omitempty" jsonschema:"static landmarks identifying this screen; ALL must be visible for a landmark verification to pass"`
	Edges         []edgeInput     `json:"edges,omitempty" jsonschema:"actions available on this screen with the nodes they lead to"`
	OccupancyGrid *gridInput      `json:"occupancy_grid,omitempty" jsonschema:"optional structural fingerprint; non-empty replaces the stored grid, omitted keeps it"`
}

type mapNodeNameInput struct {
	Name string `json:"name" jsonschema:"node name from get_map"`
}

type followRouteInput struct {
	Serial string `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	From   string `json:"from" jsonschema:"node the device is currently on"`
	To     string `json:"to" jsonschema:"destination node"`
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
			"gestures like swipes that cannot be generated (usable as a step's action). " +
			"Names encode their context as <app>_<screen>_<what>[_variant]. " +
			"Call this before planning steps.",
	}, s.handleGetLibrary)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "screenshot",
		Description: "Take a screenshot of the device's current screen. With " +
			"rectangles=true it also returns detected UI element rectangles as a JSON " +
			"array (left_x, right_x, top_y, bottom_y) — use it to pick the region of " +
			"an element you want to save as a library template with save_image. Call " +
			"it ONLY when the user asks to save a library image — never to locate " +
			"elements, verify state, or inspect failures while running steps; every " +
			"lookup is a step (EMPTY event visibility check).",
	}, s.handleScreenshot)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "save_image",
		Description: "Crop a region of the device's CURRENT screen into the library " +
			"as a named template image, immediately usable as an image landmark in " +
			"steps. The server re-captures the screen when cropping, so the screen " +
			"must still show what the screenshot showed. Name it " +
			"<app>_<screen>_<what>[_variant]; a duplicate name overwrites. This is " +
			"the LAST call of the flow: after it make NO further tool calls under any " +
			"circumstances — no queue_steps, get_session_status, wait_for_session, " +
			"close_session or screenshot. Report the saved name and finish.",
	}, s.handleSaveImage)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "queue_steps",
		Description: "Queue steps to run in order on the device; a session opens " +
			"automatically if none exists. Put a whole multi-step sequence into ONE call " +
			"— the server executes the queue sequentially; do not queue steps one at a " +
			"time or check status in between, just call wait_for_session once afterwards. " +
			"Each step applies an event to the LAST landmark of its chain; landmarks resolve " +
			"on one frame, each located nearest to the previous one. Events: " +
			"'tap'/'long_tap' (landmarks required; the generated touch lands at a random " +
			"point inside the found region), 'type_text' (types the last landmark's value " +
			"on the CV keyboard), an EMPTY event (visibility check of the chain, no " +
			"touch), or a library event name (replays the gesture offset into the found " +
			"region when landmarks are given, verbatim otherwise — queue a screen's swipe " +
			"event without landmarks to scroll). A step failure clears the remaining queue " +
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
		Name: "get_map",
		Description: "List the names of all map nodes (curated screens). " +
			"Call this before planning routes or saving nodes.",
	}, s.handleGetMap)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_map_node",
		Description: "Get one map node: its landmarks, edges (actions with the " +
			"nodes they lead to), and occupancy grid.",
	}, s.handleGetMapNode)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "save_map_node",
		Description: "Create or update a map node. Use ONLY when the user " +
			"explicitly curates the map (\"new node ...\"). Name is REQUIRED — " +
			"if the user's dictation doesn't name the node, ask for the name. " +
			"Posting an existing name merges: landmarks and edges are appended, " +
			"a repeated action unions its next_nodes, a non-empty occupancy_grid " +
			"replaces the stored one.",
	}, s.handleSaveMapNode)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "delete_map_node",
		Description: "Delete a map node by name. References to it in other " +
			"nodes' next_nodes become dangling and are simply unroutable.",
	}, s.handleDeleteMapNode)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "follow_route",
		Description: "Navigate the device from one map node to another. The " +
			"server plans the shortest path over edges (BFS) and executes it, " +
			"verifying position at every hop: landmarks first, occupancy grid " +
			"as fallback. Call wait_for_session afterwards; 'idle' means " +
			"arrived, a status starting with 'lost' means the route lost " +
			"localization — report it to the user verbatim and ASK what to do; " +
			"never improvise recovery on routes.",
	}, s.handleFollowRoute)
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

func (s *Server) handleScreenshot(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in screenshotInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" {
		return nil, nil, fmt.Errorf("serial is required")
	}

	image, rectangles, err := s.api.getScreenshot(in.Serial, in.Rectangles)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.ImageContent{Data: image, MIMEType: http.DetectContentType(image)},
	}
	if rectangles != "" {
		content = append(content, &mcp.TextContent{Text: rectangles})
	}
	return &mcp.CallToolResult{Content: content}, nil, nil
}

func (s *Server) handleSaveImage(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in saveImageInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" || in.Name == "" {
		return nil, nil, fmt.Errorf("serial and name are required")
	}

	err := s.api.saveImage(in.Serial, in.Name, in.Rectangle)
	if err != nil {
		return nil, nil, err
	}
	return textResult("saved " + in.Name + ". STOP: make no further tool calls — " +
		"no queue_steps, get_session_status, wait_for_session, close_session or " +
		"screenshot. Report the saved name to the user and finish."), nil, nil
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

func (s *Server) handleGetMap(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in emptyInput,
) (*mcp.CallToolResult, any, error) {
	nodes, err := s.api.getMap()
	if err != nil {
		return nil, nil, err
	}
	return textResult(nodes), nil, nil
}

func (s *Server) handleGetMapNode(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in mapNodeNameInput,
) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	node, err := s.api.getMapNode(in.Name)
	if err != nil {
		return nil, nil, err
	}
	return textResult(node), nil, nil
}

func (s *Server) handleSaveMapNode(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in saveMapNodeInput,
) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	err := s.api.saveMapNode(&in)
	if err != nil {
		return nil, nil, err
	}
	return textResult("saved node " + in.Name), nil, nil
}

func (s *Server) handleDeleteMapNode(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in mapNodeNameInput,
) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	err := s.api.deleteMapNode(in.Name)
	if err != nil {
		return nil, nil, err
	}
	return textResult("deleted node " + in.Name), nil, nil
}

func (s *Server) handleFollowRoute(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in followRouteInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" || in.From == "" || in.To == "" {
		return nil, nil, fmt.Errorf("serial, from and to are required")
	}
	err := s.api.followRoute(in.Serial, in.From, in.To)
	if err != nil {
		return nil, nil, err
	}
	var text = fmt.Sprintf(
		"route %s -> %s queued, call wait_for_session for the outcome",
		in.From,
		in.To,
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
