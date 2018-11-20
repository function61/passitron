package command

import (
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/state"
)

type Ctx struct {
	State *state.State
	Meta  domain.EventMeta

	RemoteAddr string
	UserAgent  string

	SendLoginCookieUserId string

	raisedEvents []domain.Event
}

func (c *Ctx) GetRaisedEvents() []domain.Event {
	return c.raisedEvents
}

func (c *Ctx) RaisesEvent(event domain.Event) {
	c.raisedEvents = append(c.raisedEvents, event)
}

type Command interface {
	Key() string
	Validate() error
	MiddlewareChain() string
	Invoke(*Ctx) error
}
