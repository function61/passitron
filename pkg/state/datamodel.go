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
	EnrolledAt       time.Time
	KeyHandle        string
	RegistrationData string
	ClientData       string
	Version          string
	Counter          uint32
}

type Statefile struct {
	WrappedAccounts []WrappedAccount
	Folders         []apitypes.Folder
	AuditLog        []apitypes.AuditlogEntry
	U2FTokens       map[string]*U2FToken
}

const maxAuditLogEntries = 10

func (s *Statefile) Audit(message string, meta *event.EventMeta) {
	entry := apitypes.AuditlogEntry{
		Timestamp: meta.Timestamp,
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
		WrappedAccounts: []WrappedAccount{},
		Folders:         []apitypes.Folder{rootFolder},
		AuditLog:        []apitypes.AuditlogEntry{},
		U2FTokens:       map[string]*U2FToken{},
	}
}
