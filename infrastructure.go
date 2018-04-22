package main

import (
	"github.com/function61/pi-security-module/domain"
	"github.com/function61/pi-security-module/state"
)

type Ctx struct {
	State *state.State
	// UserId    string
	EventMeta domain.EventMeta

	// loginCookie  string
	raisedEvents []domain.Event
}

func (c *Ctx) GetRaisedEvents() []domain.Event {
	return c.raisedEvents
}

/*
func (c *Ctx) SetLoginCookie(value string) {
	c.loginCookie = value
}
*/

func (c *Ctx) RaisesEvent(event domain.Event) {
	c.raisedEvents = append(c.raisedEvents, event)
}

type Command interface {
	Validate() error
	Invoke(*Ctx) error
}
