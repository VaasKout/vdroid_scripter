package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
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
		client:     &http.Client{Timeout: 60 * time.Second},
		pingClient: &http.Client{Timeout: pingTimeout},
	}
}

func (c *apiClient) request(method string, path string, reqBody io.Reader) ([]byte, error) {
	body, _, err := c.requestStatus(method, path, reqBody)
	return body, err
}

func (c *apiClient) requestStatus(method string, path string, reqBody io.Reader) ([]byte, int, error) {
	body, status, err := c.send(method, path, reqBody)
	if err == nil {
		return body, status, nil
	}
	started, recoverErr := c.recoverServer()
	if recoverErr != nil {
		return nil, 0, recoverErr
	}
	if !started {
		return nil, 0, err
	}
	if err := rewind(reqBody); err != nil {
		return nil, 0, err
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

func (c *apiClient) send(method string, path string, reqBody io.Reader) ([]byte, int, error) {
	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, err
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, resp.StatusCode, fmt.Errorf("%s %s failed (%d): %s", method, path, resp.StatusCode, string(body))
	}
	return body, resp.StatusCode, nil
}

func (c *apiClient) pingServer() error {
	_, err := c.request(http.MethodGet, "/ping", nil)
	return err
}

func (c *apiClient) getJSON(path string, target any) error {
	body, err := c.request(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

type deviceInfo struct {
	Serial        string `json:"serial"`
	Brand         string `json:"brand"`
	Device        string `json:"device"`
	Locale        string `json:"locale"`
	Model         string `json:"model"`
	OsVersion     string `json:"os_version"`
	MarketingName string `json:"marketing_name"`
	ScrcpyRunning bool   `json:"scrcpy_running"`
}

type devicesResponse struct {
	Devices []deviceInfo `json:"devices"`
}

func (c *apiClient) getDevices() ([]deviceInfo, error) {
	var response devicesResponse
	err := c.getJSON("/devices", &response)
	if err != nil {
		return nil, err
	}
	return response.Devices, nil
}

type libraryResponse struct {
	Images  []string `json:"images"`
	Actions []string `json:"actions"`
}

func (c *apiClient) getLibrary() (libraryResponse, error) {
	var response libraryResponse
	err := c.getJSON("/library", &response)
	return response, err
}

type scanRectangle struct {
	LeftX   int `json:"left_x"`
	RightX  int `json:"right_x"`
	TopY    int `json:"top_y"`
	BottomY int `json:"bottom_y"`
}

type scanLandmark struct {
	Type       string        `json:"type"`
	Value      string        `json:"value"`
	Locale     string        `json:"locale,omitempty"`
	Confidence *int          `json:"confidence,omitempty"`
	Rectangle  scanRectangle `json:"rectangle"`
}

type scanResponse struct {
	Landmarks []scanLandmark `json:"landmarks"`
}

func (c *apiClient) scan(serial string, images []string, locale string) ([]scanLandmark, error) {
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

	var response scanResponse
	err := c.getJSON(path, &response)
	if err != nil {
		return nil, err
	}
	return response.Landmarks, nil
}

func (c *apiClient) capture(serial string) ([]byte, error) {
	return c.request(http.MethodGet, "/devices/"+url.PathEscape(serial)+"/capture", nil)
}

type saveImageRectangle struct {
	LeftX   int    `json:"left_x"`
	RightX  int    `json:"right_x"`
	TopY    int    `json:"top_y"`
	BottomY int    `json:"bottom_y"`
	Label   string `json:"label"`
}

type saveImageRequest struct {
	Serial    string             `json:"serial"`
	Rectangle saveImageRectangle `json:"rectangle"`
}

func (c *apiClient) saveImage(serial string, rectangle saveImageRectangle) error {
	body, err := json.Marshal(saveImageRequest{Serial: serial, Rectangle: rectangle})
	if err != nil {
		return err
	}
	_, err = c.request(http.MethodPost, "/save_image", bytes.NewReader(body))
	return err
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

func (c *apiClient) recordAction(serial string, name string) (bool, error) {
	payload, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		return false, err
	}
	var path = "/devices/" + url.PathEscape(serial) + "/record"
	_, status, err := c.requestStatus(http.MethodPost, path, bytes.NewReader(payload))
	if err != nil {
		return false, err
	}
	return status == http.StatusOK, nil
}

func (c *apiClient) closeSession(serial string) error {
	_, err := c.request(http.MethodDelete, "/devices/"+url.PathEscape(serial)+"/session", nil)
	return err
}

func (c *apiClient) getSessionStatus(serial string) (string, error) {
	var response map[string]string
	err := c.getJSON("/devices/"+url.PathEscape(serial)+"/session", &response)
	if err != nil {
		return "", err
	}
	return response["status"], nil
}

type routesResponse struct {
	Routes []string `json:"routes"`
}

func (c *apiClient) getRoutes() ([]string, error) {
	var response routesResponse
	err := c.getJSON("/routes", &response)
	if err != nil {
		return nil, err
	}
	return response.Routes, nil
}

type routeStep struct {
	ID        int             `json:"id"`
	Event     string          `json:"event"`
	Landmarks []landmarkInput `json:"landmarks"`
	Timeout   int             `json:"timeout"`
	Delay     int             `json:"delay"`
}

type routeResponse struct {
	Name  string      `json:"name"`
	Steps []routeStep `json:"steps"`
}

func (c *apiClient) getRoute(name string) (routeResponse, error) {
	var response routeResponse
	err := c.getJSON("/routes/"+url.PathEscape(name), &response)
	return response, err
}

func (c *apiClient) appendToRoute(name string, steps []stepInput) (int, error) {
	saved, err := c.savedRouteSteps(name)
	if err != nil {
		return 0, err
	}

	route := saveRouteInput{Name: name, Steps: append(saved, steps...)}
	err = c.saveRoute(&route)
	if err != nil {
		return 0, err
	}
	return len(route.Steps), nil
}

func (c *apiClient) savedRouteSteps(name string) ([]stepInput, error) {
	names, err := c.getRoutes()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(names, name) {
		return nil, nil
	}

	route, err := c.getRoute(name)
	if err != nil {
		return nil, err
	}
	return routeInputs(route.Steps), nil
}

func (s routeStep) toInput() stepInput {
	return stepInput{
		Event:     s.Event,
		Landmarks: s.Landmarks,
		Timeout:   intPtr(s.Timeout),
		Delay:     intPtr(s.Delay),
	}
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

func (c *apiClient) runRoute(in runRouteInput) error {
	var query = url.Values{}
	query.Set("serial", in.Serial)
	query.Set("name", in.Name)
	if in.StartID > 0 {
		var startID = strconv.Itoa(in.StartID)
		query.Set("start_id", startID)
	}
	if in.EndID > 0 {
		var endID = strconv.Itoa(in.EndID)
		query.Set("end_id", endID)
	}
	for _, arg := range in.Args {
		query.Add("args", arg)
	}
	if in.ArgsLocale != "" {
		query.Set("args_locale", in.ArgsLocale)
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
