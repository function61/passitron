package domain

import (
	"github.com/function61/pi-security-module/util/cryptorandombytes"
	"time"
)

type Event interface {
	Meta() *EventMeta
	MetaType() string
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

func RandomId() string {
	return cryptorandombytes.Hex(4)
}
