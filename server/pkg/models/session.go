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

// RecordDurationSeconds ...
const RecordDurationSeconds = 5

// IsRunningStatus ...
func IsRunningStatus(status string) bool {
	return strings.HasPrefix(status, StatusRunningPrefix)
}

// Session ...
type Session struct {
	ServerPort  int
	VideoPort   int
	ControlPort int
	Query       []Step
	Status      string
	DoneCh      chan struct{}
}
