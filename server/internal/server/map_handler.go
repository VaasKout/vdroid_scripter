package server

import (
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

const (
	MapPath       = "/map"
	MapNodeByName = MapPath + "/{" + NameKey + "}"
	FollowRoute   = "/follow_route"
)

func (s *serverImpl) handleMapFunctions() {
	http.HandleFunc(MapPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetMap(w)
			return
		}
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveMapNode(w, r)
			return
		}
		http.Error(w, "use GET or POST method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(MapNodeByName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetMapNode(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleDeleteMapNode(w, r)
			return
		}
		http.Error(w, "use GET or DELETE method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(FollowRoute, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleFollowRoute(w, r)
			return
		}
		http.Error(w, "use POST method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleGetMap(w http.ResponseWriter) {
	var response = map[string]any{
		"nodes": s.interactor.GetMapNodes(),
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}

func (s *serverImpl) handleGetMapNode(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	node, err := s.interactor.GetMapNode(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(node)
}

func (s *serverImpl) handleSaveMapNode(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var node = &models.Node{}
	err := json.NewDecoder(r.Body).Decode(node)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = s.interactor.SaveMapNode(node)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.sendStatusOk(w)
}

func (s *serverImpl) handleDeleteMapNode(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	deleted := s.interactor.DeleteMapNode(name)
	if deleted {
		s.sendStatusOk(w)
		return
	}
	http.Error(w, "node not found", http.StatusNotFound)
}

func (s *serverImpl) handleFollowRoute(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data = struct {
		Serial string `json:"serial"`
		From   string `json:"from"`
		To     string `json:"to"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Serial == "" || data.From == "" || data.To == "" {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = s.interactor.FollowRoute(data.Serial, data.From, data.To, s.serverProps.SocketPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}
