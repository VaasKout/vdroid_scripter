package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type editRouteInput struct {
	Name      string          `json:"name" jsonschema:"route name"`
	ID        int             `json:"id" jsonschema:"step id"`
	Delete    bool            `json:"delete,omitempty" jsonschema:"remove the step"`
	Timeout   *int            `json:"timeout,omitempty" jsonschema:"new timeout, ms"`
	Delay     *int            `json:"delay,omitempty" jsonschema:"new delay, ms"`
	Landmarks []landmarkInput `json:"landmarks,omitempty" jsonschema:"new target chain"`
}

func (s *Server) handleEditRoute(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in editRouteInput,
) (*mcp.CallToolResult, any, error) {
	if in.Name == "" || in.ID < 1 {
		return nil, nil, fmt.Errorf("name and id are required")
	}
	if in.Delete == in.hasChanges() {
		return nil, nil, fmt.Errorf("send either delete or new timeout, delay or landmarks")
	}

	route, err := s.api.getRoute(in.Name)
	if err != nil {
		return nil, nil, err
	}
	index := routeStepIndex(route.Steps, in.ID)
	if index < 0 {
		return nil, nil, fmt.Errorf("route %s has no step %d", in.Name, in.ID)
	}

	steps, summary := in.apply(route, index)
	err = s.api.saveRoute(&saveRouteInput{Name: in.Name, Steps: steps})
	if err != nil {
		return nil, nil, err
	}
	return textResult(summary), nil, nil
}

func (e editRouteInput) hasChanges() bool {
	return e.Timeout != nil || e.Delay != nil || len(e.Landmarks) > 0
}

func (e editRouteInput) apply(route routeResponse, index int) ([]stepInput, string) {
	steps := routeInputs(route.Steps)
	if e.Delete {
		steps = slices.Delete(steps, index, index+1)
		return steps, fmt.Sprintf(
			"route %s step %d deleted, %d steps left, later ids shift down by one",
			e.Name, e.ID, len(steps),
		)
	}

	changes := e.changes(route.Steps[index])
	steps[index] = e.patch(steps[index])
	return steps, fmt.Sprintf("route %s step %d: %s", e.Name, e.ID, strings.Join(changes, ", "))
}

func (e editRouteInput) changes(step routeStep) []string {
	changes := make([]string, 0, 3)
	if e.Timeout != nil {
		changes = append(changes, fmt.Sprintf("timeout %d -> %d", step.Timeout, *e.Timeout))
	}
	if e.Delay != nil {
		changes = append(changes, fmt.Sprintf("delay %d -> %d", step.Delay, *e.Delay))
	}
	if len(e.Landmarks) > 0 {
		changes = append(changes, fmt.Sprintf(
			"landmarks %s -> %s",
			formatChain(step.Landmarks), formatChain(e.Landmarks),
		))
	}
	return changes
}

func (e editRouteInput) patch(step stepInput) stepInput {
	if e.Timeout != nil {
		step.Timeout = e.Timeout
	}
	if e.Delay != nil {
		step.Delay = e.Delay
	}
	if len(e.Landmarks) > 0 {
		step.Landmarks = e.Landmarks
	}
	return step
}

func routeStepIndex(steps []routeStep, id int) int {
	for index, step := range steps {
		if step.ID == id {
			return index
		}
	}
	return -1
}

func routeInputs(steps []routeStep) []stepInput {
	inputs := make([]stepInput, 0, len(steps))
	for _, step := range steps {
		inputs = append(inputs, step.toInput())
	}
	return inputs
}
