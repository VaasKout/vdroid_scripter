// Package models ...
package models

import (
	"strings"
	"unicode"
)

// AdbDevice ...
type AdbDevice struct {
	Serial        string `json:"serial"`
	Brand         string `json:"brand"`
	Device        string `json:"device"`
	Locale        string `json:"locale"`
	Model         string `json:"model"`
	OsVersion     string `json:"os_version"`
	ModelOs       string `json:"model_os"`
	Manufacturer  string `json:"manufacturer"`
	MarketingName string `json:"marketing_name"`
	Density       int    `json:"density"`
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

func (a *AdbDevice) ToModelOs() string {
	if a == nil {
		return ""
	}
	model := libraryToken(a.Model)
	osVersion := libraryToken(a.OsVersion)
	if model == "" || osVersion == "" {
		return ""
	}
	return model + "_" + osVersion
}

func libraryToken(value string) string {
	var token strings.Builder
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			token.WriteRune(r)
			continue
		}
		token.WriteRune('_')
	}
	return strings.Trim(token.String(), "_")
}
