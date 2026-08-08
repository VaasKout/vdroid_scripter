package test

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/bashcmd"
	"android_vision_scripter/internal/cv"
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/internal/scrcpy"
	"android_vision_scripter/internal/server"
	"android_vision_scripter/internal/usecases"
	"android_vision_scripter/internal/yolo"
	"android_vision_scripter/pkg/core/network"
	"android_vision_scripter/pkg/logger"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
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

func TestBuildFlow(t *testing.T) {
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

	var cfg = config.New()
	var logAPI = logger.New(logger.INFO, true)
	var filesDB = filesdb.New(cfg.FilesProps)
	var cmdRunner = bashcmd.New(filesDB, logAPI)
	var cvAPI = cv.New(cmdRunner, logAPI)
	var scrcpyAPI = scrcpy.New(cmdRunner, cvAPI, filesDB, cfg.ScrcpyProps, logAPI)
	var yoloAPI = yolo.New(filesDB, logAPI)
	var networkAPI = network.New(logAPI)
	var interactor = usecases.New(
		cvAPI, cmdRunner, filesDB, scrcpyAPI, yoloAPI, networkAPI, logAPI,
	)

	var flow = interactor.BuildFlow(TestFlowNode+"1", TestFlowNode+"3")
	if len(flow) == 0 {
		t.Fatal("empty flow")
	}

	var names = make([]string, 0, len(flow))
	for _, script := range flow {
		names = append(names, script.Name)
	}
	t.Log(names)
}
