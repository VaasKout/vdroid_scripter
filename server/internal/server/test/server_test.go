// Package test ...
package test

import (
	"android_vision_scripter/pkg/core/network"
	"android_vision_scripter/pkg/logger"
	"bytes"
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
	client := network.New(logAPI)
	request := &network.HTTPRequest{
		URL:     url,
		Method:  method,
		Body:    bytes.NewBuffer(body),
		LogBody: true,
		LogReq:  true,
	}

	client.MakeRequest(request, data)
}
