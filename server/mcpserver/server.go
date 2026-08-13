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

Workflow: list_devices for a serial, get_library for the available images (template crops) and actions (recorded gestures), then queue steps and wait for the outcome. Library names carry their context as <app>_<screen>_<what>[_variant] — e.g. swipe_x5_catalog_1 is a swipe recorded on the Pyaterochka catalog screen, first variant.

Batching: when given a sequence of steps ("tap text1, tap yolo class home, swipe, type hi..."), translate the WHOLE sequence into ONE queue_steps call with the steps in the given order. Never queue one step at a time and never poll get_session_status between steps — the server executes the queue sequentially on its own. After the single call, call wait_for_session once: 'idle' means every step succeeded.

Steps: a step is an event applied to a CV target (type = image | text | yolo, value = library image name / text to find / yolo class). Events tap and long_tap require a target and touch it. type_text types 'value' on the CV-detected keyboard (locale = keyboard locale). An EMPTY event is a pure visibility check of the target. Any other event name replays that recorded library event: anchored at the target's region when a target is given (e.g. a drag starting from an icon), verbatim without one (e.g. a scroll swipe).

Failure and recovery: a failed step clears the remaining queue and stores the error as the session status, so the final status names the target that could not be found. Recover from the failure point: scroll with the screen's swipe action (try variants _1, _2, ...) or probe an unknown screen with find_text, then re-queue the remaining steps from the failed one onward — again in one call.`

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

type stepInput struct {
	Event   string `json:"event,omitempty" jsonschema:"tap, long_tap, type_text, the name of a library event to replay, or EMPTY for a pure visibility check of the target"`
	Type    string `json:"type,omitempty" jsonschema:"target type: image (template match of a library image), text (OCR), or yolo (detected class); leave empty to replay a library event verbatim"`
	Value   string `json:"value,omitempty" jsonschema:"target value: library image name, text to find on screen, or yolo class name; for type_text the text to type"`
	Locale  string `json:"locale,omitempty" jsonschema:"OCR language for text targets and keyboard locale for type_text, e.g. eng or rus"`
	Timeout int    `json:"timeout,omitempty" jsonschema:"seconds to keep locating the target before failing, default 15"`
}

type queueStepsInput struct {
	Serial string      `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Steps  []stepInput `json:"steps" jsonschema:"steps to queue, executed in the given order"`
}

type findTextInput struct {
	Serial string `json:"serial" jsonschema:"device serial number"`
	Text   string `json:"text" jsonschema:"text to locate on the current screen"`
	Locale string `json:"locale,omitempty" jsonschema:"OCR language, e.g. eng or rus"`
}

func (s *stepInput) describe() string {
	if s.Event == "" {
		return fmt.Sprintf("check on %s %s", s.Type, s.Value)
	}
	if s.Event == "type_text" {
		return fmt.Sprintf("%s %q", s.Event, s.Value)
	}
	if s.Type != "" {
		return fmt.Sprintf("%s on %s %s", s.Event, s.Type, s.Value)
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
		Name: "queue_steps",
		Description: "Queue steps to run in order on the device; a session opens " +
			"automatically if none exists. Put a whole multi-step sequence into ONE call " +
			"— the server executes the queue sequentially; do not queue steps one at a " +
			"time or check status in between, just call wait_for_session once afterwards. " +
			"Each step applies an event to a CV-located target (type + value). Events: " +
			"'tap'/'long_tap' (target required; the generated touch lands at a random " +
			"point inside the found region), 'type_text' (types 'value' on the CV " +
			"keyboard), an EMPTY event (visibility check of the target, no touch), or a " +
			"library event name (replays the gesture anchored to the target region when " +
			"given, verbatim otherwise — queue a screen's swipe event without a target " +
			"to scroll). A step failure clears the remaining queue and sets the error " +
			"status.",
	}, s.handleQueueSteps)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "find_text",
		Description: "OCR the current screen for a text string without queueing a step; " +
			"returns the matching regions as JSON, [] when not found. Use it to probe an " +
			"unknown screen or to quickly verify an effect. Works without an open session.",
	}, s.handleFindText)

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

func (s *Server) handleQueueSteps(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in queueStepsInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" || len(in.Steps) == 0 {
		return nil, nil, fmt.Errorf("serial and steps are required")
	}
	for index := range in.Steps {
		if in.Steps[index].Event == "" && in.Steps[index].Value == "" {
			return nil, nil, fmt.Errorf("every step needs an event or a target")
		}
	}

	err := s.api.queueSteps(in.Serial, in.Steps)
	if err != nil {
		return nil, nil, err
	}

	var names = make([]string, 0, len(in.Steps))
	for index := range in.Steps {
		names = append(names, in.Steps[index].describe())
	}
	var text = "queued " + strings.Join(names, ", ")
	return textResult(text), nil, nil
}

func (s *Server) handleFindText(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in findTextInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" || in.Text == "" {
		return nil, nil, fmt.Errorf("serial and text are required")
	}
	result, err := s.api.findText(in.Serial, in.Text, in.Locale)
	if err != nil {
		return nil, nil, err
	}
	return textResult(result), nil, nil
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

func sleepCtx(ctx context.Context, duration time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}
