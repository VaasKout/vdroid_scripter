package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type saveRouteInput struct {
	Name  string      `json:"name" jsonschema:"<app>_<flow>"`
	Steps []stepInput `json:"steps" jsonschema:"the steps that ran to success, in order"`
}

type routeNameInput struct {
	Name string `json:"name" jsonschema:"route name"`
}

type runRouteInput struct {
	Serial     string   `json:"serial" jsonschema:"device serial"`
	Name       string   `json:"name" jsonschema:"route name"`
	StartID    int      `json:"start_id,omitempty" jsonschema:"first step id to run; omit to start at step 1"`
	EndID      int      `json:"end_id,omitempty" jsonschema:"last step id to run; omit to run to the end"`
	Args       []string `json:"args,omitempty" jsonschema:"one value per %s of the route, in step order"`
	ArgsLocale string   `json:"args_locale,omitempty" jsonschema:"Tesseract code of the args' language, default eng"`
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
		"route %s, %d steps, %d args: id event timeout/delay landmarks (check = empty event)",
		route.Name, len(route.Steps), routeArgsCount(route),
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

func routeArgsCount(route routeResponse) int {
	count := 0
	for _, step := range route.Steps {
		for _, landmark := range step.Landmarks {
			if landmark.Type != textLandmark {
				continue
			}
			count += strings.Count(landmark.Value, routeArgPlaceholder)
		}
	}
	return count
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
	err := s.api.runRoute(in)
	if err != nil {
		return nil, nil, err
	}
	status, finished, err := s.waitUntilNotRunning(ctx, in.Serial, toolWaitSec*time.Second)
	if err != nil {
		return nil, nil, err
	}
	return textResult(batchOutcome("route "+in.Name, status, finished)), nil, nil
}
