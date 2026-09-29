package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	modeDefault   = "default"
	modeExplorer  = "explorer"
	modeNavigator = "navigator"
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
		return modeDefault, ""
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
	Mode  string `json:"mode" jsonschema:"default | explorer | navigator"`
	Route string `json:"route,omitempty" jsonschema:"explorer only: route name to record into"`
}

func (s *Server) handleSetMode(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in setModeInput,
) (*mcp.CallToolResult, any, error) {
	switch in.Mode {
	case modeDefault:
		s.mode.set(modeDefault, "")
		return textResult("mode default"), nil, nil
	case modeNavigator:
		s.mode.set(modeNavigator, "")
		return textResult("mode navigator: saved routes only, fix them with edit_route"), nil, nil
	case modeExplorer:
		return s.enterExplorer(in.Route)
	}
	return nil, nil, fmt.Errorf("unknown mode %q: use default, explorer or navigator", in.Mode)
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

func (s *Server) refuseInNavigator(tool string) error {
	mode, _ := s.mode.get()
	if mode != modeNavigator {
		return nil
	}
	return navigatorRefusal(tool)
}

func navigatorRefusal(tool string) error {
	return fmt.Errorf(
		"%s is not available in navigator mode: follow saved routes with run_route "+
			"and fix them with edit_route, or ask the user to switch to explorer",
		tool,
	)
}

func withoutChecks(steps []stepInput) []stepInput {
	kept := make([]stepInput, 0, len(steps))
	for _, step := range steps {
		if step.Event == "" {
			continue
		}
		kept = append(kept, step)
	}
	return kept
}
