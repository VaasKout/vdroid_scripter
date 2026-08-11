package main

import (
	"android_vision_scripter/pkg/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type apiClient struct {
	baseURL string
	client  *http.Client
}

type scriptRef struct {
	Location string `json:"location"`
	Name     string `json:"name"`
}

func newAPIClient(baseURL string) *apiClient {
	return &apiClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *apiClient) request(method string, path string, reqBody io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("%s %s failed (%d): %s", method, path, resp.StatusCode, string(body))
	}
	return body, nil
}

func (c *apiClient) getDevices() (string, error) {
	body, err := c.request(http.MethodGet, "/devices", nil)
	return string(body), err
}

func (c *apiClient) getLocations() ([]string, error) {
	body, err := c.request(http.MethodGet, "/locations", nil)
	if err != nil {
		return nil, err
	}
	var locations []string
	err = json.Unmarshal(body, &locations)
	return locations, err
}

func (c *apiClient) getLocationScripts(location string) ([]string, error) {
	body, err := c.request(http.MethodGet, "/locations/"+url.PathEscape(location), nil)
	if err != nil {
		return nil, err
	}
	var names []string
	err = json.Unmarshal(body, &names)
	return names, err
}

func (c *apiClient) getScript(location string, name string) (*models.Script, error) {
	var path = fmt.Sprintf(
		"/locations/%s/%s",
		url.PathEscape(location),
		url.PathEscape(name),
	)
	body, err := c.request(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var script = &models.Script{}
	err = json.Unmarshal(body, script)
	return script, err
}

func (c *apiClient) queueScript(serial string, location string, name string) error {
	var payload = struct {
		Serial  string      `json:"serial"`
		Scripts []scriptRef `json:"scripts"`
	}{
		Serial:  serial,
		Scripts: []scriptRef{{Location: location, Name: name}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = c.request(http.MethodPost, "/run_scripts", bytes.NewReader(body))
	return err
}

func (c *apiClient) openSession(serial string) (string, error) {
	body, err := c.request(http.MethodPost, "/devices/"+url.PathEscape(serial)+"/session", nil)
	return string(body), err
}

func (c *apiClient) closeSession(serial string) error {
	_, err := c.request(http.MethodDelete, "/devices/"+url.PathEscape(serial)+"/session", nil)
	return err
}

func (c *apiClient) getSessionStatus(serial string) (string, error) {
	body, err := c.request(http.MethodGet, "/devices/"+url.PathEscape(serial)+"/session", nil)
	if err != nil {
		return "", err
	}
	var response map[string]string
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}
	return response["status"], nil
}
