package eventlog

import (
	"github.com/function61/pi-security-module/pkg/eventkit/event"
)

type Log interface {
	Append(events []event.Event) error
}
