package models

import (
	"strings"
	"unicode"
)

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

func (a *AdbDevice) Update(fresh AdbDevice) {
	a.Brand = firstNonEmpty(fresh.Brand, a.Brand)
	a.Device = firstNonEmpty(fresh.Device, a.Device)
	a.Locale = firstNonEmpty(fresh.Locale, a.Locale)
	a.Model = firstNonEmpty(fresh.Model, a.Model)
	a.OsVersion = firstNonEmpty(fresh.OsVersion, a.OsVersion)
	a.Manufacturer = firstNonEmpty(fresh.Manufacturer, a.Manufacturer)
	a.MarketingName = firstNonEmpty(fresh.MarketingName, a.MarketingName)
	if fresh.Density > 0 {
		a.Density = fresh.Density
	}
	a.ModelOs = a.ToModelOs()
}

func (a *AdbDevice) Ready() bool {
	return a.ModelOs != "" && a.Locale != "" && a.Density > 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
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
