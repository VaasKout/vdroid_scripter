package main

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type landmarkInput struct {
	Type   string `json:"type" jsonschema:"image | text | yolo | point"`
	Value  string `json:"value" jsonschema:"image name, text, yolo class or point as x:y in frame pixels; for type_text the text to type"`
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

	status, finished, err := s.waitUntilNotRunning(ctx, in.Serial, toolWaitSec*time.Second)
	if err != nil {
		return nil, nil, err
	}
	var outcome = batchOutcome(fmt.Sprintf("queued %d steps", len(in.Steps)), status, finished)
	if mode != modeExplorer {
		return textResult(outcome), nil, nil
	}
	var succeeded = succeededSteps(recorded, status, finished)
	var recordable = recordableSteps(succeeded, in.Dictated)
	var recordNote = s.recordRoute(route, recordable)
	return textResult(outcome + "; " + recordNote), nil, nil
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
