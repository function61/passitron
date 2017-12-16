package state

import (
	"errors"
	"github.com/function61/pi-security-module/util/eventapplicator"
)

const (
	statefilePath = "state.json"
	logfilePath   = "events.log"
)

var Inst *State

type State struct {
	masterPassword string
	sealed         bool
	State          *Statefile
	EventLog       *eventapplicator.EventApplicator
}

func Initialize() {
	if Inst != nil {
		panic(errors.New("statefile: initialize called twice"))
	}

	ea := eventapplicator.NewEventApplicator(logfilePath)

	// state from the event log is computed & populated here
	Inst = &State{
		masterPassword: "",
		State:          NewStatefile(),
		sealed:         true,
		EventLog:       ea,
	}
}

func (s *State) Close() {
	s.EventLog.Close()

	Inst = nil
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
