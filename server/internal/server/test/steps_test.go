package test

import (
	"android_vision_scripter/internal/server"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

const (
	FindTextPath = LocalURL + server.FindText
)

func TestFindText(t *testing.T) {
	var serialPath = fmt.Sprintf("{%s}", server.SerialKey)
	var url = strings.ReplaceAll(FindTextPath, serialPath, TestSerial)
	url = fmt.Sprintf("%s?%s=%s", url, server.TextKey, "Contacts")

	var data = ""
	makeHTTPRequest(
		url,
		http.MethodGet,
		nil,
		&data,
	)
	t.Log(data)
}
