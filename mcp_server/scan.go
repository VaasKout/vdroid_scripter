package main

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const captureMimeType = "image/jpeg"

type scanInput struct {
	Serial string   `json:"serial" jsonschema:"device serial"`
	Images []string `json:"images,omitempty" jsonschema:"library image names to look for"`
	Locale string   `json:"locale,omitempty" jsonschema:"Tesseract code for the OCR, default eng"`
}

func (s *Server) handleScan(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in scanInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" {
		return nil, nil, fmt.Errorf("serial is required")
	}
	landmarks, err := s.api.scan(in.Serial, in.Images, in.Locale)
	if err != nil {
		return nil, nil, err
	}
	return textResult(formatLandmarks(landmarks, in.Locale)), nil, nil
}

func formatLandmarks(landmarks []scanLandmark, requestedLocale string) string {
	locale := textLocale(landmarks, requestedLocale)
	kept, dropped := confidentLandmarks(landmarks)
	if len(kept) == 0 {
		return fmt.Sprintf("no landmarks found (text locale %s%s)", locale, droppedNote(dropped))
	}

	var lines strings.Builder
	fmt.Fprintf(
		&lines,
		"%d landmarks in reading order, text locale %s%s; columns: type left,top,right,bottom value",
		len(kept), locale, droppedNote(dropped),
	)
	for _, landmark := range kept {
		rect := landmark.Rectangle
		fmt.Fprintf(
			&lines,
			"\n%s %d,%d,%d,%d %s",
			landmark.Type, rect.LeftX, rect.TopY, rect.RightX, rect.BottomY, landmark.Value,
		)
	}
	return lines.String()
}

func confidentLandmarks(landmarks []scanLandmark) ([]scanLandmark, int) {
	kept := make([]scanLandmark, 0, len(landmarks))
	dropped := 0
	for _, landmark := range landmarks {
		if lowConfidence(landmark) {
			dropped++
			continue
		}
		kept = append(kept, landmark)
	}
	return kept, dropped
}

func lowConfidence(landmark scanLandmark) bool {
	if landmark.Type != "text" || landmark.Confidence == nil {
		return false
	}
	return *landmark.Confidence < scanMinConfidence
}

func droppedNote(dropped int) string {
	if dropped == 0 {
		return ""
	}
	return fmt.Sprintf(", %d text entries below confidence %d dropped", dropped, scanMinConfidence)
}

func textLocale(landmarks []scanLandmark, requestedLocale string) string {
	for _, landmark := range landmarks {
		if landmark.Locale != "" {
			return landmark.Locale
		}
	}
	if requestedLocale != "" {
		return requestedLocale
	}
	return "eng"
}

func (s *Server) handleCapture(
	ctx context.Context,
	req *mcp.CallToolRequest,
	in serialInput,
) (*mcp.CallToolResult, any, error) {
	if in.Serial == "" {
		return nil, nil, fmt.Errorf("serial is required")
	}
	data, err := s.api.capture(in.Serial)
	if err != nil {
		return nil, nil, err
	}
	config, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	var text = fmt.Sprintf("frame %dx%d pixels, same coordinates as scan", config.Width, config.Height)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
			&mcp.ImageContent{Data: data, MIMEType: captureMimeType},
		},
	}, nil, nil
}
