package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

func (c *apiClient) getLibrary() (string, error) {
	body, err := c.request(http.MethodGet, "/library", nil)
	return string(body), err
}

func (c *apiClient) scan(serial string, images []string, locale string) (string, error) {
	var query = url.Values{}
	query.Set("serial", serial)
	if len(images) > 0 {
		query.Set("images", strings.Join(images, ","))
	}
	if locale != "" {
		query.Set("locale", locale)
	}

	body, err := c.request(http.MethodGet, "/scan?"+query.Encode(), nil)
	return string(body), err
}

func (c *apiClient) queueSteps(serial string, steps []stepInput) error {
	var payload = struct {
		Serial string      `json:"serial"`
		Steps  []stepInput `json:"steps"`
	}{
		Serial: serial,
		Steps:  steps,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = c.request(http.MethodPost, "/run_steps", bytes.NewReader(body))
	return err
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

func (c *apiClient) getRoutes() (string, error) {
	body, err := c.request(http.MethodGet, "/routes", nil)
	return string(body), err
}

func (c *apiClient) getRoute(name string) (string, error) {
	body, err := c.request(http.MethodGet, "/routes/"+url.PathEscape(name), nil)
	return string(body), err
}

func (c *apiClient) saveRoute(route *saveRouteInput) error {
	body, err := json.Marshal(route)
	if err != nil {
		return err
	}
	_, err = c.request(http.MethodPost, "/routes", bytes.NewReader(body))
	return err
}

func (c *apiClient) deleteRoute(name string) error {
	_, err := c.request(http.MethodDelete, "/routes/"+url.PathEscape(name), nil)
	return err
}

func (c *apiClient) runRoute(serial string, name string) error {
	var query = url.Values{}
	query.Set("serial", serial)
	query.Set("name", name)

	_, err := c.request(http.MethodGet, "/run_route?"+query.Encode(), nil)
	return err
}
