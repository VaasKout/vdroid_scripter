package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type waitInput struct {
	Serial         string `json:"serial" jsonschema:"device serial"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"max seconds, default 50"`
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
		timeoutSec = toolWaitSec
	}
	var wait = time.Duration(timeoutSec) * time.Second

	if recording, ok := s.recordings.get(in.Serial); ok {
		return s.finishRecording(ctx, in.Serial, recording, wait)
	}
	status, finished, err := s.waitUntilNotRunning(ctx, in.Serial, wait)
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
		prefix, toolWaitSec, status,
	)
}
