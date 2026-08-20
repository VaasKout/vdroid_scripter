package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
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

func (c *apiClient) getScripts() (string, error) {
	body, err := c.request(http.MethodGet, "/scripts", nil)
	return string(body), err
}

func (c *apiClient) getScript(name string) (string, error) {
	body, err := c.request(http.MethodGet, "/scripts/"+url.PathEscape(name), nil)
	return string(body), err
}

func (c *apiClient) saveScript(name string, steps []stepInput) error {
	var payload = struct {
		Name  string      `json:"name"`
		Steps []stepInput `json:"steps"`
	}{
		Name:  name,
		Steps: steps,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = c.request(http.MethodPost, "/scripts", bytes.NewReader(body))
	return err
}

func (c *apiClient) deleteScript(name string) error {
	_, err := c.request(http.MethodDelete, "/scripts/"+url.PathEscape(name), nil)
	return err
}

func (c *apiClient) runScript(serial string, name string) error {
	var params = url.Values{}
	params.Set("serial", serial)
	params.Set("name", name)
	_, err := c.request(http.MethodGet, "/run_script?"+params.Encode(), nil)
	return err
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

func (c *apiClient) getScreenshot(
	serial string,
	withRectangles bool,
) ([]byte, string, error) {
	var path = "/devices/" + url.PathEscape(serial) + "/screenshot"
	if withRectangles {
		path += "?rectangles=true"
	}

	resp, err := c.client.Get(c.baseURL + path)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("GET %s failed (%d): %s", path, resp.StatusCode, string(body))
	}

	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, "", err
	}

	var image []byte
	var rectangles string
	reader := multipart.NewReader(resp.Body, params["boundary"])
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", err
		}

		data, err := io.ReadAll(part)
		if err != nil {
			return nil, "", err
		}
		if part.FormName() == "image" {
			image = data
		}
		if part.FormName() == "rectangles" {
			rectangles = string(data)
		}
	}

	if len(image) == 0 {
		return nil, "", fmt.Errorf("screenshot response has no image")
	}
	return image, rectangles, nil
}

func (c *apiClient) saveImage(serial string, name string, rect rectangleInput) error {
	var payload = struct {
		Serial    string `json:"serial"`
		Rectangle struct {
			rectangleInput
			Label string `json:"label"`
		} `json:"rectangle"`
	}{}
	payload.Serial = serial
	payload.Rectangle.rectangleInput = rect
	payload.Rectangle.Label = name

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = c.request(http.MethodPost, "/save_image", bytes.NewReader(body))
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
