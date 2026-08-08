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
	BuildFlow(startNode string, endNode string) []string
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

	nodes := i.BuildFlow(startNode, endNode)
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

	go i.startFlow(serial, nodes, newConnection)
	return nil
}

func (i *interactorImpl) BuildFlow(startNode string, endNode string) []string {
	scriptsByNode, err := i.getAllScripts()
	if err != nil {
		i.logger.Error(err.Error())
		return []string{}
	}
	return i.findNodesPath(scriptsByNode, startNode, endNode)
}

func (i *interactorImpl) startFlow(
	serial string,
	nodes []string,
	newConnection bool,
) {
	var doneCh = make(chan struct{})

	go func() {
		defer close(doneCh)
		if newConnection {
			go i.scrcpy.ReadVideoStream(serial, nil)
		}
		i.executeFlow(serial, nodes)
	}()

	<-doneCh
	if newConnection {
		i.CloseConnection(serial)
	}
}

func (i *interactorImpl) executeFlow(
	serial string,
	nodes []string,
) {
	for index, node := range nodes {

		var scripts = scriptsByNode[node]

		for index := range scripts {
			if scripts[index].Name == models.InitialScript {
				var initScript = scripts[index]
				if initScript.IsEmpty() {
					return
				}
				completed := i.executeScript(serial, &initScript)
				if !completed {
					return
				}
			}
		}

		if index == len(nodes)-1 {
			return
		}

		var nextNode = nodes[index+1]
		var candidates = []models.Script{}
		for _, script := range scripts {
			if script.Name == models.InitialScript {
				continue
			}
			if strings.TrimSpace(script.NextNode) == nextNode {
				candidates = append(candidates, script)
			}
		}
		if len(candidates) == 0 {
			i.logger.Error(fmt.Sprintf("no script leads from %s to %s", node, nextNode))
			return
		}

		var script = candidates[rand.IntN(len(candidates))]
		i.executeScript(serial, &script)
	}
}

func (i *interactorImpl) getAllScripts() (map[string][]models.Script, error) {
	nodes, err := i.GetNodes("")
	if err != nil {
		return nil, err
	}

	var scriptsByNode = make(map[string][]models.Script, len(nodes))
	for _, node := range nodes {
		scriptNames, err := i.GetNodes(node) //take all nodes
		if err != nil {
			return nil, err
		}
		if len(scriptNames) == 0 {
			continue
		}

		var scripts = make([]models.Script, 0, len(scriptNames))
		for _, name := range scriptNames {
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
