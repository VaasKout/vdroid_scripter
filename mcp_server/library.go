package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getLibraryInput struct {
	Serial string `json:"serial,omitempty" jsonschema:"device serial: lists the common items and this device's variants, hiding other devices' variants"`
}

type saveImageInput struct {
	Serial         string `json:"serial" jsonschema:"device serial"`
	Name           string `json:"name" jsonschema:"<app>_<screen>_<what>[_variant]"`
	Left           int    `json:"left" jsonschema:"edges in capture pixels"`
	Top            int    `json:"top"`
	Right          int    `json:"right"`
	Bottom         int    `json:"bottom"`
	DeviceSpecific bool   `json:"device_specific,omitempty" jsonschema:"true only when the user wants the image for this device alone; the server appends _<model_os> to name"`
}

type libraryNameInput struct {
	Name string `json:"name" jsonschema:"library item name"`
}

func (s *Server) handleGetLibrary(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in getLibraryInput,
) (*mcp.CallToolResult, any, error) {
	library, err := s.api.getLibrary()
	if err != nil {
		return nil, nil, err
	}
	scope := libraryScope{}
	if in.Serial != "" {
		devices, err := s.api.getDevices()
		if err != nil {
			return nil, nil, err
		}
		scope = scopeFor(devices, in.Serial)
	}
	return textResult(formatLibrary(library, scope)), nil, nil
}

func formatNames(kind string, names []string) string {
	if len(names) == 0 {
		return kind + ": none"
	}
	return fmt.Sprintf("%s (%d): %s", kind, len(names), strings.Join(names, ", "))
}

func (s *Server) handleSaveImage(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in saveImageInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" || in.Name == "" {
		return nil, nil, fmt.Errorf("serial and name are required")
	}
	if in.Right <= in.Left || in.Bottom <= in.Top {
		return nil, nil, fmt.Errorf("rectangle must have right > left and bottom > top")
	}
	rectangle := saveImageRectangle{
		LeftX:   in.Left,
		RightX:  in.Right,
		TopY:    in.Top,
		BottomY: in.Bottom,
		Label:   in.Name,
	}
	name, err := s.api.saveImage(in.Serial, rectangle, in.DeviceSpecific)
	if err != nil {
		return nil, nil, err
	}
	return textResult("saved image " + name), nil, nil
}

func (s *Server) handleDeleteImage(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in libraryNameInput,
) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	err := s.api.deleteImage(in.Name)
	if err != nil {
		return nil, nil, err
	}
	return textResult("deleted image " + in.Name), nil, nil
}

func (s *Server) handleDeleteAction(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in libraryNameInput,
) (*mcp.CallToolResult, any, error) {
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	err := s.api.deleteAction(in.Name)
	if err != nil {
		return nil, nil, err
	}
	return textResult("deleted action " + in.Name), nil, nil
}

const modelOsSeparator = "_"

func (s *Server) libraryName(serial string, name string, deviceSpecific bool) (string, error) {
	if !deviceSpecific {
		return name, nil
	}
	devices, err := s.api.getDevices()
	if err != nil {
		return "", err
	}
	for _, device := range devices {
		if device.Serial != serial {
			continue
		}
		if device.ModelOs == "" {
			return "", fmt.Errorf(
				"model and Android version of %s are unknown, cannot name a device-specific item",
				serial,
			)
		}
		return deviceSpecificName(name, device.ModelOs), nil
	}
	return "", fmt.Errorf("device %s is not connected", serial)
}

func deviceSpecificName(name string, modelOs string) string {
	suffix := modelOsSeparator + modelOs
	if strings.HasSuffix(name, suffix) {
		return name
	}
	return name + suffix
}

type libraryScope struct {
	modelOs string
	others  []string
}

func scopeFor(devices []deviceInfo, serial string) libraryScope {
	scope := libraryScope{}
	for _, device := range devices {
		if device.Serial == serial {
			scope.modelOs = device.ModelOs
		}
	}
	for _, device := range devices {
		tag := device.ModelOs
		if tag == "" || tag == scope.modelOs || slices.Contains(scope.others, tag) {
			continue
		}
		scope.others = append(scope.others, tag)
	}
	return scope
}

type scopedNames struct {
	common []string
	device []string
	hidden int
}

func (s libraryScope) split(names []string) scopedNames {
	split := scopedNames{common: []string{}, device: []string{}}
	for _, name := range names {
		if common, ok := s.deviceVariant(name); ok {
			split.device = append(split.device, common+" ("+name+")")
			continue
		}
		if s.belongsToOther(name) {
			split.hidden++
			continue
		}
		split.common = append(split.common, name)
	}
	return split
}

func (s libraryScope) deviceVariant(name string) (string, bool) {
	if s.modelOs == "" {
		return "", false
	}
	common, ok := strings.CutSuffix(name, modelOsSeparator+s.modelOs)
	if !ok || common == "" {
		return "", false
	}
	return common, true
}

func (s libraryScope) belongsToOther(name string) bool {
	for _, other := range s.others {
		if strings.HasSuffix(name, modelOsSeparator+other) {
			return true
		}
	}
	return false
}

func formatLibrary(library libraryResponse, scope libraryScope) string {
	if scope.modelOs == "" && len(scope.others) == 0 {
		return formatNames("images", library.Images) +
			"\n" + formatNames("actions", library.Actions)
	}

	images := scope.split(library.Images)
	actions := scope.split(library.Actions)
	lines := []string{
		formatNames("images", images.common),
		formatNames("actions", actions.common),
	}
	if scope.modelOs != "" {
		lines = append(
			lines,
			formatNames("images for this device "+scope.modelOs, images.device),
			formatNames("actions for this device "+scope.modelOs, actions.device),
		)
	}
	if hidden := images.hidden + actions.hidden; hidden > 0 {
		lines = append(lines, fmt.Sprintf("other connected devices: %d item(s) hidden", hidden))
	}
	return strings.Join(lines, "\n")
}
