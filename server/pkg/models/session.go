package models

// Session statuses
const (
	StatusIdle    = "idle"
	StatusClosed  = "closed"
	StatusRunning = "running %s on %s"
	StatusError   = "unable to find parameter %s with value %s in script %s"
)

// Session ...
type Session struct {
	ServerPort  int
	VideoPort   int
	CVPort      int
	ControlPort int
	Query       []string
	Status      string
	DoneCh      chan struct{}
}
