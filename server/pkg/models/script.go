package models

import (
	"encoding/json"
	"fmt"
)

// Script flags
const (
	EventOnTemplate   = 1 << 0
	EventOnText       = 1 << 1
	EventOnClass      = 1 << 2
	TemplateIsVisible = 1 << 3
	TextIsVisible     = 1 << 4
	ClassIsVisible    = 1 << 5
	TypeText          = 1 << 6
)

// Script ...
type Script struct {
	Name  string       `json:"name"`
	Steps []ScriptStep `json:"steps"`
}

// ToJSON ...
func (s *Script) ToJSON() []byte {
	result, err := json.Marshal(s)
	if err != nil {
		fmt.Println("Script ToJson " + err.Error())
		return []byte{}
	}
	return result
}

// ScriptStep ...
type ScriptStep struct {
	ID      int     `json:"id,omitempty"`
	Events  []Event `json:"events,omitempty"`
	Flags   int     `json:"flags,omitempty"`
	Text    string  `json:"text,omitempty"`
	Locale  string  `json:"locale,omitempty"`
	Command string  `json:"command,omitempty"`
}

func (s *ScriptStep) HasEventFlag() bool {
	return s != nil &&
		(s.HasFlag(EventOnTemplate) || s.HasFlag(EventOnText) || s.HasFlag(EventOnClass))
}

func (s *ScriptStep) HasVisibleFlag() bool {
	return s != nil &&
		(s.HasFlag(TemplateIsVisible) || s.HasFlag(TextIsVisible) || s.HasFlag(ClassIsVisible))
}

func (s *ScriptStep) HasAnyTemplateFlags() bool {
	return s != nil && (s.HasFlag(EventOnTemplate) || s.HasFlag(TemplateIsVisible))
}

func (s *ScriptStep) HasAnyTextFlags() bool {
	return s != nil && (s.HasFlag(EventOnText) || s.HasFlag(TextIsVisible))
}

func (s *ScriptStep) HasAnyYoloFlags() bool {
	return s != nil && (s.HasFlag(EventOnClass) || s.HasFlag(ClassIsVisible))
}

func (s *ScriptStep) HasFlag(flag int) bool {
	return s != nil && s.Flags&flag != 0
}

// Event ...
type Event struct {
	Time int64        `json:"time"`
	Data ControlBytes `json:"data"`
}
