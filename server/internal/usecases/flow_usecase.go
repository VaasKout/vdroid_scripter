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

	scriptsByNode, err := i.getScriptsByNodes()
	if err != nil {
		return err
	}
	if len(scriptsByNode) == 0 {
		return errors.New(FlowIsEmptyError)
	}

	nodes := i.findNodesPath(scriptsByNode, startNode, endNode)
	if len(nodes) == 0 {
		return errors.New(FlowPathNotFoundError)
	}

	i.logger.Info(fmt.Sprintf("flow: %v", nodes))

	var newConnection = false
	if _, ok := i.clientsCache.Get(serial); !ok {
		started := i.StartScrcpyServer(serial, basePort)
		if !started {
			var errMsg = fmt.Sprintf("couldn't start scrcpy server for %s", serial)
			return errors.New(errMsg)
		}
		newConnection = started
	}

	go i.startFlow(serial, nodes, scriptsByNode, newConnection)
	return nil
}

func (i *interactorImpl) startFlow(
	serial string,
	nodes []string,
	scriptsByNode map[string][]models.Script,
	newConnection bool,
) {
	var doneCh = make(chan struct{})
	go func() {
		defer close(doneCh)
		if newConnection {
			go i.scrcpy.ReadVideoStream(serial, nil)
		}
		i.executeFlow(serial, nodes, scriptsByNode)
	}()

	<-doneCh
	if newConnection {
		i.CloseConnection(serial)
	}
}

func (i *interactorImpl) executeFlow(
	serial string,
	nodes []string,
	scriptsByNode map[string][]models.Script,
) {
	for index, node := range nodes {
		var scripts = scriptsByNode[node]
		var nextNode = ""
		if index < len(nodes)-1 {
			nextNode = nodes[index+1]
		}
		var candidates = []models.Script{}
		for _, script := range scripts {
			if script.Name == models.InitialScript {
				completed := i.executeScript(serial, &script)
				if !completed {
					return
				}
				continue
			}

			if nextNode != "" && strings.TrimSpace(script.NextNode) == nextNode {
				candidates = append(candidates, script)
			}
		}

		if nextNode == "" {
			return
		}

		if len(candidates) == 0 {
			i.logger.Error(fmt.Sprintf("no script leads from %s to %s", node, nextNode))
			return
		}

		var script = candidates[rand.IntN(len(candidates))]
		completed := i.executeScript(serial, &script)
		if !completed && !script.CanSkip {
			return
		}
	}
}

func (i *interactorImpl) getScriptsByNodes() (map[string][]models.Script, error) {
	nodes, err := i.GetNodes()
	if err != nil {
		return map[string][]models.Script{}, err
	}

	var scriptsByNode = make(map[string][]models.Script, len(nodes))
	for _, node := range nodes {
		scripts := i.getAllScriptsFromNode(node)
		if len(scripts) == 0 {
			continue
		}
		scriptsByNode[node] = scripts
	}
	return scriptsByNode, nil
}

func (i *interactorImpl) findNodesPath(
	scriptsByNode map[string][]models.Script,
	startNode string,
	endNode string,
) []string {
	var visited = map[string]models.Script{}
	var queue = []string{startNode}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:] // pop node

		if node == endNode {
			return buildNodesPath(visited, startNode, endNode)
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

func buildNodesPath(
	visited map[string]models.Script,
	startNode string,
	endNode string,
) []string {
	var path = []string{endNode}
	for node := endNode; node != startNode; {
		script, ok := visited[node]
		if !ok {
			return nil
		}
		node = script.Node
		path = append(path, node)
	}
	slices.Reverse(path)
	return path
}
