package network

import "bytes"

// HTTPRequest ...
type HTTPRequest struct {
	URL      string
	Body     *bytes.Buffer
	Headers  map[string]string
	Method   string
	LogReq   bool
	Insecure bool
}
