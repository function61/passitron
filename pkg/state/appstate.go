package state

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/eventhorizon/pkg/ehreader"
	"github.com/function61/eventhorizon/pkg/ehreader/ehreadertest"
	"github.com/function61/eventkit/eventlog"
	"github.com/function61/gokit/cryptorandombytes"
	"github.com/function61/pi-security-module/pkg/crypto"
	"log"
)

func NewStatefile() *Statefile {
	return &Statefile{
		UserScope: map[string]*UserStorage{},
	}
}

type Statefile struct {
	UserScope map[string]*UserStorage // keyed by id
}

type AppState struct {
	masterPassword string
	macSigningKey  string // derived from masterPassword
	sealed         bool
	conf           *Config // only contains JwtPrivateKey, JwtPublicKey
	DB             *Statefile
	EventLog       eventlog.Log
}

// pushes appends directly to in-memory event log only meant for testing
type tempAdapter struct {
	reader *ehreader.Reader
	log    *ehreadertest.EventLog
	ctx    context.Context
}

func newTempAdapter(proc *UserStorage) *tempAdapter {
	eventLog := ehreadertest.NewEventLog()

	return &tempAdapter{
		reader: ehreader.New(proc, eventLog, nil),
		log:    eventLog,
		ctx:    context.Background(),
	}
}

func (m *tempAdapter) Append(events []ehevent.Event) error {
	serialized := []string{}
	for _, event := range events {
		serialized = append(serialized, ehevent.Serialize(event))
	}

	if _, err := m.log.Append(m.ctx, "/t-1/pism", serialized); err != nil {
		return err
	}

	return m.reader.LoadUntilRealtime(m.ctx)
}

func New(logger *log.Logger) (*AppState, error) {
	conf, err := readConfig()
	if err != nil {
		return nil, err
	}

	db := NewStatefile()
	db.UserScope["2"] = newUserStorage(ehreader.TenantId("1"))

	// state from the event log is computed & populated mainly under State field
	s := &AppState{
		masterPassword: "initpwd", // was accumulated from event log
		DB:             db,
		sealed:         true,
		conf:           conf,
		EventLog:       newTempAdapter(db.UserScope["2"]),
	}

	if err := createAdminUser("admin", "admin", s); err != nil {
		return nil, err
	}

	return s, nil
}

// for Keepass export
func (s *AppState) GetMasterPassword() string {
	return s.masterPassword
}

func (s *AppState) GetMacSigningKey() string {
	return s.macSigningKey
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
}

func (s *AppState) IsUnsealed() bool {
	return !s.sealed
}

func (s *AppState) SetSealed(sealed bool) {
	s.sealed = sealed
}

func RandomId() string {
	return cryptorandombytes.Base64UrlWithoutLeadingDash(4)
}
