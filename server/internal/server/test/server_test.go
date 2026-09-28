// Package test ...
package test

import (
	"android_vision_scripter/pkg/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Common test contants...
const (
	LocalURL   = "http://127.0.0.1:8080"
	TestSerial = "emulator-5554" // serial of the device to test
)

func makeHTTPRequest(
	url string,
	method string,
	body []byte,
	data any,
) {
	logAPI := logger.New(logger.INFO, true)
	logAPI.Info(fmt.Sprintf("Starting Request: %s - %s", method, url))
	logAPI.Info(fmt.Sprintf("Body: %s", string(body)))

	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		logAPI.Error(fmt.Sprintf("error due creating request: %s", err))
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		logAPI.Error(fmt.Sprintf("error due executing request: %s", err))
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		logAPI.Error(err.Error())
		return
	}
	if response.StatusCode != http.StatusOK {
		logAPI.Error(fmt.Sprintf("Status: %s, Body: %s", response.Status, string(responseBody)))
		return
	}
	logAPI.Info(fmt.Sprintf("End of request: %s - %s: %s", method, url, response.Status))

	if text, ok := data.(*string); ok {
		*text = string(responseBody)
		return
	}
	json.Unmarshal(responseBody, &data)
}
