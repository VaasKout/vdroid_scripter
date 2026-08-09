package server

import (
	"android_vision_scripter/internal/usecases"
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

// Script patterns...
const (
	Locations     = "/locations"
	Scripts       = Locations + "/{" + LocationKey + "}"
	ScriptsByName = Scripts + "/{" + NameKey + "}"
	RunScript     = ScriptsByName + "/run"

	SaveScript    = "/save_script"
	SaveRectangle = "/save_rectangle"
	FindText      = "/devices/{" + SerialKey + "}/find_text"
)

// Script query keys
const (
	TextKey = "text"
)

func (s *serverImpl) handleScriptsFunctions() {
	http.HandleFunc(Locations, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetLocations(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(Scripts, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetScripts(w, r)
			return
		}

		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleDeleteScript(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(ScriptsByName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetScript(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			s.logURL(r)
			s.handleDeleteScript(w, r)
			return
		}
		http.Error(w, "use another method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(RunScript, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleRunScript(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(SaveRectangle, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveRectangle(w, r)
			return
		}
		http.Error(w, "use POST method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(SaveScript, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleSaveScript(w, r)
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

func (s *serverImpl) handleGetLocations(w http.ResponseWriter, r *http.Request) {
	names, err := s.interactor.GetLocations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.setHeaders(w)
	json.NewEncoder(w).Encode(names)
}

func (s *serverImpl) handleGetScripts(w http.ResponseWriter, r *http.Request) {
	var location = r.PathValue(LocationKey)
	names, err := s.interactor.GetScriptNames(location)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.setHeaders(w)
	json.NewEncoder(w).Encode(names)
}

func (s *serverImpl) handleGetScript(w http.ResponseWriter, r *http.Request) {
	var location = r.PathValue(LocationKey)
	var name = r.PathValue(NameKey)
	if location == "" {
		http.Error(w, `"location" query needed`, http.StatusBadRequest)
		return
	}

	script, err := s.interactor.GetScript(location, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.setHeaders(w)
	json.NewEncoder(w).Encode(script)
}

func (s *serverImpl) handleDeleteScript(w http.ResponseWriter, r *http.Request) {
	var location = r.PathValue(LocationKey)
	var name = r.PathValue(NameKey)
	if location == "" {
		http.Error(w, `"location" query needed`, http.StatusBadRequest)
		return
	}

	s.interactor.DeleteScript(location, name)
	s.sendStatusOk(w)
}

func (s *serverImpl) handleRunScript(w http.ResponseWriter, r *http.Request) {
	var serial = r.URL.Query().Get(SerialKey)
	var location = r.PathValue(LocationKey)
	var name = r.PathValue(NameKey)

	if serial == "" || location == "" || name == "" {
		http.Error(w, `check queries`, http.StatusBadRequest)
		return
	}

	err := s.interactor.RunScript(serial, location, name, s.serverProps.SocketPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}

func (s *serverImpl) handleSaveRectangle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var saveZone = &usecases.SaveZoneDto{}
	err := json.NewDecoder(r.Body).Decode(saveZone)
	if err != nil || !saveZone.Valid() {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	saved := s.interactor.SaveZone(saveZone)
	if saved {
		s.sendStatusOk(w)
		return
	}

	http.Error(w, "Something went wrong", http.StatusInternalServerError)
}

func (s *serverImpl) handleSaveScript(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data = &models.Script{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Location == "" || data.Name == "" {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	saved := s.interactor.SaveScript(data)
	if saved {
		s.sendStatusOk(w)
		return
	}

	http.Error(w, "Something went wrong", http.StatusInternalServerError)
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
