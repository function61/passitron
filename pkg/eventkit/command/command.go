package command

import (
	"github.com/function61/pi-security-module/pkg/eventkit/event"
	"net/http"
)

type Ctx struct {
	Meta event.EventMeta

	RemoteAddr string
	UserAgent  string

	SetCookie *http.Cookie

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
	Invoke(ctx *Ctx, handlers interface{}) error
}

// map keyed by command name (command.Key()) values are functions that allocates a new
// specific command struct
type AllocatorMap map[string]func() Command
