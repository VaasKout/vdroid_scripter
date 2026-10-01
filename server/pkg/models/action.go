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

// FitScreen ...
func (a *Action) FitScreen(screenWidth int, screenHeight int) {
	if a == nil || !a.recordedOnOtherScreen(screenWidth, screenHeight) {
		return
	}

	for _, event := range a.Events {
		event.Data.rescale(a.ScreenWidth, a.ScreenHeight, screenWidth, screenHeight)
	}
	a.ScreenWidth = screenWidth
	a.ScreenHeight = screenHeight
}

func (a *Action) recordedOnOtherScreen(screenWidth int, screenHeight int) bool {
	if a.ScreenWidth <= 0 || a.ScreenHeight <= 0 || screenWidth <= 0 || screenHeight <= 0 {
		return false
	}
	return a.ScreenWidth != screenWidth || a.ScreenHeight != screenHeight
}
