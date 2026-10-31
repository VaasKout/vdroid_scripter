package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) handlePing(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in emptyInput,
) (*mcp.CallToolResult, any, error) {
	err := s.api.pingServer()
	if err != nil {
		return nil, nil, err
	}
	return textResult("vdroid server is up at " + s.api.baseURL), nil, nil
}

func (s *Server) handleStopServer(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in emptyInput,
) (*mcp.CallToolResult, any, error) {
	stopped, err := s.api.stopServer()
	if err != nil {
		return nil, nil, err
	}
	if !stopped {
		return textResult("vdroid server is not running"), nil, nil
	}
	return textResult("vdroid server stopped"), nil, nil
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
	return textResult(formatDevices(devices)), nil, nil
}

func formatDevices(devices []deviceInfo) string {
	if len(devices) == 0 {
		return "no devices connected"
	}

	var lines strings.Builder
	fmt.Fprintf(
		&lines,
		"%d devices; columns: serial model, brand device, Android version, locale, model_os",
		len(devices),
	)
	for _, device := range devices {
		fmt.Fprintf(
			&lines,
			"\n%s %s, %s %s, Android %s, locale %s, model_os %s",
			device.Serial, deviceName(device), device.Brand, device.Device,
			device.OsVersion, device.Locale, modelOsLabel(device),
		)
		if device.ScrcpyRunning {
			lines.WriteString(", session open")
		}
	}
	return lines.String()
}

func deviceName(device deviceInfo) string {
	if device.MarketingName != "" {
		return device.MarketingName
	}
	return device.Model
}

func modelOsLabel(device deviceInfo) string {
	if device.ModelOs == "" {
		return "unknown"
	}
	return device.ModelOs
}
