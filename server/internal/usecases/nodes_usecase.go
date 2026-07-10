package usecases

import (
	"android_vision_scripter/internal/filesdb"
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Common node errors
const (
	NodeNameIsEmpty = "node name is empty"
)

// NodesUseCase ...
type NodesUseCase interface {
	GetScreenNodes() []string
	GetScreenNode(name string) (*models.ScreenNode, error)
	SaveScreenNode(
		serial string,
		name string,
		anchors map[string]string,
		actions []models.Action,
	) error
	DeleteScreenNode(name string) error
}

func (i *interactorImpl) GetScreenNodes() []string {
	screenNodesDir := i.filesDB.CreateScreenNodesDir()
	dirs := i.filesDB.GetDirs(screenNodesDir)
	names := make([]string, len(dirs))
	for index, node := range dirs {
		names[index] = filepath.Base(node)
	}
	return names

}

func (i *interactorImpl) GetScreenNode(name string) (*models.ScreenNode, error) {
	name = strings.TrimSpace(name)

	nodePath := i.getScreenNodePath(name)
	if nodePath == "" {
		return &models.ScreenNode{}, fmt.Errorf("unable to resolve node path for %s", name)
	}

	bytes, err := os.ReadFile(nodePath)
	if err != nil {
		return &models.ScreenNode{}, err
	}

	var node = &models.ScreenNode{}
	if err := node.FromJSON(bytes); err != nil {
		return &models.ScreenNode{}, err
	}
	return node, nil
}

func (i *interactorImpl) SaveScreenNode(
	serial string,
	name string,
	anchors map[string]string,
	actions []models.Action,
) error {
	if strings.TrimSpace(serial) == "" || strings.TrimSpace(name) == "" {
		return errors.New("node is empty")
	}

	nodePath := i.getScreenNodePath(name)
	if nodePath == "" {
		return fmt.Errorf("unable to resolve node path for %s", name)
	}

	device := i.GetDevice(serial).ToModelOs()

	node := models.ScreenNode{
		Name:    name,
		Device:  device,
		Anchors: anchors,
		Actions: actions,
	}

	bytes := node.ToJSON()
	if len(bytes) == 0 {
		return fmt.Errorf("unable to marshal node %s", node.Name)
	}
	return os.WriteFile(nodePath, bytes, 0644)
}

func (i *interactorImpl) DeleteScreenNode(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New(NodeNameIsEmpty)
	}

	nodesDir := i.filesDB.CreateScreenNodesDir()
	i.filesDB.DeleteFileByName(nodesDir, name)
	return nil
}

func (i *interactorImpl) getScreenNodePath(name string) string {
	scriptDir := i.filesDB.CreateScreenNodesDir(name)
	if scriptDir == "" {
		return ""
	}

	nodeJSON := filepath.Join(scriptDir, filesdb.NodeJSON)
	if ok := file.CreateFileIfNotExist(nodeJSON); !ok {
		return ""
	}
	return nodeJSON
}
