package models

import (
	"android_vision_scripter/pkg/core/file"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultTimeout    int = 3
	FirstFrameTimeout int = 15
)

const (
	TapEvent      = "tap"
	LongTapEvent  = "long_tap"
	TypeTextEvent = "type_text"

	SwipeUpEvent    = "swipe_up"
	SwipeDownEvent  = "swipe_down"
	SwipeLeftEvent  = "swipe_left"
	SwipeRightEvent = "swipe_right"
)

const (
	Image = "image"
	Text  = "text"
	Yolo  = "yolo"
)

type Landmark struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Locale string `json:"locale,omitempty"`
}

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

func (l *Landmark) ToString() string {
	if l == nil {
		return ""
	}
	return fmt.Sprintf("%s %s", l.Type, l.Value)
}

type Step struct {
	Event     string     `json:"event"`
	Landmarks []Landmark `json:"landmarks,omitempty"`
	Timeout   int        `json:"timeout,omitempty"`
}

func (s *Step) GetTimeout() time.Duration {
	if s == nil || s.Timeout <= 0 {
		return time.Duration(DefaultTimeout) * time.Second
	}
	return time.Duration(s.Timeout) * time.Second
}

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

func (s *Step) Valid() bool {
	if s == nil {
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
		return last != nil && strings.TrimSpace(last.Value) != ""
	}
	if s.IsCheckEvent() {
		return s.ValidLandmarks()
	}
	return len(s.Landmarks) == 0 || s.ValidLandmarks()
}

func (s *Step) IsCheckEvent() bool {
	return s != nil && strings.TrimSpace(s.Event) == ""
}

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

func (s *Step) LastLandmark() *Landmark {
	if s == nil || len(s.Landmarks) == 0 {
		return nil
	}
	return &s.Landmarks[len(s.Landmarks)-1]
}

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

func (s *Step) ToString() string {
	if s == nil {
		return ""
	}
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
