package server

import (
	"android_vision_scripter/pkg/models"
	"encoding/json"
	"net/http"
)

// Keyboard paths
const (
	GetKeyboard   = "/devices/{" + SerialKey + "}/keyboard"
	EditKeyboard  = "/devices/{" + SerialKey + "}/edit_keyboard"
	DeleteButton  = "/devices/{" + SerialKey + "}/delete_button"
	ResetKeyboard = "/devices/{" + SerialKey + "}/reset_keyboard"
)

// Keyboard query keys
const (
	LocaleKey    = "locale"
	UpperCaseKey = "upper_case"
)

func (s *serverImpl) handleKeyboardFunctions() {
	http.HandleFunc(GetKeyboard, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleGetKeyboard(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(EditKeyboard, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.logURL(r)
			s.handleEditKeyboard(w, r)
			return
		}
		http.Error(w, "use POST method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(ResetKeyboard, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleResetKeyboard(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})

	http.HandleFunc(DeleteButton, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.logURL(r)
			s.handleDeleteButton(w, r)
			return
		}
		http.Error(w, "use GET method", http.StatusMethodNotAllowed)
	})
}

func (s *serverImpl) handleGetKeyboard(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var serial = r.PathValue(SerialKey)
	var locale = r.URL.Query().Get(LocaleKey)

	if serial == "" {
		http.Error(w, `"serial" query needed`, http.StatusBadRequest)
		return
	}

	result := s.interactor.GetKeyboardKeys(serial, locale)
	response := map[string]any{
		"buttons": result,
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}

func (s *serverImpl) handleEditKeyboard(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var data struct {
		Serial    string           `json:"serial"`
		Locale    string           `json:"locale"`
		Name      string           `json:"name"`
		Rectangle models.Rectangle `json:"rectangle"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Serial == "" || data.Name == "" || data.Rectangle.IsEmpty() {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	result := s.interactor.EditKeyboardKey(
		data.Serial,
		data.Locale,
		data.Name,
		&data.Rectangle,
	)

	if !result {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	s.sendStatusOk(w)
}

func (s *serverImpl) handleResetKeyboard(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var serial = r.PathValue(SerialKey)
	var locale = r.URL.Query().Get(LocaleKey)
	var upperCase = r.URL.Query().Get(UpperCaseKey) == "true"

	if serial == "" {
		http.Error(w, `"serial" query needed`, http.StatusBadRequest)
		return
	}

	result := s.interactor.ResetKeyboardKeys(serial, locale, upperCase)
	response := map[string]any{
		"buttons": result,
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}

func (s *serverImpl) handleDeleteButton(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var serial = r.PathValue(SerialKey)
	var locale = r.URL.Query().Get(LocaleKey)
	var name = r.URL.Query().Get(NameKey)

	if serial == "" {
		http.Error(w, `"serial" query needed`, http.StatusBadRequest)
		return
	}

	result := s.interactor.DeleteButton(serial, locale, name)
	if !result {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	s.sendStatusOk(w)
}
