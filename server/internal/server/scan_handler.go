package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Scan path
const (
	ScanPath = "/scan"
)

// Scan query keys
const (
	ImagesKey = "images"
)

func (s *serverImpl) handleScanFunctions() {
	http.HandleFunc(ScanPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleScan(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleScan(w http.ResponseWriter, r *http.Request) {
	var serial = r.URL.Query().Get(SerialKey)
	if serial == "" {
		http.Error(w, `"serial" query needed`, http.StatusBadRequest)
		return
	}

	var images = []string{}
	var imagesParam = strings.TrimSpace(r.URL.Query().Get(ImagesKey))
	if imagesParam != "" {
		for _, name := range strings.Split(imagesParam, ",") {
			images = append(images, strings.TrimSpace(name))
		}
	}

	var locale = r.URL.Query().Get(LocaleKey)
	landmarks, err := s.interactor.Scan(serial, images, locale, s.serverProps.SocketPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response = map[string]any{
		"landmarks": landmarks,
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}
