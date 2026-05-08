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

// Script Path constants...
const (
	ScriptsPath       = LocalURL + server.Scripts
	ScriptsByNamePath = LocalURL + server.ScriptsByName
	SaveStepPath      = LocalURL + server.SaveStep
	SaveRectangle     = LocalURL + server.SaveRectangle
	RunScriptPath     = LocalURL + server.RunScript
	FindTextPath      = LocalURL + server.FindText
)

// Script params constants
const (
	TestScript = "test_script"
)

// Script default data
var scriptData = make([]byte, 32)

func TestGetScripts(t *testing.T) {
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var url = strings.ReplaceAll(ScriptsPath, serialPath, TestSerial)
	var data = ""
	makeHTTPRequest(
		url,
		http.MethodGet,
		[]byte{},
		&data,
	)
	t.Log(data)
}

func TestGetScriptByName(t *testing.T) {
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var namePath = fmt.Sprintf("{%s}", server.NameKey)
	var url = strings.ReplaceAll(ScriptsByNamePath, serialPath, TestSerial)
	url = strings.ReplaceAll(url, namePath, TestScript)

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodGet,
		[]byte{},
		&data,
	)
	t.Log(data)
}

func TestDeleteScript(t *testing.T) {
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var namePath = fmt.Sprintf("{%s}", server.NameKey)
	var url = strings.ReplaceAll(ScriptsByNamePath, serialPath, TestSerial)
	url = strings.ReplaceAll(url, namePath, TestScript)

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodDelete,
		[]byte{},
		&data,
	)
	t.Log(data)
}

func TestAddStepZone(t *testing.T) {
	var body = struct {
		Serial string           `json:"serial"`
		Zone   models.Rectangle `json:"zone"`
	}{
		Serial: TestSerial,
		Zone: models.Rectangle{
			LeftX:   500,
			TopY:    500,
			RightX:  700,
			BottomY: 700,
		},
	}

	bytes, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	var data = map[string]any{}
	makeHTTPRequest(
		SaveRectangle,
		http.MethodPost,
		bytes,
		&data,
	)
	t.Log(data)
}

func TestSaveStep(t *testing.T) {
	var body = struct {
		Serial string            `json:"serial"`
		Name   string            `json:"name"`
		Step   models.ScriptStep `json:"step"`
	}{
		Serial: TestSerial,
		Name:   TestScript,
		Step: models.ScriptStep{
			Text: "test",
		},
	}
	bytes, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	var data = ""

	makeHTTPRequest(
		SaveStepPath,
		http.MethodPost,
		bytes,
		&data,
	)
	t.Log(data)
}

func TestRunScript(t *testing.T) {
	var data = ""
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var namePath = fmt.Sprintf("{%s}", server.NameKey)
	var url = strings.ReplaceAll(RunScriptPath, serialPath, TestSerial)
	url = strings.ReplaceAll(url, namePath, TestScript)
	fmt.Println(url)

	makeHTTPRequest(
		url,
		http.MethodGet,
		nil,
		&data,
	)
	t.Log(data)
}

func TestFindText(t *testing.T) {
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var url = strings.ReplaceAll(FindTextPath, serialPath, TestSerial)
	url = fmt.Sprintf("%s?%s=%s", url, server.TextKey, "Контакты")

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodGet,
		nil,
		&data,
	)
	t.Log(data)
}
