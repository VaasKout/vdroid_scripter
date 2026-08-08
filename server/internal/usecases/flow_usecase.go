package usecases

import (
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
)

const (
	FlowIsEmptyError      = "flow is empty"
	FlowPathNotFoundError = "no flow path found"
)

type FlowUseCase interface {
	RunFlow(
		serial string,
		startNode string,
		endNode string,
		basePort int,
	) error
	BuildFlow(startNode string, endNode string) []models.Script
}

func (i *interactorImpl) RunFlow(
	serial string,
	startNode string,
	endNode string,
	basePort int,
) error {
	serial = strings.TrimSpace(serial)
	startNode = strings.TrimSpace(startNode)
	endNode = strings.TrimSpace(endNode)

	if serial == "" {
		return errors.New(SerialIsEmptyError)
	}
	if startNode == "" || endNode == "" {
		return errors.New(FlowIsEmptyError)
	}

	flowScripts := i.BuildFlow(startNode, endNode)
	if len(flowScripts) == 0 {
		return errors.New(FlowPathNotFoundError)
	}

	i.logger.Info(fmt.Sprintf("flow: %v", flowScripts))

	var newConnection = false
	if _, ok := i.clientsCache.Get(serial); !ok {
		started := i.StartScrcpyServer(serial, basePort)
		if !started {
			var errMsg = fmt.Sprintf("couldn't start scrcpy server for %s", serial)
			return errors.New(errMsg)
		}
		newConnection = started
	}

	go i.startFlowListener(serial, flowScripts, newConnection)
	return nil
}

func (i *interactorImpl) startFlowListener(
	serial string,
	scripts []models.Script,
	newConnection bool,
) {
	var doneCh = make(chan struct{})

	go func() {
		defer close(doneCh)
		for index := range scripts {
			i.executeScript(serial, &scripts[index], newConnection && index == 0)
		}
	}()

	<-doneCh
	if newConnection {
		i.CloseConnection(serial)
	}
}

func (i *interactorImpl) BuildFlow(startNode string, endNode string) []models.Script {
	scriptsByNode, err := i.getAllScripts()
	if err != nil {
		i.logger.Error(err.Error())
		return []models.Script{}
	}
	return i.findScriptsPath(scriptsByNode, startNode, endNode)
}

func (i *interactorImpl) getAllScripts() (map[string][]models.Script, error) {
	nodes, err := i.GetNodes("")
	if err != nil {
		return nil, err
	}

	var scriptsByNode = make(map[string][]models.Script, len(nodes))
	for _, node := range nodes {
		names, err := i.GetNodes(node) //take all nodes
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			continue
		}

		var scripts = make([]models.Script, 0, len(names))
		for _, name := range names {
			script, err := i.GetScript(node, name) //take all scripts by name ""
			if err != nil {
				i.logger.Error(err.Error())
				continue
			}
			if script.IsEmpty() {
				continue
			}
			scripts = append(scripts, *script)
		}
		scriptsByNode[node] = scripts
	}
	return scriptsByNode, nil
}

// start from the endNode, check all scripts to have nextNode: lastNode
func (i *interactorImpl) findScriptsPath(
	scriptsByNode map[string][]models.Script,
	startNode string,
	endNode string,
) []models.Script {
	var visited = map[string]models.Script{}
	var queue = []string{startNode}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:] // pop node

		if node == endNode {
			return buildScriptsPath(visited, startNode, endNode)
		}

		var candidates = map[string][]models.Script{}
		for _, script := range scriptsByNode[node] {
			var next = strings.TrimSpace(script.NextNode)
			if next == "" {
				continue
			}
			if _, ok := visited[next]; ok {
				continue
			}
			candidates[next] = append(candidates[next], script)
		}

		for next, scripts := range candidates {
			visited[next] = scripts[rand.IntN(len(scripts))]
			queue = append(queue, next)
		}
	}
	return nil
}

func buildScriptsPath(
	visited map[string]models.Script,
	startNode string,
	endNode string,
) []models.Script {
	var path = make([]models.Script, 0, len(visited))
	for node := endNode; node != startNode; {
		script, ok := visited[node]
		if !ok {
			return nil
		}
		path = append(path, script)
		node = script.Node
	}
	slices.Reverse(path)
	return path
}
