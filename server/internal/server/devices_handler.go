package server

import (
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

// Devices paths
const (
	Ping    = "/ping"
	Devices = "/devices"
)

func (s *serverImpl) handleDeviceFunctions() {
	http.HandleFunc(Ping, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.sendStatusOk(w)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(Devices, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleDeviceList(w)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleDeviceList(w http.ResponseWriter) {
	var allDevices = s.interactor.GetDevices()
	var response = map[string]any{
		"devices": []models.AdbDevice{},
	}
	if len(allDevices) > 0 {
		response["devices"] = allDevices
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}
