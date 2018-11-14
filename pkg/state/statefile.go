package state

import (
	"fmt"
	"github.com/function61/pi-security-module/pkg/crypto"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventlog"
)

const (
	statefilePath = "state.json"
	logfilePath   = "events.log"
)

type State struct {
	masterPassword string
	macSigningKey  string // derived from masterPassword
	sealed         bool
	State          *Statefile
	EventLog       *eventlog.EventLog
	S3ExportBucket string
	S3ExportApiKey string
	S3ExportSecret string
}

func NewTesting() *State {
	s := &State{
		masterPassword: "",
		State:          NewStatefile(),
		sealed:         false,
	}
	s.EventLog = eventlog.NewForTesting(func(event domain.Event) error {
		return domain.DispatchEvent(event, s)
	})
	return s
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
		return domain.DispatchEvent(event, s)
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

func (s *State) GetMacSigningKey() string {
	return s.macSigningKey
}

func (s *State) SetMasterPassword(password string) {
	s.masterPassword = password

	// FIXME: if we scan entire event log at startup, and there's 100x
	// "master password changed" events, that's going to yield N amount of calls to here
	// and due to nature of a KDFs are designed to be slow, that'd be real slow
	s.macSigningKey = fmt.Sprintf("%x", crypto.DeriveKey100k(
		[]byte(s.masterPassword),
		[]byte("macSalt")))
}

func (s *State) IsUnsealed() bool {
	return !s.sealed
}

func (s *State) SetSealed(sealed bool) {
	s.sealed = sealed
}
