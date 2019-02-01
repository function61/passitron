package state

import (
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventkit/event"
	"time"
)

type UserStorage struct {
	SensitiveUser   SensitiveUser
	WrappedAccounts []WrappedAccount
	Folders         []apitypes.Folder
	U2FTokens       []*U2FToken
}

func NewUserStorage(sensitiveUser SensitiveUser) *UserStorage {
	rootFolder := apitypes.Folder{
		Id:       domain.RootFolderId,
		ParentId: "",
		Name:     domain.RootFolderName,
	}

	return &UserStorage{
		SensitiveUser:   sensitiveUser,
		WrappedAccounts: []WrappedAccount{},
		Folders:         []apitypes.Folder{rootFolder},
		U2FTokens:       []*U2FToken{},
	}
}

type Statefile struct {
	UserScope map[string]*UserStorage // keyed by id
	AuditLog  []apitypes.AuditlogEntry
}

func NewStatefile() *Statefile {
	return &Statefile{
		UserScope: map[string]*UserStorage{},
		AuditLog:  []apitypes.AuditlogEntry{},
	}
}

type WrappedSecret struct {
	Secret             apitypes.Secret
	SshPrivateKey      string
	OtpProvisioningUrl string
	KeylistKeys        []apitypes.SecretKeylistKey
}

// FIXME: this has the same name as with apitypes.WrappedAccount
type WrappedAccount struct {
	Account apitypes.Account
	Secrets []WrappedSecret
}

type U2FToken struct {
	Name             string
	EnrolledAt       time.Time
	KeyHandle        string
	RegistrationData string
	ClientData       string
	Version          string
	Counter          uint32
}

type SensitiveUser struct {
	User         apitypes.User
	AccessToken  string // stores only the latest. TODO: support multiple
	PasswordHash string
}

const maxAuditLogEntries = 30

func (s *Statefile) Audit(message string, meta *event.EventMeta) {
	entry := apitypes.AuditlogEntry{
		Timestamp: meta.Timestamp,
		UserId:    meta.UserId,
		Message:   message,
	}

	high := len(s.AuditLog)
	if high > maxAuditLogEntries-1 {
		high = maxAuditLogEntries - 1
	}

	s.AuditLog = append(
		[]apitypes.AuditlogEntry{entry},
		s.AuditLog[0:high]...)
}
