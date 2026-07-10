package server

import (
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

// Screen node patterns...
const (
	ScreenNodes       = "/screen_nodes"
	ScreenNodesByName = ScreenNodes + "/{" + NameKey + "}"
)

func (s *serverImpl) handleNodesFunctions() {
	http.HandleFunc(ScreenNodes, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetScreenNodes(w, r)
			return
		}
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveScreenNode(w, r)
			return
		}
		http.Error(w, "use another method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(ScreenNodesByName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetScreenNode(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleDeleteScreenNode(w, r)
			return
		}
		http.Error(w, "use another method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleGetScreenNodes(w http.ResponseWriter, r *http.Request) {
	names := s.interactor.GetScreenNodes()
	s.setHeaders(w)
	json.NewEncoder(w).Encode(names)
}

func (s *serverImpl) handleGetScreenNode(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	if name == "" {
		http.Error(w, `"name" query needed`, http.StatusBadRequest)
		return
	}

	node, err := s.interactor.GetScreenNode(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.setHeaders(w)
	json.NewEncoder(w).Encode(node)
}

func (s *serverImpl) handleSaveScreenNode(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data struct {
		Serial  string            `json:"serial"`
		Name    string            `json:"name"`
		Anchors map[string]string `json:"anchors,omitempty"`
		Actions []models.Action   `json:"actions,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Serial == "" || data.Name == "" {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = s.interactor.SaveScreenNode(
		data.Serial,
		data.Name,
		data.Anchors,
		data.Actions,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}

func (s *serverImpl) handleDeleteScreenNode(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	if name == "" {
		http.Error(w, `"name" query needed`, http.StatusBadRequest)
		return
	}
	s.interactor.DeleteScreenNode(name)
	s.sendStatusOk(w)
}
