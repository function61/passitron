package state

import (
	"context"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/eventhorizon/pkg/ehreader"
	"github.com/function61/eventhorizon/pkg/ehreader/ehreadertest"
	"github.com/function61/eventkit/eventlog"
	"github.com/function61/gokit/cryptorandombytes"
	"log"
)

type AppState struct {
	validatedJwtConf *JwtConfig
	users            map[string]*UserStorage // keyed by id
	EventLog         eventlog.Log            // FIXME: outdated (non-stream-aware) interface
}

func New(logger *log.Logger) (*AppState, error) {
	validatedJwtConf, err := readAndValidateJwtConfig()
	if err != nil {
		return nil, err
	}

	users := map[string]*UserStorage{}
	users["2"] = newUserStorage(ehreader.TenantId("1"))

	// state from the event log is computed & populated mainly under State field
	s := &AppState{
		validatedJwtConf: validatedJwtConf,
		EventLog:         newTempAdapter(users["2"]),
		users:            users,
	}

	if err := createAdminUser("admin", "admin", s); err != nil {
		return nil, err
	}

	return s, nil
}

// lists user known user IDs
func (a *AppState) UserIds() []string {
	return []string{"2"}
}

func (a *AppState) FindUserByUsername(username string) *SensitiveUser {
	for _, userId := range a.UserIds() {
		user := a.users[userId].SensitiveUser()
		if user.User.Username == username {
			return &user
		}
	}

	return nil
}

func (a *AppState) User(id string) *UserStorage {
	return a.users[id]
}

func (a *AppState) ValidatedJwtConf() *JwtConfig {
	return a.validatedJwtConf
}

func RandomId() string {
	return cryptorandombytes.Base64UrlWithoutLeadingDash(4)
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
