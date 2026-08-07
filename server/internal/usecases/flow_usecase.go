package usecases

import (
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
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

	scriptsByNode, err := i.getAllScripts()
	if err != nil {
		return err
	}

	flowScripts := i.findScriptsPath(scriptsByNode, startNode, endNode)
	if flowScripts == nil {
		return errors.New(FlowPathNotFoundError)
	}

	i.logger.Info(fmt.Sprintf("flow: %v", flowScripts))

	return nil
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
	var visited = map[string]models.Script{startNode: {}}
	var queue = []string{startNode}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:] // pop node

		if node == endNode {
			return buildScriptsPath(visited, startNode, endNode)
		}

		for _, script := range scriptsByNode[node] {
			var next = strings.TrimSpace(script.NextNode)
			if next == "" {
				continue
			}
			if _, ok := visited[next]; ok {
				continue
			}
			visited[next] = script
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
	for node := endNode; node != startNode; { // iterate until node != startNode
		var script = visited[node]
		path = append(path, script)
		node = script.Node
	}
	slices.Reverse(path)
	return path
}
