package server

import (
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

// Script patterns...
const (
	Scripts       = "/scripts"
	ScriptsByName = Scripts + "/{" + NodeKey + "}" + "/{" + NameKey + "}"
	RunScript     = "/devices/{" + SerialKey + "}" + ScriptsByName + "/run"

	SaveScript    = "/save_script"
	SaveRectangle = "/save_rectangle"
	FindText      = "/devices/{" + SerialKey + "}/find_text"
)

// Script query keys
const (
	TextKey = "text"
)

func (s *serverImpl) handleScriptsFunctions() {
	http.HandleFunc(Scripts, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetScriptList(w, r)
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

func (s *serverImpl) handleGetScriptList(w http.ResponseWriter, r *http.Request) {
	names, err := s.interactor.GetScriptNames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.setHeaders(w)
	json.NewEncoder(w).Encode(names)
}

func (s *serverImpl) handleGetScript(w http.ResponseWriter, r *http.Request) {
	var node = r.PathValue(NodeKey)
	if node == "" {
		http.Error(w, `"node" query needed`, http.StatusBadRequest)
		return
	}
	var name = r.PathValue(NameKey)
	if name == "" {
		http.Error(w, `"name" query needed`, http.StatusBadRequest)
		return
	}

	script, err := s.interactor.GetScript(node, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.setHeaders(w)
	json.NewEncoder(w).Encode(script)
}

func (s *serverImpl) handleDeleteScript(w http.ResponseWriter, r *http.Request) {
	var node = r.PathValue(NodeKey)
	if node == "" {
		http.Error(w, `"node" query needed`, http.StatusBadRequest)
		return
	}
	var name = r.PathValue(NameKey)
	if name == "" {
		http.Error(w, `"serial" and "name" queries needed`, http.StatusBadRequest)
		return
	}
	s.interactor.DeleteScript(node, name)
	s.sendStatusOk(w)
}

func (s *serverImpl) handleRunScript(w http.ResponseWriter, r *http.Request) {
	var serial = r.PathValue(SerialKey)
	var node = r.PathValue(NodeKey)
	var name = r.PathValue(NameKey)

	err := s.interactor.RunScript(serial, node, name, s.serverProps.SocketPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendStatusOk(w)
}

func (s *serverImpl) handleSaveRectangle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data struct {
		Serial    string           `json:"serial"`
		Rectangle models.Rectangle `json:"rectangle"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Serial == "" {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	saved := s.interactor.SaveZone(data.Serial, &data.Rectangle)
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
	if err != nil || data.Node == "" || data.Name == "" {
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
