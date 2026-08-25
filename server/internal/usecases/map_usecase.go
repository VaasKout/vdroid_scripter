package usecases

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const nodeFileName = "node" + file.JsonExt

type MapUseCase interface {
	GetMapNodes() []string
	GetMapNode(name string) (*models.Node, error)
	SaveMapNode(node *models.Node) error
	DeleteMapNode(name string) bool
}

func (i *interactorImpl) GetMapNodes() []string {
	names := []string{}
	mapDir := i.filesDB.CreateMapDir()
	if mapDir == "" {
		return names
	}
	for _, dir := range i.filesDB.GetDirs(mapDir, false) {
		names = append(names, file.GetFileName(dir))
	}
	sort.Strings(names)
	return names
}

func (i *interactorImpl) GetMapNode(name string) (*models.Node, error) {
	name = strings.TrimSpace(name)
	if !file.ValidName(name) {
		return nil, fmt.Errorf("invalid node name: %s", name)
	}

	mapDir := i.filesDB.CreateMapDir()
	if mapDir == "" {
		return nil, errors.New("map dir not found")
	}

	nodePath := filepath.Join(mapDir, name, nodeFileName)
	data, err := os.ReadFile(nodePath)
	if err != nil {
		return nil, fmt.Errorf("node not found in map: %s", name)
	}

	var node = &models.Node{}
	err = json.Unmarshal(data, node)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (i *interactorImpl) SaveMapNode(node *models.Node) error {
	if node == nil {
		return errors.New("node is empty")
	}
	node.Name = strings.TrimSpace(node.Name)
	if node.OccupancyGrid.IsEmpty() {
		node.OccupancyGrid = nil
	}
	if !node.Valid() {
		return errors.New("invalid node")
	}
	if err := i.checkNodeAssets(node); err != nil {
		return err
	}

	stored, err := i.GetMapNode(node.Name)
	if err == nil {
		stored.Merge(node)
		node = stored
	}
	return i.writeMapNode(node)
}

func (i *interactorImpl) writeMapNode(node *models.Node) error {
	nodeDir := i.filesDB.CreateMapDir(node.Name)
	if nodeDir == "" {
		return fmt.Errorf("couldn't create node dir: %s", node.Name)
	}

	data, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(nodeDir, nodeFileName), data, 0644)
}

func (i *interactorImpl) DeleteMapNode(name string) bool {
	if !file.ValidName(name) {
		return false
	}
	mapDir := i.filesDB.CreateMapDir()
	if mapDir == "" {
		return false
	}
	return i.filesDB.DeleteDirByName(mapDir, strings.TrimSpace(name))
}

func (i *interactorImpl) checkNodeAssets(node *models.Node) error {
	steps := []models.Step{}
	if len(node.Landmarks) > 0 {
		steps = append(steps, models.Step{Landmarks: node.Landmarks})
	}
	for _, edge := range node.Edges {
		steps = append(steps, edge.Action)
	}
	if len(steps) == 0 {
		return nil
	}
	return i.checkStepAssets(steps)
}
