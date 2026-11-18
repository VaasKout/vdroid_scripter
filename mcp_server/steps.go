package main

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type landmarkInput struct {
	Type   string `json:"type" jsonschema:"image | text | yolo | point"`
	Value  string `json:"value" jsonschema:"image name, text, yolo class or point as x:y in frame pixels; for type_text the text to type"`
	Locale string `json:"locale,omitempty" jsonschema:"Tesseract code of value, default eng"`
}

type stepInput struct {
	Event     string          `json:"event,omitempty" jsonschema:"tap, long_tap, type_text, swipe_up/down/left/right (half the screen) or the same with _small, _medium or _big for a sized swipe (swipe_up_big), a library action, or empty = visibility check"`
	Landmarks []landmarkInput `json:"landmarks,omitempty" jsonschema:"target chain, event on the last; empty = replay verbatim"`
	Timeout   *int            `json:"timeout,omitempty" jsonschema:"ms"`
	Delay     *int            `json:"delay,omitempty" jsonschema:"ms"`
}

type queueStepsInput struct {
	Serial string      `json:"serial" jsonschema:"device serial"`
	Steps  []stepInput `json:"steps" jsonschema:"in execution order"`
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

	status, finished, err := s.waitUntilNotRunning(ctx, in.Serial, toolWaitSec*time.Second)
	if err != nil {
		return nil, nil, err
	}
	var outcome = batchOutcome(fmt.Sprintf("queued %d steps", len(in.Steps)), status, finished)
	return textResult(outcome), nil, nil
}
