package state

import (
	"errors"
	"github.com/function61/pi-security-module/util/eventlog"
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
	EventLog       *eventlog.EventLog
	S3ExportBucket string
	S3ExportApiKey string
	S3ExportSecret string
}

func Initialize() {
	if Inst != nil {
		panic(errors.New("statefile: initialize called twice"))
	}

	// state from the event log is computed & populated here
	Inst = &State{
		masterPassword: "",
		State:          NewStatefile(),
		sealed:         true,
	}

	// needs to be instantiated later, because handleEvent requires presence of "Inst"
	Inst.EventLog = eventlog.New(logfilePath, handleEvent)
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
