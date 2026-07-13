// Package server ...
package server

import (
	"android_vision_scripter/config"
	"android_vision_scripter/internal/usecases"
	"android_vision_scripter/pkg/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Query keys
const (
	SerialKey = "serial"
	NameKey   = "name"
)

// Multipart consts
const (
	ImagesFormField = "image"
)

// API ...
type API interface {
	ListenAndServe()
}

type serverImpl struct {
	interactor  usecases.Interactor
	serverProps *config.ServerProps
	logger      *logger.Logger
}

// New instance of API
func New(
	interactor usecases.Interactor,
	serverProps *config.ServerProps,
	logger *logger.Logger,
) API {
	return &serverImpl{
		interactor:  interactor,
		serverProps: serverProps,
		logger:      logger,
	}
}

func (s *serverImpl) ListenAndServe() {
	s.handleDeviceFunctions()
	s.handleScriptsFunctions()
	s.handleKeyboardFunctions()
	s.handleSocketFunctions()

	s.logger.Info(
		fmt.Sprintf("START LISTENING AT PORT %s", s.serverProps.Port),
	)
	err := http.ListenAndServe(s.serverProps.Port, nil)
	if err != nil {
		panic(err.Error())
	}
}

func (s *serverImpl) sendStatusOk(w http.ResponseWriter) {
	var response = map[string]string{
		"status": "ok",
	}
	s.setHeaders(w)
	json.NewEncoder(w).Encode(response)
}

func (s *serverImpl) handleImageResponse(w http.ResponseWriter, imagePath string) {
	imgFile, err := os.Open(imagePath)
	if err != nil {
		http.Error(w, "cannot open image", http.StatusInternalServerError)
		return
	}
	defer imgFile.Close()

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	defer writer.Close()

	// Create a form file field
	part, err := writer.CreateFormFile(ImagesFormField, filepath.Base(imagePath))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy file data into the part
	io.Copy(part, imgFile)
	w.Header().Set("Content-Type", writer.FormDataContentType())
	w.Write(buffer.Bytes())
}

func (s *serverImpl) setHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}

func (s *serverImpl) logURL(r *http.Request) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	fullURL := scheme + "://" + r.Host + r.RequestURI
	s.logger.Info(fullURL)
}
