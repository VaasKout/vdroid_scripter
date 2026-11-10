package main

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "vdroid-scripter"
	serverVersion = "0.1.0"
	runningPrefix = "running"
	idleStatus    = "idle"
	pollInterval  = 500 * time.Millisecond
	toolWaitSec   = 50

	defaultTimeoutMs = 5000
	defaultDelayMs   = 1000
	firstStepDelayMs = 0

	scanMinConfidence = 40

	routeArgPlaceholder = "%s"
	textLandmark        = "text"
)

// Server ...
type Server struct {
	api        *apiClient
	mcp        *mcp.Server
	mode       agentMode
	recordings recordings
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
		Name: "get_library",
		Description: "List library image and action names: the common items and, with serial, " +
			"that device's variants (name_<model_os>); other devices' variants are hidden.",
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
			"image name. Only for a target without text or yolo class. The image is common to " +
			"every device unless device_specific is true on the user's ask.",
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
			"status, or until a recording reported still running ends and return its result; " +
			"'closed' means the video stream ended, just queue again. Only for a batch or " +
			"recording reported still running, or a run started by the Android client.",
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
		Description: "Delete a saved route by name. Only on the user's ask.",
	}, s.handleDeleteRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "delete_image",
		Description: "Delete a library image by name. Only on the user's ask.",
	}, s.handleDeleteImage)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "delete_action",
		Description: "Delete a library action by name. Only on the user's ask.",
	}, s.handleDeleteAction)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "rename_image",
		Description: "Rename a library image in place, the same picture under the new name; a device-specific variant keeps its suffix. The only way to rename: never delete and save again.",
	}, s.handleRenameImage)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "rename_action",
		Description: "Rename a library action in place, the same gesture under the new name; a device-specific variant keeps its suffix. The only way to rename: never delete and record again.",
	}, s.handleRenameAction)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "run_route",
		Description: "Run a saved route, or its start_id..end_id slice, and block until " +
			"it finishes: 'idle' or the failed step's error.",
	}, s.handleRunRoute)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name: "record_action",
		Description: "Record a gesture the HUMAN performs on the device and save it as " +
			"library action name: touches and hardware keys (volume, power, home, back, " +
			"menu, recents, camera) with their timing. The first touch " +
			"or key must come within timeout of the call and the recording ends timeout " +
			"after the last one; 'nothing recorded' = no input, ask them to retry; a " +
			"recording still running after 50 s finishes through wait_for_session. Only " +
			"when the user asks; tell them to start right away. The action is common to " +
			"every device unless device_specific is true on the user's ask.",
	}, s.handleRecordAction)
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func sleepCtx(ctx context.Context, duration time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}
