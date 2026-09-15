package server

import (
	"android_vision_scripter/internal/usecases"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"errors"
	"net/http"
)

// Steps paths
const (
	QueueSteps = Devices + "/{" + SerialKey + "}/queue_steps"
)

func (s *serverImpl) handleStepsFunctions() {
	http.HandleFunc(QueueSteps, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleQueueSteps(w, r)
			return
		}
		http.Error(w, "use POST method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleQueueSteps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var serial = r.PathValue(SerialKey)
	if serial == "" {
		http.Error(w, `"serial" param required`, http.StatusBadRequest)
		return
	}

	var steps []models.Step
	err := json.NewDecoder(r.Body).Decode(&steps)
	if err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if !models.ValidQueue(steps) {
		http.Error(w, "invalid steps", http.StatusBadRequest)
		return
	}

	err = s.interactor.RunSteps(serial, steps, s.serverProps.SocketPort)
	if errors.Is(err, usecases.ErrRecordingInProgress) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}
