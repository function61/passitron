package domain

import (
	"time"
)

type Event interface {
	Meta() *EventMeta
	Type() string
	Serialize() string
}

type EventMeta struct {
	Timestamp time.Time
	UserId    string
}

func Meta(timestamp time.Time, userId string) EventMeta {
	return EventMeta{
		Timestamp: timestamp,
		UserId:    userId,
	}
}
