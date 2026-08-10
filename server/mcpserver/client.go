package main

import (
	"android_vision_scripter/pkg/models"
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

func newAPIClient(baseURL string) *apiClient {
	return &apiClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *apiClient) request(method string, path string) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURL+path, nil)
	if err != nil {
		return nil, err
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
	body, err := c.request(http.MethodGet, "/devices")
	return string(body), err
}

func (c *apiClient) getLocations() ([]string, error) {
	body, err := c.request(http.MethodGet, "/locations")
	if err != nil {
		return nil, err
	}
	var locations []string
	err = json.Unmarshal(body, &locations)
	return locations, err
}

func (c *apiClient) getLocationScripts(location string) ([]string, error) {
	body, err := c.request(http.MethodGet, "/locations/"+url.PathEscape(location))
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
	body, err := c.request(http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	var script = &models.Script{}
	err = json.Unmarshal(body, script)
	return script, err
}

func (c *apiClient) queueScript(serial string, location string, name string) error {
	var path = fmt.Sprintf(
		"/locations/%s/%s/run?serial=%s",
		url.PathEscape(location),
		url.PathEscape(name),
		url.QueryEscape(serial),
	)
	_, err := c.request(http.MethodGet, path)
	return err
}

func (c *apiClient) openSession(serial string) (string, error) {
	body, err := c.request(http.MethodPost, "/devices/"+url.PathEscape(serial)+"/session")
	return string(body), err
}

func (c *apiClient) closeSession(serial string) error {
	_, err := c.request(http.MethodDelete, "/devices/"+url.PathEscape(serial)+"/session")
	return err
}

func (c *apiClient) getSessionStatus(serial string) (string, error) {
	body, err := c.request(http.MethodGet, "/devices/"+url.PathEscape(serial)+"/session")
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
