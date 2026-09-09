package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type apiClient struct {
	baseURL    string
	client     *http.Client
	pingClient *http.Client
	startMu    sync.Mutex
}

func newAPIClient(baseURL string) *apiClient {
	return &apiClient{
		baseURL:    baseURL,
		client:     &http.Client{Timeout: 30 * time.Second},
		pingClient: &http.Client{Timeout: pingTimeout},
	}
}

func (c *apiClient) request(method string, path string, reqBody io.Reader) ([]byte, error) {
	body, err := c.send(method, path, reqBody)
	if err == nil {
		return body, nil
	}
	started, recoverErr := c.recoverServer()
	if recoverErr != nil {
		return nil, recoverErr
	}
	if !started {
		return nil, err
	}
	if err := rewind(reqBody); err != nil {
		return nil, err
	}
	return c.send(method, path, reqBody)
}

func rewind(reqBody io.Reader) error {
	seeker, ok := reqBody.(io.Seeker)
	if !ok {
		return nil
	}
	_, err := seeker.Seek(0, io.SeekStart)
	return err
}

func (c *apiClient) send(method string, path string, reqBody io.Reader) ([]byte, error) {
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

func (c *apiClient) pingServer() error {
	_, err := c.request(http.MethodGet, "/ping", nil)
	return err
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
	if len(images) > 0 {
		query.Set("images", strings.Join(images, ","))
	}
	if locale != "" {
		query.Set("locale", locale)
	}

	var path = "/devices/" + url.PathEscape(serial) + "/scan"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.request(http.MethodGet, path, nil)
	return string(body), err
}

func (c *apiClient) queueSteps(serial string, steps []stepInput) error {
	fillStepDefaults(steps)
	body, err := json.Marshal(steps)
	if err != nil {
		return err
	}
	_, err = c.request(
		http.MethodPost,
		"/devices/"+url.PathEscape(serial)+"/queue_steps",
		bytes.NewReader(body),
	)
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
	fillStepDefaults(route.Steps)
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

func (c *apiClient) runRoute(serial string, name string, startID int) error {
	var query = url.Values{}
	query.Set("serial", serial)
	query.Set("name", name)
	if startID > 0 {
		query.Set("start_id", strconv.Itoa(startID))
	}

	_, err := c.request(http.MethodGet, "/run_route?"+query.Encode(), nil)
	return err
}

func fillStepDefaults(steps []stepInput) {
	for index := range steps {
		step := &steps[index]
		if step.Timeout == nil {
			step.Timeout = intPtr(defaultTimeoutMs)
		}
		if step.Delay != nil {
			continue
		}
		if index == 0 {
			step.Delay = intPtr(firstStepDelayMs)
			continue
		}
		step.Delay = intPtr(defaultDelayMs)
	}
}

func intPtr(value int) *int {
	return &value
}
