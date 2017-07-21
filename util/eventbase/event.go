package eventbase

import (
	"time"
)

func NewEvent() Event {
	return Event{time.Now()}
}

type Event struct {
	Timestamp time.Time
}
