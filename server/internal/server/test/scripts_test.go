package test

import (
	"android_vision_scripter/internal/server"
	"android_vision_scripter/internal/usecases"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// Script Path constants...
const (
	LocationsPath     = LocalURL + server.Locations
	ScriptsPath       = LocalURL + server.Scripts
	ScriptsByNamePath = LocalURL + server.ScriptsByName
	SaveScriptPath    = LocalURL + server.SaveScript
	SaveRectangle     = LocalURL + server.SaveRectangle
	RunScriptsPath    = LocalURL + server.RunScripts
	FindTextPath      = LocalURL + server.FindText
)

// Script params constants
const (
	TestScript   = "test_script"
	TestLocation = "main_screen"
)

var testScript1 = models.Script{
	Name:         TestScript + "1",
	Location:     TestLocation,
	NextLocation: []string{"profile"},
	Params: []models.Parameter{
		{
			Type:  models.Text,
			Value: "test",
		},
	},
}

var testScript2 = models.Script{
	Name:         TestScript + "2",
	Location:     TestLocation,
	NextLocation: []string{"settings"},
	Params: []models.Parameter{
		{
			Type:  models.Text,
			Value: "settings",
		},
	},
}

var testScript3 = models.Script{
	Name:         TestScript + "3",
	Location:     "login",
	NextLocation: []string{TestLocation},
	Params: []models.Parameter{
		{
			Type:  models.Text,
			Value: "test",
		},
	},
}

func TestGetLocations(t *testing.T) {
	var data = ""
	makeHTTPRequest(
		LocationsPath,
		http.MethodGet,
		[]byte{},
		&data,
	)
	t.Log(data)
}

func TestGetScripts(t *testing.T) {
	var locationPath = fmt.Sprintf("{%s}", server.LocationKey)
	var url = strings.ReplaceAll(ScriptsPath, locationPath, TestLocation)

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
	var locationPath = fmt.Sprintf("{%s}", server.LocationKey)
	var namePath = fmt.Sprintf("{%s}", server.NameKey)
	var url = strings.ReplaceAll(ScriptsByNamePath, locationPath, TestLocation)
	url = strings.ReplaceAll(url, namePath, TestScript+"1")

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodGet,
		[]byte{},
		&data,
	)
	t.Log(data)
}

func TestDeleteLocation(t *testing.T) {
	var locationPath = fmt.Sprintf("{%s}", server.LocationKey)
	var url = strings.ReplaceAll(ScriptsPath, locationPath, TestLocation)

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodDelete,
		[]byte{},
		&data,
	)
	t.Log(data)
}

func TestDeleteScript(t *testing.T) {
	var locationPath = fmt.Sprintf("{%s}", server.LocationKey)
	var namePath = fmt.Sprintf("{%s}", server.NameKey)
	var url = strings.ReplaceAll(ScriptsByNamePath, locationPath, TestLocation)
	url = strings.ReplaceAll(url, namePath, TestScript+"1")

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

func TestSaveScript(t *testing.T) {
	var testObjs = []models.Script{
		testScript1, testScript2, testScript3,
	}

	for _, body := range testObjs {
		bytes, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		var data = ""

		makeHTTPRequest(
			SaveScriptPath,
			http.MethodPost,
			bytes,
			&data,
		)
		t.Log(data)

	}
}

func TestRunScript(t *testing.T) {
	var body = usecases.RunScriptsDto{
		Serial: TestSerial,
		Scripts: []usecases.ScriptRef{
			{Location: TestLocation, Name: TestScript + "1"},
		},
	}
	bytes, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	var data = ""
	makeHTTPRequest(
		RunScriptsPath,
		http.MethodPost,
		bytes,
		&data,
	)
	t.Log(data)
}

func TestFindText(t *testing.T) {
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var url = strings.ReplaceAll(FindTextPath, serialPath, TestSerial)
	url = fmt.Sprintf("%s?%s=%s", url, server.TextKey, "Contacts")

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodGet,
		nil,
		&data,
	)
	t.Log(data)
}
