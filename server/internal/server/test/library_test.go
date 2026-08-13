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
	LibraryPath       = LocalURL + server.Library
	SaveImagePath     = LocalURL + server.SaveImage
	SaveActionPath    = LocalURL + server.SaveAction
	ImagesByNamePath  = LocalURL + server.ImagesByName
	ActionsByNamePath = LocalURL + server.ActionsByName
)

const (
	TestImageName  = "test_image"
	TestActionName = "test_action"
)

func TestGetLibrary(t *testing.T) {
	var data = ""
	makeHTTPRequest(
		LibraryPath,
		http.MethodGet,
		[]byte{},
		&data,
	)
	t.Log(data)
}

func TestSaveImage(t *testing.T) {
	var body = struct {
		Serial    string           `json:"serial"`
		Rectangle models.Rectangle `json:"rectangle"`
	}{
		Serial: TestSerial,
		Rectangle: models.Rectangle{
			LeftX:   100,
			TopY:    200,
			RightX:  300,
			BottomY: 400,
			Label:   TestImageName,
		},
	}

	bytes, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	var data = ""
	makeHTTPRequest(
		SaveImagePath,
		http.MethodPost,
		bytes,
		&data,
	)
	t.Log(data)
}

func TestSaveAction(t *testing.T) {
	var body = models.Action{
		Name:         TestActionName,
		ScreenWidth:  1080,
		ScreenHeight: 2400,
		Events:       models.GenerateTapEvents(1080, 2400),
	}

	bytes, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	var data = ""
	makeHTTPRequest(
		SaveActionPath,
		http.MethodPost,
		bytes,
		&data,
	)
	t.Log(data)
}

func TestDeleteImage(t *testing.T) {
	var namePath = fmt.Sprintf("{%s}", server.NameKey)
	var url = strings.ReplaceAll(ImagesByNamePath, namePath, TestImageName)

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodDelete,
		[]byte{},
		&data,
	)
	t.Log(data)
}

func TestDeleteAction(t *testing.T) {
	var namePath = fmt.Sprintf("{%s}", server.NameKey)
	var url = strings.ReplaceAll(ActionsByNamePath, namePath, TestActionName)

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodDelete,
		[]byte{},
		&data,
	)
	t.Log(data)
}
