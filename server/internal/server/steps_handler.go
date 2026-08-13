package server

import (
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

const (
	RunSteps = "/run_steps"
	FindText = "/devices/{" + SerialKey + "}/find_text"
)

const (
	TextKey = "text"
)

func (s *serverImpl) handleStepsFunctions() {
	http.HandleFunc(RunSteps, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleRunSteps(w, r)
			return
		}
		http.Error(w, "use POST method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(FindText, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleFindText(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleRunSteps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data = struct {
		Serial string        `json:"serial"`
		Steps  []models.Step `json:"steps"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Serial == "" || !models.ValidQueue(data.Steps) {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = s.interactor.RunSteps(data.Serial, data.Steps, s.serverProps.SocketPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}

func (s *serverImpl) handleFindText(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var serial = r.PathValue(SerialKey)
	var text = r.URL.Query().Get(TextKey)
	var locale = r.URL.Query().Get(LocaleKey)

	if serial == "" {
		http.Error(w, `"serial" query needed`, http.StatusBadRequest)
		return
	}

	result := s.interactor.FindText(serial, text, locale)
	s.setHeaders(w)
	json.NewEncoder(w).Encode(result)
}
