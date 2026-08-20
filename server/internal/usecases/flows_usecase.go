package usecases

import (
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const flowRunFile = "run.json"

type FlowsUseCase interface {
	GetFlowNodes() ([]models.FlowNode, error)
	SaveFlowNode(node string, edge *models.FlowEdge) error
	RunFlow(serial string, start string, end string, basePort int) error
}

func (i *interactorImpl) GetFlowNodes() ([]models.FlowNode, error) {
	flowsDir := i.filesDB.CreateFlowsDir()
	if flowsDir == "" {
		return nil, errors.New("flows dir not found")
	}

	nodes := []models.FlowNode{}
	for _, appDir := range i.filesDB.GetDirs(flowsDir, false) {
		for _, nodeDir := range i.filesDB.GetDirs(appDir, false) {
			name := filepath.Base(appDir) + models.FlowNodeSeparator + filepath.Base(nodeDir)
			edges, err := i.readFlowEdges(nodeDir)
			if err != nil {
				i.logger.Error(fmt.Sprintf("skipping flow node %s: %s", name, err.Error()))
				continue
			}
			nodes = append(nodes, models.FlowNode{Name: name, Edges: edges})
		}
	}

	sort.Slice(nodes, func(a, b int) bool {
		return nodes[a].Name < nodes[b].Name
	})
	return nodes, nil
}

func (i *interactorImpl) readFlowEdges(nodeDir string) ([]models.FlowEdge, error) {
	bytes, err := os.ReadFile(filepath.Join(nodeDir, flowRunFile))
	if os.IsNotExist(err) {
		return []models.FlowEdge{}, nil
	}
	if err != nil {
		return nil, err
	}

	edges := []models.FlowEdge{}
	err = json.Unmarshal(bytes, &edges)
	if err != nil {
		return nil, err
	}
	return edges, nil
}

func (i *interactorImpl) SaveFlowNode(node string, edge *models.FlowEdge) error {
	application, nodeName, ok := models.SplitFlowNodeName(node)
	if !ok {
		return fmt.Errorf("invalid node name: %s, use application_name/node_name", node)
	}
	if !edge.Valid() {
		return errors.New("invalid edge: next_node must be application_name/node_name and steps must be valid")
	}
	if err := i.checkStepAssets(edge.Steps); err != nil {
		return err
	}

	edge.NextNode = models.NormalizeFlowNodeName(edge.NextNode)
	nodeDir := i.filesDB.CreateFlowsDir(application, nodeName)
	if nodeDir == "" {
		return errors.New("couldn't create flow node dir")
	}

	edges, err := i.readFlowEdges(nodeDir)
	if err != nil {
		return err
	}

	replaced := false
	for index := range edges {
		if edges[index].NextNode != edge.NextNode {
			continue
		}
		edges[index] = *edge
		replaced = true
		break
	}
	if !replaced {
		edges = append(edges, *edge)
	}

	bytes, err := json.MarshalIndent(edges, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(nodeDir, flowRunFile), bytes, 0644)
}

func (i *interactorImpl) RunFlow(
	serial string,
	start string,
	end string,
	basePort int,
) error {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return errors.New(SerialIsEmptyError)
	}

	start = models.NormalizeFlowNodeName(start)
	end = models.NormalizeFlowNodeName(end)
	if start == "" || end == "" {
		return errors.New("start and end must be valid node names: application_name/node_name")
	}

	nodes, err := i.GetFlowNodes()
	if err != nil {
		return err
	}

	path, found := models.FindFlowPath(nodes, start, end)
	if !found {
		return fmt.Errorf("no flow path from %s to %s", start, end)
	}

	steps := []models.Step{}
	for _, edge := range path {
		if !edge.Valid() {
			return fmt.Errorf("invalid flow edge to %s", edge.NextNode)
		}
		if err := i.checkStepAssets(edge.Steps); err != nil {
			return err
		}
		steps = append(steps, edge.Steps...)
		steps[len(steps)-1].NextNode = edge.NextNode
	}

	if err := i.ensureSession(serial, basePort); err != nil {
		return err
	}
	i.addStepsToQueue(serial, start, steps)
	return nil
}
