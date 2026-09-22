// Package models ...
package models

import (
	"strings"
)

// AdbDevice ...
type AdbDevice struct {
	Serial        string `json:"serial"`
	Brand         string `json:"brand"`
	Device        string `json:"device"`
	Locale        string `json:"locale"`
	Model         string `json:"model"`
	OsVersion     string `json:"os_version"`
	Manufacturer  string `json:"manufacturer"`
	MarketingName string `json:"marketing_name"`
	ScrcpyRunning bool   `json:"scrcpy_running"`
}

// FilterLocales ...
func (a *AdbDevice) FilterLocales() {
	if a == nil || a.Locale == "" {
		return
	}
	var systemLocales = strings.Split(a.Locale, ",")
	if len(systemLocales) == 0 {
		return
	}
	a.Locale = systemLocales[0]
}
