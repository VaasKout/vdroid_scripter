package main

import (
	"android_vision_scripter/pkg/models"
	"context"
	"encoding/json"
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

const serverInstructions = `vdroid-scripter drives Android devices by replaying recorded CV automation scripts.

Building a flow: call list_devices, then get_location_graph, then plan a chain of scripts where each script's next_location contains the location of the next step. Queue the whole flow with one queue_scripts call (scripts run in the given order) and wait for the result with wait_for_session.

Failure recovery: an error status means the failed script could not find its CV params on screen, so the device is probably not on the location you assumed. To find out the actual location, queue the previous successful script (the last one that completed before the failure) and wait for the result. If it succeeds, the device was still on that script's location and has now moved to its next_location — re-queue the failed script and continue the flow from there. If it fails too, the device is on an unknown screen: re-plan the flow from get_location_graph, or close_session and start over.`

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

type queuedScript struct {
	Location string `json:"location" jsonschema:"location (screen) name the script belongs to"`
	Script   string `json:"script" jsonschema:"script name inside the location"`
}

type queueScriptsInput struct {
	Serial  string         `json:"serial" jsonschema:"device serial number, get it from list_devices"`
	Scripts []queuedScript `json:"scripts" jsonschema:"scripts to queue, executed in the given order"`
}

type waitInput struct {
	Serial         string `json:"serial" jsonschema:"device serial number"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"max seconds to wait, default 60"`
}

type scriptInfo struct {
	Name         string             `json:"name"`
	NextLocation []string           `json:"next_location,omitempty"`
	Params       []models.Parameter `json:"params,omitempty"`
	Timeout      int                `json:"timeout,omitempty"`
	EventsCount  int                `json:"events_count"`
}

type locationInfo struct {
	Location string       `json:"location"`
	Scripts  []scriptInfo `json:"scripts"`
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "list_devices",
		Description: "List connected Android devices with their serial numbers. " +
			"Call this first to find the serial required by every other tool.",
	}, s.handleListDevices)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_location_graph",
		Description: "Return the full automation map: every location (a screen in an app) " +
			"with its scripts. Each script lists next_location — the locations it navigates to " +
			"when executed — plus its CV params (what it expects to see on screen) and whether " +
			"it has recorded touch events. Use this to plan a flow: pick a chain of scripts " +
			"where each script's next_location contains the next step's location, from the " +
			"start location to the goal.",
	}, s.handleGetLocationGraph)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "open_session",
		Description: "Open a device session (starts screen capture). A session is required " +
			"before scripts can execute. queue_scripts opens one automatically, so call this " +
			"only when you want the session up before queueing.",
	}, s.handleOpenSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "close_session",
		Description: "Close the device session and stop screen capture. " +
			"Call when the flow is finished.",
	}, s.handleCloseSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "get_session_status",
		Description: "Get the session status for a device. Values: 'closed' (no session), " +
			"'idle' (session open, script queue empty — previous scripts all succeeded), " +
			"'running <script> on <location>' (a script is executing), or an error text like " +
			"'unable to find parameter ...' meaning the last script failed and the remaining " +
			"queue was cleared. Poll this after queueing scripts to track flow progress. " +
			"On an error status the device may not be on the location you assumed: queue the " +
			"previous successful script to discover the actual location before continuing.",
	}, s.handleGetSessionStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "queue_scripts",
		Description: "Queue one or more scripts for execution on the device. Scripts run " +
			"sequentially in the given order on a single session; if no session exists one " +
			"is opened automatically. Queue the whole planned flow in one call, then poll " +
			"get_session_status or call wait_for_session. A script failure clears the " +
			"remaining queue and puts the error into the session status.",
	}, s.handleQueueScripts)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "wait_for_session",
		Description: "Block until the device session stops running queued scripts, then " +
			"return the final status: 'idle' means the flow completed successfully, an error " +
			"text means a script failed (remaining queue was cleared), 'closed' means the " +
			"session ended. Call this after queue_scripts instead of polling manually. " +
			"On an error status, recover by queueing the previous successful script: success " +
			"means the device was still on that script's location (re-queue the failed script " +
			"and continue), another failure means the screen is unknown — re-plan the flow.",
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

func (s *Server) handleGetLocationGraph(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in emptyInput,
) (*mcp.CallToolResult, any, error) {
	locations, err := s.api.getLocations()
	if err != nil {
		return nil, nil, err
	}

	var graph = make([]locationInfo, 0, len(locations))
	for _, location := range locations {
		names, err := s.api.getLocationScripts(location)
		if err != nil {
			return nil, nil, err
		}

		var scripts = make([]scriptInfo, 0, len(names))
		for _, name := range names {
			script, err := s.api.getScript(location, name)
			if err != nil || script.Name == "" {
				continue
			}
			scripts = append(scripts, scriptInfo{
				Name:         script.Name,
				NextLocation: script.NextLocation,
				Params:       script.Params,
				Timeout:      script.Timeout,
				EventsCount:  len(script.Events),
			})
		}
		graph = append(graph, locationInfo{Location: location, Scripts: scripts})
	}

	bytes, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return textResult(string(bytes)), nil, nil
}

func (s *Server) handleOpenSession(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in serialInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" {
		return nil, nil, fmt.Errorf("serial is required")
	}
	ports, err := s.api.openSession(in.Serial)
	if err != nil {
		return nil, nil, err
	}
	return textResult("session opened: " + ports), nil, nil
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

func (s *Server) handleQueueScripts(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in queueScriptsInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" || len(in.Scripts) == 0 {
		return nil, nil, fmt.Errorf("serial and scripts are required")
	}

	var refs = make([]scriptRef, 0, len(in.Scripts))
	for _, script := range in.Scripts {
		if script.Location == "" || script.Script == "" {
			return nil, nil, fmt.Errorf("location and script are required in every entry")
		}
		refs = append(refs, scriptRef{Location: script.Location, Name: script.Script})
	}

	err := s.api.queueScripts(in.Serial, refs)
	if err != nil {
		return nil, nil, err
	}

	var names = make([]string, 0, len(refs))
	for _, ref := range refs {
		names = append(names, ref.Location+"/"+ref.Name)
	}
	var text = "queued " + strings.Join(names, ", ")
	return textResult(text), nil, nil
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
