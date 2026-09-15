package server

import (
	"android_vision_scripter/internal/usecases"
	"android_vision_scripter/pkg/core/file"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Session paths
const (
	DeviceSession = Devices + "/{" + SerialKey + "}/session"
	DeviceRecord  = Devices + "/{" + SerialKey + "}/record"
)

type recordRequest struct {
	Name string `json:"name"`
}

func (s *serverImpl) handleSessionFunctions() {
	http.HandleFunc(DeviceSession, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetSession(w, r)
			return
		}
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleOpenSession(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleCloseSession(w, r)
			return
		}
		http.Error(w, "use another method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(DeviceRecord, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleRecordAction(w, r)
			return
		}
		http.Error(w, "use POST method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleRecordAction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var serial = r.PathValue(SerialKey)
	if serial == "" {
		http.Error(w, `"serial" param required`, http.StatusBadRequest)
		return
	}

	var request recordRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || !file.ValidName(request.Name) {
		http.Error(w, `valid "name" required`, http.StatusBadRequest)
		return
	}

	recorded, err := s.interactor.RecordAction(serial, request.Name, s.serverProps.SocketPort)
	if errors.Is(err, usecases.ErrDeviceBusy) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !recorded {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	s.sendStatusOk(w)
}

func (s *serverImpl) handleGetSession(w http.ResponseWriter, r *http.Request) {
	var serial = r.PathValue(SerialKey)
	if serial == "" {
		http.Error(w, `"serial" param required`, http.StatusBadRequest)
		return
	}

	var response = map[string]string{
		"status": s.interactor.GetSessionStatus(serial),
	}

	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}

func (s *serverImpl) handleOpenSession(w http.ResponseWriter, r *http.Request) {
	var serial = r.PathValue(SerialKey)
	if serial == "" {
		http.Error(w, `"serial" param required`, http.StatusBadRequest)
		return
	}

	var started = s.interactor.StartSession(serial, s.serverProps.SocketPort)
	if !started {
		var errMsg = fmt.Sprintf("couldn't start scrcpy server for %s", serial)
		http.Error(w, errMsg, http.StatusInternalServerError)
		s.interactor.CloseSession(serial)
		return
	}

	var data = s.interactor.GetPortsJSON(serial)

	s.setHeaders(w)
	json.NewEncoder(w).Encode(data)

	go s.acceptSocketConnections(serial)
}

func (s *serverImpl) handleCloseSession(w http.ResponseWriter, r *http.Request) {
	var serial = r.PathValue(SerialKey)
	if serial == "" {
		http.Error(w, `"serial" param required`, http.StatusBadRequest)
		return
	}

	s.interactor.CloseSession(serial)
	s.sendStatusOk(w)
}

func (s *serverImpl) acceptSocketConnections(serial string) {
	doneCtx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		s.interactor.AcceptVideoConnections(doneCtx, serial)
	}()

	go func() {
		defer cancel()
		s.interactor.AcceptControlConnection(doneCtx, serial)
	}()

	<-doneCtx.Done()

	s.interactor.CloseSession(serial)
}
