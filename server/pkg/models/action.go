package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Action ...
type Action struct {
	Name         string  `json:"name"`
	ScreenWidth  int     `json:"screen_width"`
	ScreenHeight int     `json:"screen_height"`
	Events       []Event `json:"events"`
}

// ToJSON ...
func (a *Action) ToJSON() []byte {
	result, err := json.Marshal(a)
	if err != nil {
		fmt.Println("Action ToJSON " + err.Error())
		return []byte{}
	}
	return result
}

// IsEmpty ...
func (a *Action) IsEmpty() bool {
	return a == nil || strings.TrimSpace(a.Name) == "" || len(a.Events) == 0
}
