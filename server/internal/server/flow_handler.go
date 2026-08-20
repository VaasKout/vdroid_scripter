package server

import (
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

const (
	Flow    = "/flow"
	RunFlow = "/run_flow"

	StartKey = "start"
	EndKey   = "end"
)

type flowNodeResponse struct {
	Name      string   `json:"name"`
	NextNodes []string `json:"next_nodes"`
}

func (s *serverImpl) handleFlowFunctions() {
	http.HandleFunc(Flow, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetFlowNodes(w, r)
			return
		}
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveFlowNode(w, r)
			return
		}
		http.Error(w, "use GET or POST method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(RunFlow, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleRunFlow(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleGetFlowNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.interactor.GetFlowNodes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]flowNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		nextNodes := make([]string, 0, len(node.Edges))
		for _, edge := range node.Edges {
			nextNodes = append(nextNodes, edge.NextNode)
		}
		response = append(response, flowNodeResponse{
			Name:      node.Name,
			NextNodes: nextNodes,
		})
	}

	s.setHeaders(w)
	json.NewEncoder(w).Encode(map[string][]flowNodeResponse{"nodes": response})
}

func (s *serverImpl) handleSaveFlowNode(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data = struct {
		Node     string        `json:"node"`
		NextNode string        `json:"next_node"`
		Steps    []models.Step `json:"steps"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if !models.ValidFlowNodeName(data.Node) || !models.ValidFlowNodeName(data.NextNode) {
		http.Error(w, "node and next_node must be application_name/node_name", http.StatusBadRequest)
		return
	}
	if !models.ValidQueue(data.Steps) {
		http.Error(w, "invalid steps: tap/long_tap and visibility checks need anchors, type_text needs a value in its last anchor", http.StatusBadRequest)
		return
	}

	edge := &models.FlowEdge{Steps: data.Steps, NextNode: data.NextNode}
	err = s.interactor.SaveFlowNode(data.Node, edge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}

func (s *serverImpl) handleRunFlow(w http.ResponseWriter, r *http.Request) {
	var serial = r.URL.Query().Get(SerialKey)
	var start = r.URL.Query().Get(StartKey)
	var end = r.URL.Query().Get(EndKey)
	if serial == "" || start == "" || end == "" {
		http.Error(w, `"serial", "start" and "end" params required`, http.StatusBadRequest)
		return
	}

	err := s.interactor.RunFlow(serial, start, end, s.serverProps.SocketPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}
