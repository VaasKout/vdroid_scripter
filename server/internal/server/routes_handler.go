package server

import (
	"android_vision_scripter/internal/usecases"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	EndIDKey   = "end_id"
	ArgsKey    = "args"
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
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
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

	startID, err := stepIDQuery(r, StartIDKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	endID, err := stepIDQuery(r, EndIDKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params := &usecases.RunRouteParams{
		Serial:   serial,
		Name:     name,
		StartID:  startID,
		EndID:    endID,
		Args:     r.URL.Query()[ArgsKey],
		BasePort: s.serverProps.SocketPort,
	}
	err = s.interactor.RunRoute(params)
	if errors.Is(err, usecases.ErrRecordingInProgress) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if errors.Is(err, models.ErrRouteArgs) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}

func stepIDQuery(r *http.Request, key string) (int, error) {
	value := r.URL.Query().Get(key)
	raw := strings.TrimSpace(value)
	if raw == "" {
		return 0, nil
	}

	id, err := strconv.Atoi(raw)
	if err != nil || id < 0 {
		return 0, fmt.Errorf("invalid %s: %s", key, raw)
	}
	return id, nil
}
