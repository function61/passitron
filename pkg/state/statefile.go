package state

import (
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventlog"
)

const (
	statefilePath = "state.json"
	logfilePath   = "events.log"
)

type State struct {
	masterPassword string
	sealed         bool
	State          *Statefile
	EventLog       *eventlog.EventLog
	S3ExportBucket string
	S3ExportApiKey string
	S3ExportSecret string
}

func New() *State {
	// state from the event log is computed & populated here
	s := &State{
		masterPassword: "",
		State:          NewStatefile(),
		sealed:         true,
	}

	// needs to be instantiated later, because handleEvent requires access to State
	s.EventLog = eventlog.New(logfilePath, func(event domain.Event) error {
		return handleEvent(event, s)
	})

	return s
}

func (s *State) Close() {
	s.EventLog.Close()
}

// for Keepass export
func (s *State) GetMasterPassword() string {
	return s.masterPassword
}

func (s *State) SetMasterPassword(password string) {
	s.masterPassword = password
}

func (s *State) IsUnsealed() bool {
	return !s.sealed
}

func (s *State) SetSealed(sealed bool) {
	s.sealed = sealed
}
