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
	RunFlowPath = LocalURL + server.RunFlow
)

const (
	TestFlowNode   = "test_node_"
	TestFlowScript = "test_script_"
)

func newFlowScript(suffix string, node string, nextNode string) models.Script {
	return models.Script{
		Name:     TestFlowScript + suffix,
		Node:     node,
		NextNode: nextNode,
		Params: []models.Parameter{
			{
				Type:  models.Text,
				Value: "test_" + suffix,
			},
		},
	}
}

var flowScripts = []models.Script{
	newFlowScript("1_1", TestFlowNode+"1", TestFlowNode+"2"),
	newFlowScript("1_2", TestFlowNode+"1", TestFlowNode+"2"),
	newFlowScript("2_1", TestFlowNode+"2", TestFlowNode+"3"),
	newFlowScript("2_2", TestFlowNode+"2", TestFlowNode+"3"),
	newFlowScript("3_1", TestFlowNode+"3", ""),
	newFlowScript("3_2", TestFlowNode+"3", ""),
}

func TestRunFlow(t *testing.T) {
	var nodePath = fmt.Sprintf("{%s}", server.NodeKey)
	for num := 1; num <= 3; num++ {
		var node = fmt.Sprintf("%s%d", TestFlowNode, num)
		var url = strings.ReplaceAll(ScriptsPath, nodePath, node)

		var data = ""
		makeHTTPRequest(
			url,
			http.MethodDelete,
			[]byte{},
			&data,
		)
		t.Log(data)
	}

	for _, script := range flowScripts {
		bytes, err := json.Marshal(script)
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

	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var url = strings.ReplaceAll(RunFlowPath, serialPath, TestSerial)
	url = fmt.Sprintf(
		"%s?%s=%s&%s=%s",
		url,
		server.StartKey, TestFlowNode+"1",
		server.EndKey, TestFlowNode+"3",
	)

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodGet,
		nil,
		&data,
	)
	t.Log(data)
}
