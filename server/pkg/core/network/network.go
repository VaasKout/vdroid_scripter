// Package network ...
package network

import (
	"android_vision_scripter/pkg/logger"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client ...
type Client interface {
	DownloadFile(r *HTTPRequest, filePath string) error
}

type clientImpl struct {
	api    *http.Client
	logAPI *logger.Logger
}

// New instance of Client
func New(
	logAPI *logger.Logger,
) Client {
	api := &http.Client{
		Timeout: time.Second * 30,
	}

	return &clientImpl{
		api:    api,
		logAPI: logAPI,
	}
}

func (c *clientImpl) DownloadFile(r *HTTPRequest, filePath string) error {
	if r == nil || r.URL == "" {
		return errors.New("empty request")
	}

	if r.Body == nil {
		r.Body = bytes.NewBuffer([]byte(""))
	}

	var client = c.api
	if r.Insecure {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}

		tr := &http.Transport{
			TLSClientConfig: tlsConfig,
		}

		client = &http.Client{Transport: tr}
	}

	if r.LogReq {
		c.logAPI.Info("-------")
		c.logAPI.Info(fmt.Sprintf("Starting Request: %s - %s", r.Method, r.URL))
	}

	request, err := http.NewRequest(r.Method, r.URL, r.Body)
	if err != nil {
		var requestErr = errors.New("error due creating request: " + err.Error())
		if r.LogReq {
			c.logAPI.Error(fmt.Sprintf("ERROR: %s", requestErr))
		}
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	if r.Headers != nil {
		for k, v := range r.Headers {
			request.Header.Set(k, v)
		}
	}

	response, err := client.Do(request)
	if err != nil {
		var requestErr = errors.New("error due executing request: " + err.Error())
		if r.LogReq {
			c.logAPI.Error(fmt.Sprintf("ERROR: %s", requestErr))
		}
		return err
	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		var err = fmt.Errorf("failed to download %s", r.URL)
		c.logAPI.Error(err.Error())
		return err
	}

	c.logAPI.Info(fmt.Sprintf("Downloaded %s", r.URL))
	out, err := os.Create(filePath)
	if err != nil {
		c.logAPI.Error(err.Error())
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		c.logAPI.Error(err.Error())
		return err
	}

	c.logAPI.Info(fmt.Sprintf("file successfully downloaded: %s ✅", filePath))
	return nil
}
