package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Type consts
const (
	Template  = "template"
	Text      = "text"
	TypeText  = "type_text"
	YoloClass = "yolo_class"
	Command   = "command"
)

// Type of strategies
const (
	OnFailure = "on_failure"
)

// Hardcoded names of actions
const (
	InitialScript = "init"
)

// DefaultTimeout ...
const DefaultTimeout int = 15

// Actions
const (
	ActionDown byte = 0
	ActionUp   byte = 1
)

// Script ...
type Script struct {
	Name         string      `json:"name"`
	Location     string      `json:"location"`
	NextLocation []string    `json:"next_location,omitempty"`
	Params       []Parameter `json:"params,omitempty"`
	Events       []Event     `json:"events,omitempty"`
	Timeout      int         `json:"timeout"`
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

// GetTimeout ...
func (s *Script) GetTimeout() time.Duration {
	if s == nil || s.Timeout <= 0 {
		return time.Duration(DefaultTimeout) * time.Second
	}
	return time.Duration(s.Timeout) * time.Second
}

func (s *Script) IsEmpty() bool {
	return s == nil || s.Name == "" || (len(s.Events) == 0 && len(s.Params) == 0)
}

// Parameter ...
type Parameter struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Locale string `json:"locale,omitempty"`
}

// Event ...
type Event struct {
	Time int64        `json:"time"`
	Data ControlBytes `json:"data"`
}

func ExtractTapEvents(events []Event) []Event {
	for index := 0; index < len(events)-1; index++ {
		downData := events[index].Data
		upData := events[index+1].Data
		if len(downData) != ControlBytesSize || len(upData) != ControlBytesSize {
			continue
		}
		if downData[1] == ActionDown && upData[1] == ActionUp {
			return []Event{events[index], events[index+1]}
		}
	}
	return events
}
