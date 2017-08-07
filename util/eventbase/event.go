package eventbase

import (
	"github.com/function61/pi-security-module/util/cryptorandombytes"
	"time"
)

func RandomId() string {
	return cryptorandombytes.Hex(4)
}

func NewEvent() Event {
	return Event{time.Now()}
}

func NewEventWithTimestamp(timestamp time.Time) Event {
	return Event{timestamp}
}

type Event struct {
	Timestamp time.Time
}
