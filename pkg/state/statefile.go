package state

import (
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventkit/event"
	"time"
)

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
	UserId           string
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

type Statefile struct {
	Users           map[string]SensitiveUser // keyed by id
	WrappedAccounts []WrappedAccount
	Folders         []apitypes.Folder
	AuditLog        []apitypes.AuditlogEntry
	U2FTokens       map[string]*U2FToken
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

func NewStatefile() *Statefile {
	rootFolder := apitypes.Folder{
		Id:       domain.RootFolderId,
		ParentId: "",
		Name:     domain.RootFolderName,
	}

	return &Statefile{
		Users:           map[string]SensitiveUser{},
		WrappedAccounts: []WrappedAccount{},
		Folders:         []apitypes.Folder{rootFolder},
		AuditLog:        []apitypes.AuditlogEntry{},
		U2FTokens:       map[string]*U2FToken{},
	}
}
