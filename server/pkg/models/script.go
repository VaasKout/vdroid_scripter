package models

import (
	"encoding/json"
	"fmt"
)

// Script flags
const (
	EventOnTemplate   = 1 << 0
	EventOnText       = 1 << 1
	TypeText          = 1 << 2
	TemplateIsVisible = 1 << 3
	TextIsVisible     = 1 << 4
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
	Command string  `json:"command,omitempty"`
}

// Event ...
type Event struct {
	Time int64        `json:"time"`
	Data ControlBytes `json:"data"`
}
