package models

const (
	StatusIdle         = "idle"
	StatusClosed       = "closed"
	StatusRunningStep  = "running %s"
	StatusRunningRoute = "running route %s -> %s: %s"
	StatusError        = "unable to find %s %s on screen"
	StatusLostRoute    = "lost on route %s -> %s: node %s not confirmed"
)

type Route struct {
	From string
	To   string
}

type QueueItem struct {
	Step  *Step
	Route *Route
}

type Session struct {
	ServerPort  int
	VideoPort   int
	CVPort      int
	ControlPort int
	Query       []QueueItem
	Status      string
	DoneCh      chan struct{}
}
