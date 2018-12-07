package eventlog

import (
	"github.com/function61/pi-security-module/pkg/event"
)

type Log interface {
	Append(events []event.Event) error
}
