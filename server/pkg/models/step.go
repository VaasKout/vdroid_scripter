package models

import (
	"android_vision_scripter/pkg/core/file"
	"fmt"
	"strings"
	"time"
)

// Timeouts in seconds
const (
	FirstFrameTimeout int = 15
)

// Landmark types
const (
	Image = "image"
	Text  = "text"
	Yolo  = "yolo"
)

// Landmark ...
type Landmark struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Locale string `json:"locale,omitempty"`
}

// Valid ...
func (l *Landmark) Valid() bool {
	if l == nil {
		return false
	}
	if l.Type == Image && !file.ValidName(l.Value) {
		return false
	}

	switch l.Type {
	case Image, Text, Yolo:
		return strings.TrimSpace(l.Value) != ""
	}
	return false
}

// ToString ...
func (l *Landmark) ToString() string {
	if l == nil {
		return ""
	}
	return fmt.Sprintf("%s %s", l.Type, l.Value)
}

// Step ...
type Step struct {
	ID        int        `json:"id,omitempty"`
	Event     string     `json:"event"`
	Landmarks []Landmark `json:"landmarks,omitempty"`
	Timeout   int        `json:"timeout,omitempty"`
	Delay     int        `json:"delay,omitempty"`
}

// GetTimeout ...
func (s *Step) GetTimeout() time.Duration {
	if s == nil || s.Timeout <= 0 {
		return 0
	}
	return time.Duration(s.Timeout) * time.Millisecond
}

// GetDelay ...
func (s *Step) GetDelay() time.Duration {
	if s == nil || s.Delay <= 0 {
		return 0
	}
	return time.Duration(s.Delay) * time.Millisecond
}

// FillStepIDs ...
func FillStepIDs(steps []Step) {
	for index := range steps {
		if steps[index].ID > 0 {
			continue
		}
		steps[index].ID = index + 1
	}
}

// ClearStartDelay ...
func ClearStartDelay(steps []Step) {
	if len(steps) == 0 {
		return
	}
	steps[0].Delay = 0
}

// ValidQueue ...
func ValidQueue(steps []Step) bool {
	if len(steps) == 0 {
		return false
	}
	for _, step := range steps {
		if !step.Valid() {
			return false
		}
	}
	return true
}

// Valid ...
func (s *Step) Valid() bool {
	if s == nil || s.Timeout < 0 || s.Delay < 0 {
		return false
	}
	if s.IsCustomEvent() && !file.ValidName(s.Event) {
		return false
	}

	switch s.Event {
	case TapEvent, LongTapEvent:
		return s.ValidLandmarks()
	case TypeTextEvent:
		last := s.LastLandmark()
		if last == nil {
			return false
		}
		return strings.TrimSpace(last.Value) != ""
	}
	if s.IsCheckEvent() {
		return s.ValidLandmarks()
	}
	return len(s.Landmarks) == 0 || s.ValidLandmarks()
}

// IsCheckEvent ...
func (s *Step) IsCheckEvent() bool {
	return s != nil && strings.TrimSpace(s.Event) == ""
}

// IsCustomEvent ...
func (s *Step) IsCustomEvent() bool {
	if s == nil || s.IsCheckEvent() || s.IsSwipeEvent() {
		return false
	}

	switch s.Event {
	case TapEvent, LongTapEvent, TypeTextEvent:
		return false
	}
	return true
}

// IsSwipeEvent ...
func (s *Step) IsSwipeEvent() bool {
	if s == nil {
		return false
	}

	switch s.Event {
	case SwipeUpEvent, SwipeDownEvent, SwipeLeftEvent, SwipeRightEvent:
		return true
	}
	return false
}

// ValidLandmarks ...
func (s *Step) ValidLandmarks() bool {
	if s == nil || len(s.Landmarks) == 0 {
		return false
	}
	for _, landmark := range s.Landmarks {
		if !landmark.Valid() {
			return false
		}
	}
	return true
}

// LastLandmark ...
func (s *Step) LastLandmark() *Landmark {
	if s == nil || len(s.Landmarks) == 0 {
		return nil
	}
	return &s.Landmarks[len(s.Landmarks)-1]
}

// LandmarksToString ...
func (s *Step) LandmarksToString() string {
	if s == nil {
		return ""
	}
	parts := make([]string, 0, len(s.Landmarks))
	for _, landmark := range s.Landmarks {
		parts = append(parts, landmark.ToString())
	}
	return strings.Join(parts, " -> ")
}

// ToString ...
func (s *Step) ToString() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("step %d: %s", s.ID, s.description())
}

func (s *Step) description() string {
	if s.IsCheckEvent() {
		return fmt.Sprintf("check on %s", s.LandmarksToString())
	}
	if s.Event == TypeTextEvent {
		last := s.LastLandmark()
		if last == nil {
			return s.Event
		}
		return fmt.Sprintf("%s %q", s.Event, last.Value)
	}
	if s.ValidLandmarks() {
		return fmt.Sprintf("%s on %s", s.Event, s.LandmarksToString())
	}
	return s.Event
}

// ArgsCount ...
func (l *Landmark) ArgsCount() int {
	if l == nil || l.Type != Text {
		return 0
	}
	return strings.Count(l.Value, RouteArgPlaceholder)
}

// ArgsCount ...
func (s *Step) ArgsCount() int {
	if s == nil {
		return 0
	}

	count := 0
	for _, landmark := range s.Landmarks {
		count += landmark.ArgsCount()
	}
	return count
}

// WithArgs ...
func (s Step) WithArgs(args []string, argsLocale string) Step {
	landmarks := make([]Landmark, 0, len(s.Landmarks))
	used := 0
	for _, landmark := range s.Landmarks {
		count := landmark.ArgsCount()
		if count > 0 {
			landmarkArgs := args[used : used+count]
			landmark = landmark.WithArgs(landmarkArgs, argsLocale)
		}
		landmarks = append(landmarks, landmark)
		used += count
	}
	s.Landmarks = landmarks
	return s
}

// WithArgs ...
func (l Landmark) WithArgs(args []string, argsLocale string) Landmark {
	l.Value = fillPlaceholders(l.Value, args)
	l.Locale = argsLocale
	return l
}

func fillPlaceholders(value string, args []string) string {
	parts := strings.Split(value, RouteArgPlaceholder)
	filled := &strings.Builder{}
	for index, part := range parts {
		filled.WriteString(part)
		if index < len(args) {
			filled.WriteString(args[index])
		}
	}
	return filled.String()
}
