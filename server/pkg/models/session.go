package models

const (
	StatusIdle        = "idle"
	StatusClosed      = "closed"
	StatusRunningStep = "running %s"
	StatusError       = "unable to find %s %s on screen"
)

type Session struct {
	ServerPort  int
	VideoPort   int
	CVPort      int
	ControlPort int
	Query       []Step
	Status      string
	DoneCh      chan struct{}
}
