package models

import (
	"fmt"
	"strings"
	"time"
)

const DefaultTimeout int = 15

const (
	StepTap      = "tap"
	StepLongTap  = "long_tap"
	StepTypeText = "type_text"
	StepCheck    = "check"
)

const (
	TargetImage = "image"
	TargetText  = "text"
	TargetYolo  = "yolo"
)

type Step struct {
	Action  string      `json:"action"`
	Target  *StepTarget `json:"target,omitempty"`
	Text    string      `json:"text,omitempty"`
	Locale  string      `json:"locale,omitempty"`
	Timeout int         `json:"timeout,omitempty"`
}

type StepTarget struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Locale string `json:"locale,omitempty"`
}

func (s *Step) GetTimeout() time.Duration {
	if s == nil || s.Timeout <= 0 {
		return time.Duration(DefaultTimeout) * time.Second
	}
	return time.Duration(s.Timeout) * time.Second
}

func (s *Step) Valid() bool {
	if s == nil || strings.TrimSpace(s.Action) == "" {
		return false
	}

	switch s.Action {
	case StepTap, StepLongTap, StepCheck:
		return s.Target.Valid()
	case StepTypeText:
		return strings.TrimSpace(s.Text) != "" &&
			(s.Target == nil || s.Target.Valid())
	}
	return s.Target == nil || s.Target.Valid()
}

func (s *Step) IsCustomAction() bool {
	switch s.Action {
	case StepTap, StepLongTap, StepTypeText, StepCheck:
		return false
	}
	return true
}

func (s *Step) Describe() string {
	if s == nil {
		return ""
	}
	if s.Action == StepTypeText {
		return fmt.Sprintf("%s %q", s.Action, s.Text)
	}
	if s.Target != nil {
		return fmt.Sprintf("%s on %s %s", s.Action, s.Target.Type, s.Target.Value)
	}
	return s.Action
}

func (t *StepTarget) Valid() bool {
	if t == nil {
		return false
	}

	switch t.Type {
	case TargetImage, TargetText, TargetYolo:
		return strings.TrimSpace(t.Value) != ""
	}
	return false
}
