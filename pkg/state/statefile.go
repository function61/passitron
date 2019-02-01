package state

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/function61/gokit/logex"
	"github.com/function61/pi-security-module/pkg/crypto"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventkit/event"
	"github.com/function61/pi-security-module/pkg/eventkit/eventlog"
	"io/ioutil"
	"log"
	"os"
)

const (
	statefilePath = "state.json"
	logfilePath   = "events.log"
)

type AppState struct {
	masterPassword string
	macSigningKey  string // derived from masterPassword
	csrfToken      string // derived from masterPassword
	sealed         bool
	conf           *Config
	DB             *Statefile
	EventLog       eventlog.Log
	eventLogFile   *os.File
	S3ExportBucket string
	S3ExportApiKey string
	S3ExportSecret string
}

func NewTesting() *AppState {
	s := &AppState{
		masterPassword: "",
		DB:             NewStatefile(),
		sealed:         false,
		conf: &Config{ // don't worry, these aren't used anywhere else
			JwtPrivateKey: "-----BEGIN PRIVATE KEY-----\nMIHcAgEBBEIB2tjp2EsS8/3zluTu9BD2iO7CgSLW/4SbE3QP+agvZ4gqfX+bfUqv\nOIGJ2QXWnNUdoa959SMk16X3g/8hhV36M/CgBwYFK4EEACOhgYkDgYYABAEdq+Bc\n07oizVlgGglR3W7YaGy9X1aRQKwmz8fkGxjSnvh59rWKrRuEf/Y0YkqsvbZ57WYH\nJ6VG+zWcdGwKrsbXaAAsUs6ublzftJUDLNWhFTF3s4YzT2h3A8ClGjKhsoqRR6YC\n3U4taAsc2GqLUf+ElReqfUiCkQSHVJ2OjxNyKCAMqg==\n-----END PRIVATE KEY-----\n",
			JwtPublicKey:  "-----BEGIN PUBLIC KEY-----\nMIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQBHavgXNO6Is1ZYBoJUd1u2GhsvV9W\nkUCsJs/H5BsY0p74efa1iq0bhH/2NGJKrL22ee1mByelRvs1nHRsCq7G12gALFLO\nrm5c37SVAyzVoRUxd7OGM09odwPApRoyobKKkUemAt1OLWgLHNhqi1H/hJUXqn1I\ngpEEh1Sdjo8TciggDKo=\n-----END PUBLIC KEY-----\n",
		},
	}

	emptyLogReader := &bytes.Buffer{}

	log, err := eventlog.NewSimpleLogFile(
		emptyLogReader, // no existing data in the log
		ioutil.Discard, // do not write to disk
		func(event event.Event) error {
			return domain.DispatchEvent(event, s)
		},
		eventDeserializer,
		logex.Discard)
	if err != nil {
		panic(err)
	}

	s.EventLog = log

	return s
}

func New(logger *log.Logger) *AppState {
	conf, err := readConfig()
	if err != nil {
		panic(err)
	}

	eventLogFile, err := os.OpenFile(logfilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("OpenFile: %s", err.Error())
	}

	// state from the event log is computed & populated mainly under State field
	s := &AppState{
		masterPassword: "",
		DB:             NewStatefile(),
		sealed:         true,
		conf:           conf,
		eventLogFile:   eventLogFile,
	}

	// needs to be instantiated later, because handleEvent requires access to State
	log, err := eventlog.NewSimpleLogFile(eventLogFile, eventLogFile, func(event event.Event) error {
		return domain.DispatchEvent(event, s)
	}, eventDeserializer, logex.Prefix("SimpleLogFile", logger))
	if err != nil {
		panic(err) // TODO: return as error
	}

	s.EventLog = log

	return s
}

func (s *AppState) Close() {
	if s.eventLogFile != nil {
		s.eventLogFile.Close()
	}
}

// for Keepass export
func (s *AppState) GetMasterPassword() string {
	return s.masterPassword
}

func (s *AppState) GetMacSigningKey() string {
	return s.macSigningKey
}

// FIXME: this is relatively safe (system-wide CSRF tokens) only as long as this is a
//        single-user system
func (s *AppState) GetCsrfToken() string {
	if s.csrfToken == "" {
		panic("csrfToken not set")
	}

	return s.csrfToken
}

func (s *AppState) GetJwtValidationKey() []byte {
	if s.conf.JwtPublicKey == "" {
		panic(errors.New("JwtPublicKey not set"))
	}

	return []byte(s.conf.JwtPublicKey)
}

func (s *AppState) GetJwtSigningKey() []byte {
	if s.conf.JwtPrivateKey == "" {
		panic(errors.New("JwtPrivateKey not set"))
	}

	return []byte(s.conf.JwtPrivateKey)
}

func (s *AppState) SetMasterPassword(password string) {
	s.masterPassword = password

	// FIXME: if we scan entire event log at startup, and there's 100x
	// "master password changed" events, that's going to yield N amount of calls to here
	// and due to nature of a KDFs are designed to be slow, that'd be real slow
	s.macSigningKey = hex.EncodeToString(crypto.Pbkdf2Sha256100kDerive(
		[]byte(s.masterPassword),
		[]byte("macSalt")))

	s.csrfToken = hex.EncodeToString(crypto.Pbkdf2Sha256100kDerive(
		[]byte(s.masterPassword),
		[]byte("csrfSalt")))
}

func (s *AppState) IsUnsealed() bool {
	return !s.sealed
}

func (s *AppState) SetSealed(sealed bool) {
	s.sealed = sealed
}

func eventDeserializer(serialized string) (event.Event, error) {
	return event.Deserialize(serialized, domain.Allocators)
}
