package models

import (
	"android_vision_scripter/pkg/core/file"
	"math/rand"
	"strings"
)

const FlowNodeSeparator = "/"

type FlowEdge struct {
	Steps    []Step `json:"steps"`
	NextNode string `json:"next_node"`
}

func (e *FlowEdge) Valid() bool {
	if e == nil {
		return false
	}
	return ValidFlowNodeName(e.NextNode) && ValidQueue(e.Steps)
}

type FlowNode struct {
	Name  string     `json:"name"`
	Edges []FlowEdge `json:"edges"`
}

func SplitFlowNodeName(name string) (string, string, bool) {
	parts := strings.Split(strings.TrimSpace(name), FlowNodeSeparator)
	if len(parts) != 2 {
		return "", "", false
	}

	application := strings.TrimSpace(parts[0])
	node := strings.TrimSpace(parts[1])
	if !file.ValidName(application) || !file.ValidName(node) {
		return "", "", false
	}
	return application, node, true
}

func ValidFlowNodeName(name string) bool {
	_, _, ok := SplitFlowNodeName(name)
	return ok
}

func NormalizeFlowNodeName(name string) string {
	application, node, ok := SplitFlowNodeName(name)
	if !ok {
		return ""
	}
	return application + FlowNodeSeparator + node
}

func FindFlowPath(nodes []FlowNode, start string, end string) ([]FlowEdge, bool) {
	cache := map[string][]FlowEdge{}
	known := map[string]bool{}
	for _, node := range nodes {
		known[node.Name] = true
		cache[node.Name] = append(cache[node.Name], node.Edges...)
		for _, edge := range node.Edges {
			known[edge.NextNode] = true
		}
	}

	if !known[start] || !known[end] {
		return nil, false
	}
	if start == end {
		return []FlowEdge{}, true
	}

	parentNode := map[string]string{}
	parentEdge := map[string]FlowEdge{}
	visited := map[string]bool{start: true}
	queue := []string{start}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		edges := append([]FlowEdge{}, cache[current]...)
		rand.Shuffle(len(edges), func(a, b int) {
			edges[a], edges[b] = edges[b], edges[a]
		})

		for _, edge := range edges {
			if visited[edge.NextNode] {
				continue
			}
			visited[edge.NextNode] = true
			parentNode[edge.NextNode] = current
			parentEdge[edge.NextNode] = edge

			if edge.NextNode == end {
				return buildFlowPath(parentNode, parentEdge, start, end), true
			}
			queue = append(queue, edge.NextNode)
		}
	}
	return nil, false
}

func buildFlowPath(
	parentNode map[string]string,
	parentEdge map[string]FlowEdge,
	start string,
	end string,
) []FlowEdge {
	path := []FlowEdge{}
	for current := end; current != start; current = parentNode[current] {
		path = append(path, parentEdge[current])
	}

	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}
	return path
}
