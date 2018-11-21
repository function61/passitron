package state

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/function61/pi-security-module/pkg/crypto"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventlog"
	"io/ioutil"
	"log"
	"os"
)

const (
	statefilePath = "state.json"
	logfilePath   = "events.log"
)

type State struct {
	masterPassword string
	macSigningKey  string // derived from masterPassword
	sealed         bool
	conf           *Config
	State          *Statefile
	EventLog       *eventlog.EventLog
	eventLogFile   *os.File
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

	emptyLogReader := &bytes.Buffer{}

	s.EventLog = eventlog.New(
		emptyLogReader, // no existing data in the log
		ioutil.Discard, // do not write to disk
		func(event domain.Event) error {
			return domain.DispatchEvent(event, s)
		})

	return s
}

func New() *State {
	conf, err := readConfig()
	if err != nil {
		panic(err)
	}

	eventLogFile, err := os.OpenFile(logfilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("OpenFile: %s", err.Error())
	}

	// state from the event log is computed & populated mainly under State field
	s := &State{
		masterPassword: "",
		State:          NewStatefile(),
		sealed:         true,
		conf:           conf,
		eventLogFile:   eventLogFile,
	}

	// needs to be instantiated later, because handleEvent requires access to State
	s.EventLog = eventlog.New(eventLogFile, eventLogFile, func(event domain.Event) error {
		return domain.DispatchEvent(event, s)
	})

	return s
}

func (s *State) Close() {
	if s.eventLogFile != nil {
		s.eventLogFile.Close()
	}
}

// for Keepass export
func (s *State) GetMasterPassword() string {
	return s.masterPassword
}

func (s *State) GetMacSigningKey() string {
	return s.macSigningKey
}

func (s *State) GetCsrfToken() string {
	return "TODO"
}

func (s *State) GetJwtValidationKey() []byte {
	if s.conf.JwtPublicKey == "" {
		panic(errors.New("JwtPublicKey not set"))
	}

	return []byte(s.conf.JwtPublicKey)
}

func (s *State) GetJwtSigningKey() []byte {
	if s.conf.JwtPrivateKey == "" {
		panic(errors.New("JwtPrivateKey not set"))
	}

	return []byte(s.conf.JwtPrivateKey)
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
