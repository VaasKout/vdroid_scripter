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

// Screen node path constants...
const (
	ScreenNodesPath       = LocalURL + server.ScreenNodes
	ScreenNodesByNamePath = LocalURL + server.ScreenNodesByName
)

// Screen node params constants
const (
	TestScreenNode = "test_node"
)

func TestSaveScreenNode(t *testing.T) {
	var node = struct {
		Serial  string            `json:"serial"`
		Name    string            `json:"name"`
		Anchors map[string]string `json:"anchors"`
		Actions []models.Action   `json:"actions"`
	}{
		Serial: TestSerial,
		Name:   TestScreenNode,
		Anchors: map[string]string{
			models.TemplateKey:  "1.png",
			models.YoloLabelKey: "home",
		},
		Actions: []models.Action{
			{
				NextNode: "youtube_main",
				Script:   "open_youtube",
			},
		},
	}

	bytes, err := json.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}

	var data = ""
	makeHTTPRequest(
		ScreenNodesPath,
		http.MethodPost,
		bytes,
		&data,
	)
	t.Log(data)
}

func TestGetScreenNodes(t *testing.T) {
	var namePath = fmt.Sprintf("{%s}", server.NameKey)
	var url = strings.ReplaceAll(ScreenNodesByNamePath, namePath, TestScreenNode)

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodGet,
		[]byte{},
		&data,
	)
	t.Log(data)

	data = ""
	makeHTTPRequest(
		ScreenNodesPath,
		http.MethodGet,
		[]byte{},
		&data,
	)
	t.Log(data)

}

func TestDeleteScreenNode(t *testing.T) {
	var namePath = fmt.Sprintf("{%s}", server.NameKey)
	var url = strings.ReplaceAll(ScreenNodesByNamePath, namePath, TestScreenNode)

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodDelete,
		[]byte{},
		&data,
	)
	t.Log(data)
}
