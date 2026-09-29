package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	modeNavigator = "navigator"
	modeExplorer  = "explorer"
)

type agentMode struct {
	mutex sync.Mutex
	name  string
	route string
}

func (m *agentMode) get() (string, string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.name == "" {
		return modeNavigator, ""
	}
	return m.name, m.route
}

func (m *agentMode) set(name string, route string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.name = name
	m.route = route
}

type setModeInput struct {
	Mode  string `json:"mode" jsonschema:"navigator | explorer"`
	Route string `json:"route,omitempty" jsonschema:"explorer: route to record into"`
}

func (s *Server) handleSetMode(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in setModeInput,
) (*mcp.CallToolResult, any, error) {
	switch in.Mode {
	case modeNavigator:
		s.mode.set(modeNavigator, "")
		return textResult("mode navigator: nothing is recorded"), nil, nil
	case modeExplorer:
		return s.enterExplorer(in.Route)
	}
	return nil, nil, fmt.Errorf("unknown mode %q: use navigator or explorer", in.Mode)
}

func (s *Server) enterExplorer(route string) (*mcp.CallToolResult, any, error) {
	if route == "" {
		return nil, nil, fmt.Errorf("explorer needs a route name to record into")
	}
	saved, err := s.api.savedRouteSteps(route)
	if err != nil {
		return nil, nil, err
	}
	s.mode.set(modeExplorer, route)
	var text = fmt.Sprintf("mode explorer, recording into route %s (%d steps so far)", route, len(saved))
	return textResult(text), nil, nil
}

func filterChecks(steps []stepInput, checks bool) []stepInput {
	kept := make([]stepInput, 0, len(steps))
	for _, step := range steps {
		if (step.Event == "") != checks {
			continue
		}
		kept = append(kept, step)
	}
	return kept
}
