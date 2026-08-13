package models

import (
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

type Step struct {
	Event   string `json:"event"`
	Type    string `json:"type"`
	Value   string `json:"value"`
	Locale  string `json:"locale,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

func (s *Step) GetTimeout() time.Duration {
	if s == nil || s.Timeout <= 0 {
		return time.Duration(DefaultTimeout) * time.Second
	}
	return time.Duration(s.Timeout) * time.Second
}

func (s *Step) Valid() bool {
	if s == nil {
		return false
	}

	switch s.Event {
	case TapEvent, LongTapEvent:
		return s.HasType()
	case TypeTextEvent:
		return strings.TrimSpace(s.Value) != ""
	}
	if s.IsCheckEvent() {
		return s.HasType()
	}
	return strings.TrimSpace(s.Type) == "" || s.HasType()
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

func (s *Step) HasType() bool {
	if s == nil {
		return false
	}

	switch s.Type {
	case Image, Text, Yolo:
		return strings.TrimSpace(s.Value) != ""
	}
	return false
}

func (s *Step) ToString() string {
	if s == nil {
		return ""
	}
	if s.IsCheckEvent() {
		return fmt.Sprintf("check on %s %s", s.Type, s.Value)
	}
	if s.Event == TypeTextEvent {
		return fmt.Sprintf("%s %q", s.Event, s.Value)
	}
	if s.HasType() {
		return fmt.Sprintf("%s on %s %s", s.Event, s.Type, s.Value)
	}
	return s.Event
}
