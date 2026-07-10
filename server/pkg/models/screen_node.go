package models

import (
	"encoding/json"
	"fmt"
)

const (
	OcrTextKey   = "ocr_text"
	YoloLabelKey = "yolo_label"
	TemplateKey  = "template"
)

// ScreenNode ...
type ScreenNode struct {
	Name    string            `json:"name"`
	Device  string            `json:"device"`
	Anchors map[string]string `json:"anchors,omitempty"`
	Actions []Action          `json:"actions,omitempty"`
}

// Action ...
type Action struct {
	NextNode string `json:"next_node,omitempty"`
	Script   string `json:"script"`
}

// ToJSON ...
func (n *ScreenNode) ToJSON() []byte {
	result, err := json.Marshal(n)
	if err != nil {
		fmt.Println("ScreenNode ToJSON " + err.Error())
		return []byte{}
	}
	return result
}

// FromJSON ...
func (n *ScreenNode) FromJSON(data []byte) error {
	return json.Unmarshal(data, n)
}
