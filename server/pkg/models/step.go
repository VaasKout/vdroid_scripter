package models

import (
	"android_vision_scripter/pkg/core/file"
	"fmt"
	"strings"
	"time"
)

const DefaultTimeout int = 15

const (
	TapEvent      = "tap"
	LongTapEvent  = "long_tap"
	TypeTextEvent = "type_text"
)

const (
	Image = "image"
	Text  = "text"
	Yolo  = "yolo"
)

type Anchor struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Locale string `json:"locale,omitempty"`
}

func (a *Anchor) Valid() bool {
	if a == nil {
		return false
	}
	if a.Type == Image && !file.ValidName(a.Value) {
		return false
	}

	switch a.Type {
	case Image, Text, Yolo:
		return strings.TrimSpace(a.Value) != ""
	}
	return false
}

func (a *Anchor) ToString() string {
	if a == nil {
		return ""
	}
	return fmt.Sprintf("%s %s", a.Type, a.Value)
}

type Step struct {
	Event    string   `json:"event"`
	Anchors  []Anchor `json:"anchors,omitempty"`
	Timeout  int      `json:"timeout,omitempty"`
	NextNode string   `json:"-"`
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
		return s.ValidAnchors()
	case TypeTextEvent:
		last := s.LastAnchor()
		return last != nil && strings.TrimSpace(last.Value) != ""
	}
	if s.IsCheckEvent() {
		return s.ValidAnchors()
	}
	return len(s.Anchors) == 0 || s.ValidAnchors()
}

func (s *Step) IsCheckEvent() bool {
	return s != nil && strings.TrimSpace(s.Event) == ""
}

func (s *Step) IsCustomEvent() bool {
	if s == nil || s.IsCheckEvent() {
		return false
	}

	switch s.Event {
	case TapEvent, LongTapEvent, TypeTextEvent:
		return false
	}
	return true
}

func (s *Step) ValidAnchors() bool {
	if s == nil || len(s.Anchors) == 0 {
		return false
	}
	for _, anchor := range s.Anchors {
		if !anchor.Valid() {
			return false
		}
	}
	return true
}

func (s *Step) LastAnchor() *Anchor {
	if s == nil || len(s.Anchors) == 0 {
		return nil
	}
	return &s.Anchors[len(s.Anchors)-1]
}

func (s *Step) AnchorsToString() string {
	if s == nil {
		return ""
	}
	parts := make([]string, 0, len(s.Anchors))
	for _, anchor := range s.Anchors {
		parts = append(parts, anchor.ToString())
	}
	return strings.Join(parts, " -> ")
}

func (s *Step) ToString() string {
	if s == nil {
		return ""
	}
	if s.IsCheckEvent() {
		return fmt.Sprintf("check on %s", s.AnchorsToString())
	}
	if s.Event == TypeTextEvent {
		last := s.LastAnchor()
		if last == nil {
			return s.Event
		}
		return fmt.Sprintf("%s %q", s.Event, last.Value)
	}
	if s.ValidAnchors() {
		return fmt.Sprintf("%s on %s", s.Event, s.AnchorsToString())
	}
	return s.Event
}
