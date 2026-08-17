package server

import (
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

// Devices paths
const (
	Ping             = "/ping"
	Devices          = "/devices"
	DeviceScreenshot = Devices + "/{" + SerialKey + "}/screenshot"
	FindText         = Devices + "/{" + SerialKey + "}/find_text"
)

const (
	TextKey = "text"
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

	http.HandleFunc(DeviceScreenshot, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleDeviceScreenshot(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
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

func (s *serverImpl) handleDeviceScreenshot(w http.ResponseWriter, r *http.Request) {
	var serial = r.PathValue(SerialKey)
	if serial == "" {
		http.Error(w, `"serial" param required`, http.StatusBadRequest)
		return
	}

	var withRectangles = r.URL.Query().Get(RectanglesKey) == "true"
	imagePath, rectangles, err := s.interactor.TakeScreenshot(serial, withRectangles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.handleScreenshotResponse(w, imagePath, rectangles)
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
