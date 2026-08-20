package server

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

const (
	Scripts       = "/scripts"
	ScriptsByName = Scripts + "/{" + NameKey + "}"
	RunScript     = "/run_script"
)

func (s *serverImpl) handleScriptsFunctions() {
	http.HandleFunc(Scripts, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetScripts(w, r)
			return
		}
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveScript(w, r)
			return
		}
		http.Error(w, "use GET or POST method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(ScriptsByName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetScript(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleDeleteScript(w, r)
			return
		}
		http.Error(w, "use GET or DELETE method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(RunScript, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleRunScript(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleGetScripts(w http.ResponseWriter, r *http.Request) {
	var response = map[string][]string{
		"scripts": s.interactor.GetScripts(),
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}

func (s *serverImpl) handleGetScript(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	steps, err := s.interactor.GetScript(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	s.setHeaders(w)
	json.NewEncoder(w).Encode(steps)
}

func (s *serverImpl) handleSaveScript(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data = struct {
		Name  string        `json:"name"`
		Steps []models.Step `json:"steps"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if !file.ValidName(data.Name) {
		http.Error(w, `"name" is required`, http.StatusBadRequest)
		return
	}
	if !models.ValidQueue(data.Steps) {
		http.Error(w, "invalid steps: tap/long_tap and visibility checks need anchors, type_text needs a value in its last anchor", http.StatusBadRequest)
		return
	}

	err = s.interactor.SaveScript(data.Name, data.Steps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}

func (s *serverImpl) handleDeleteScript(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	deleted := s.interactor.DeleteScript(name)
	if deleted {
		s.sendStatusOk(w)
		return
	}

	http.Error(w, "script not found", http.StatusNotFound)
}

func (s *serverImpl) handleRunScript(w http.ResponseWriter, r *http.Request) {
	var serial = r.URL.Query().Get(SerialKey)
	var name = r.URL.Query().Get(NameKey)
	if serial == "" || name == "" {
		http.Error(w, `"serial" and "name" params required`, http.StatusBadRequest)
		return
	}

	err := s.interactor.RunScript(serial, name, s.serverProps.SocketPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}
