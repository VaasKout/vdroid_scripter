package server

import (
	"android_vision_scripter/pkg/core/file"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

const (
	Library       = "/library"
	SaveImage     = "/save_image"
	SaveAction    = "/save_action"
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

	http.HandleFunc(SaveImage, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveImage(w, r)
			return
		}
		http.Error(w, "use POST method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(SaveAction, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveAction(w, r)
			return
		}
		http.Error(w, "use POST method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(ImagesByName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleDeleteImage(w, r)
			return
		}
		http.Error(w, "use DELETE method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(ActionsByName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleDeleteAction(w, r)
			return
		}
		http.Error(w, "use DELETE method", http.StatusMethodNotAllowed)
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
		Serial    string           `json:"serial"`
		Rectangle models.Rectangle `json:"rectangle"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Serial == "" ||
		data.Rectangle.IsEmpty() || !file.ValidName(data.Rectangle.Label) {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	saved := s.interactor.SaveImage(data.Serial, &data.Rectangle)
	if saved {
		s.sendStatusOk(w)
		return
	}

	http.Error(w, "Something went wrong", http.StatusInternalServerError)
}

func (s *serverImpl) handleSaveAction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data = &models.Action{}
	err := json.NewDecoder(r.Body).Decode(data)
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

func (s *serverImpl) handleDeleteAction(w http.ResponseWriter, r *http.Request) {
	var name = r.PathValue(NameKey)
	deleted := s.interactor.DeleteAction(name)
	if deleted {
		s.sendStatusOk(w)
		return
	}

	http.Error(w, "action not found", http.StatusNotFound)
}
