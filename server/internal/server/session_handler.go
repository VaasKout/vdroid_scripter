package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Session paths
const (
	DeviceSession = Devices + "/{" + SerialKey + "}/session"
)

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
}

func (s *serverImpl) handleGetSession(w http.ResponseWriter, r *http.Request) {
	var serial = r.PathValue(SerialKey)
	if serial == "" {
		http.Error(w, `"serial" param required`, http.StatusBadRequest)
		return
	}

	var ports = s.interactor.GetPortsJSON(serial)
	if len(ports) == 0 {
		http.Error(w, "no session", http.StatusNotFound)
		return
	}

	s.setHeaders(w)
	json.NewEncoder(w).Encode(ports)
}

func (s *serverImpl) handleOpenSession(w http.ResponseWriter, r *http.Request) {
	var serial = r.PathValue(SerialKey)
	if serial == "" {
		http.Error(w, `"serial" param required`, http.StatusBadRequest)
		return
	}

	var started = s.interactor.StartScrcpyServer(serial, s.serverProps.SocketPort)
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
		s.interactor.AcceptCvConnection(doneCtx, serial)
	}()

	go func() {
		defer cancel()
		s.interactor.AcceptControlConnection(doneCtx, serial)
	}()

	<-doneCtx.Done()

	s.interactor.CloseSession(serial)
}
