package test

import (
	"android_vision_scripter/internal/server"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

const (
	GetKeyboardPath   = LocalURL + server.GetKeyboard
	EditKeyboardPath  = LocalURL + server.EditKeyboard
	ResetKeyboardPath = LocalURL + server.ResetKeyboard
	DeleteButtonPath  = LocalURL + server.DeleteButton
)

func TestGetKeyboard(t *testing.T) {
	var data = ""
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var localeKey = fmt.Sprintf("%s=eng", server.LocaleKey)
	var url = strings.ReplaceAll(GetKeyboardPath, serialPath, TestSerial)
	url = fmt.Sprintf("%s?%s", url, localeKey)
	t.Log(url)

	makeHTTPRequest(
		url,
		http.MethodGet,
		nil,
		&data,
	)
	t.Log(data)
}

func TestEditKeyboard(t *testing.T) {
	var data = ""
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var url = strings.ReplaceAll(EditKeyboardPath, serialPath, TestSerial)
	t.Log(url)

	var body = struct {
		Serial    string           `json:"serial"`
		Locale    string           `json:"locale"`
		Name      string           `json:"name"`
		Rectangle models.Rectangle `json:"rectangle"`
	}{
		Serial: TestSerial,
		Locale: "en",
		Name:   "space",
		Rectangle: models.Rectangle{
			LeftX:   500,
			TopY:    1000,
			RightX:  1000,
			BottomY: 1500,
		},
	}

	bytes, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	makeHTTPRequest(
		url,
		http.MethodPost,
		bytes,
		&data,
	)
	t.Log(data)
}

func TestResetKeyboard(t *testing.T) {
	var data = ""
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var localeKey = fmt.Sprintf("%s=eng", server.LocaleKey)
	var upperCase = fmt.Sprintf("%s=false", server.UpperCaseKey)
	var url = strings.ReplaceAll(ResetKeyboardPath, serialPath, TestSerial)
	url = fmt.Sprintf("%s?%s&%s", url, localeKey, upperCase)
	t.Log(url)

	makeHTTPRequest(
		url,
		http.MethodGet,
		nil,
		&data,
	)
	t.Log(data)
}

func TestDeleteButton(t *testing.T) {
	var data = ""
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var localeKey = fmt.Sprintf("%s=eng", server.LocaleKey)
	var nameKey = fmt.Sprintf("%s=q", server.NameKey)
	var url = strings.ReplaceAll(DeleteButtonPath, serialPath, TestSerial)
	url = fmt.Sprintf("%s?%s&%s", url, localeKey, nameKey)
	t.Log(url)

	makeHTTPRequest(
		url,
		http.MethodGet,
		nil,
		&data,
	)
	t.Log(data)
}
