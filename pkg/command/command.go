package command

import (
	"github.com/function61/pi-security-module/pkg/event"
	"github.com/function61/pi-security-module/pkg/state"
)

type Ctx struct {
	State *state.State
	Meta  event.EventMeta

	RemoteAddr string
	UserAgent  string

	SendLoginCookieUserId string

	raisedEvents []event.Event
}

func (c *Ctx) GetRaisedEvents() []event.Event {
	return c.raisedEvents
}

func (c *Ctx) RaisesEvent(event event.Event) {
	c.raisedEvents = append(c.raisedEvents, event)
}

type Command interface {
	Key() string
	Validate() error
	MiddlewareChain() string
	Invoke(*Ctx) error
}
