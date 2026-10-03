package models

import "strings"

// Session statuses
const (
	StatusIdle          = "idle"
	StatusClosed        = "closed"
	StatusRecording     = "recording"
	StatusRunningPrefix = "running "
	StatusRunningStep   = StatusRunningPrefix + "%s"
	StatusError         = "unable to find %s %s on screen"
)

const (
	RecordTimeoutMs          = 5000
	RecordMaxDurationSeconds = 600
)

// Session ...
type Session struct {
	ServerPort  int
	VideoPort   int
	ControlPort int
	Query       []Step
	Status      string
	DoneCh      chan struct{}
}

// IsBusy ...
func (s *Session) IsBusy() bool {
	return strings.HasPrefix(s.Status, StatusRunningPrefix) || s.Status == StatusRecording
}
