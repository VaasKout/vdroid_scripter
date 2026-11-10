package server

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
	"strings"
)

// Library paths
const (
	Library       = "/library"
	ImagesByName  = "/images/{" + NameKey + "}"
	ActionsByName = "/actions/{" + NameKey + "}"
)

func (s *serverImpl) handleLibraryFunctions() {
	http.HandleFunc(Library, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetLibrary(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(ImagesByName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveImage(w, r)
			return
		}
		if r.Method == http.MethodPatch {
			s.logURL(r)
			s.handleRename(w, r, s.interactor.RenameImage)
			return
		}
		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleDeleteImage(w, r)
			return
		}
		http.Error(w, "use POST, PATCH or DELETE method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(ActionsByName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveAction(w, r)
			return
		}
		if r.Method == http.MethodPatch {
			s.logURL(r)
			s.handleRename(w, r, s.interactor.RenameAction)
			return
		}
		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleDeleteAction(w, r)
			return
		}
		http.Error(w, "use POST, PATCH or DELETE method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleGetLibrary(w http.ResponseWriter, r *http.Request) {
	library := s.interactor.GetLibrary()
	s.setHeaders(w)
	json.NewEncoder(w).Encode(library)
}

func (s *serverImpl) handleSaveImage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data = struct {
		Serial         string           `json:"serial"`
		Rectangle      models.Rectangle `json:"rectangle"`
		DeviceSpecific bool             `json:"device_specific"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&data)
	name := r.PathValue(NameKey)
	if err != nil || data.Serial == "" || data.Rectangle.IsEmpty() || !file.ValidName(name) {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	name = strings.TrimSpace(name)
	if data.DeviceSpecific {
		name = s.interactor.GetDeviceSpecificName(data.Serial, name)
		if name == "" {
			http.Error(
				w,
				"device model and Android version are unknown, cannot name a device-specific item",
				http.StatusInternalServerError,
			)
			return
		}
	}
	data.Rectangle.Label = name

	err = s.interactor.SaveImage(data.Serial, &data.Rectangle, s.serverProps.SocketPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response = map[string]string{
		"status": "ok",
		"name":   name,
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}

func (s *serverImpl) handleSaveAction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data = &models.Action{}
	err := json.NewDecoder(r.Body).Decode(data)
	data.Name = r.PathValue(NameKey)
	if err != nil || data.IsEmpty() || !file.ValidName(data.Name) {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	saved := s.interactor.SaveAction(data)
	if saved {
		s.sendStatusOk(w)
		return
	}

	http.Error(w, "Something went wrong", http.StatusInternalServerError)
}

func (s *serverImpl) handleDeleteImage(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	deleted := s.interactor.DeleteImage(name)
	if deleted {
		s.sendStatusOk(w)
		return
	}

	http.Error(w, "image not found", http.StatusNotFound)
}

func (s *serverImpl) handleRename(
	w http.ResponseWriter,
	r *http.Request,
	rename func(from string, to string) (string, error),
) {
	defer r.Body.Close()

	var data = map[string]string{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || !file.ValidName(data["name"]) {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	name, err := rename(r.PathValue(NameKey), data["name"])
	if err != nil {
		http.Error(w, err.Error(), renameStatus(err))
		return
	}

	var response = map[string]string{
		"status": "ok",
		"name":   name,
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}

func renameStatus(err error) int {
	if strings.Contains(err.Error(), "not found in library") {
		return http.StatusNotFound
	}
	if strings.Contains(err.Error(), "exists in library") {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

func (s *serverImpl) handleDeleteAction(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	deleted := s.interactor.DeleteAction(name)
	if deleted {
		s.sendStatusOk(w)
		return
	}

	http.Error(w, "action not found", http.StatusNotFound)
}
