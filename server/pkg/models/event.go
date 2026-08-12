package models

const (
	ActionDown byte = 0
	ActionUp   byte = 1
)

type Event struct {
	Time int64        `json:"time"`
	Data ControlBytes `json:"data"`
}
