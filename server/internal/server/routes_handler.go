package server

import (
	"android_vision_scripter/pkg/core/strutils"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

// Route paths
const (
	RoutesPath  = "/routes"
	RouteByName = RoutesPath + "/{" + NameKey + "}"
	RunRoute    = "/run_route"
)

// Route query keys
const (
	StartIDKey = "start_id"
)

func (s *serverImpl) handleRouteFunctions() {
	http.HandleFunc(RoutesPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetRoutes(w)
			return
		}
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveRoute(w, r)
			return
		}
		http.Error(w, "use GET or POST method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(RouteByName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetRoute(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleDeleteRoute(w, r)
			return
		}
		http.Error(w, "use GET or DELETE method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(RunRoute, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleRunRoute(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleGetRoutes(w http.ResponseWriter) {
	var response = map[string]any{
		"routes": s.interactor.GetRoutes(),
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}

func (s *serverImpl) handleGetRoute(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	route, err := s.interactor.GetRoute(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(route)
}

func (s *serverImpl) handleSaveRoute(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var route = &models.Route{}
	err := json.NewDecoder(r.Body).Decode(route)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = s.interactor.SaveRoute(route)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.sendStatusOk(w)
}

func (s *serverImpl) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	deleted := s.interactor.DeleteRoute(name)
	if deleted {
		s.sendStatusOk(w)
		return
	}
	http.Error(w, "route not found", http.StatusNotFound)
}

func (s *serverImpl) handleRunRoute(w http.ResponseWriter, r *http.Request) {
	var serial = r.URL.Query().Get(SerialKey)
	var name = r.URL.Query().Get(NameKey)
	if serial == "" || name == "" {
		http.Error(w, `"serial" and "name" queries needed`, http.StatusBadRequest)
		return
	}

	var startID int
	rawStartID := r.URL.Query().Get(StartIDKey)
	if rawStartID != "" {
		startID = strutils.ToInt(rawStartID)
		if startID < 1 {
			http.Error(w, "invalid start_id: "+rawStartID, http.StatusBadRequest)
			return
		}
	}

	err := s.interactor.RunRoute(serial, name, startID, s.serverProps.SocketPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}
