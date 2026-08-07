package server

import (
	"net/http"
)

const (
	RunFlow = "/run_flow/{" + SerialKey + "}"
)

const (
	StartKey = "start"
	EndKey   = "end"
)

func (s *serverImpl) handleFlowFunctions() {
	http.HandleFunc(RunFlow, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleFlow(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleFlow(w http.ResponseWriter, r *http.Request) {
	var serial = r.PathValue(SerialKey)
	var startNode = r.URL.Query().Get(StartKey)
	var endNode = r.URL.Query().Get(EndKey)

	if serial == "" || startNode == "" || endNode == "" {
		http.Error(w, `check queries`, http.StatusBadRequest)
		return
	}

	err := s.interactor.RunFlow(serial, startNode, endNode, s.serverProps.SocketPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}
