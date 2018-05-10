package state

import (
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
)

type WrappedSecret struct {
	Secret             apitypes.Secret
	SshPrivateKey      string
	OtpProvisioningUrl string
}

type WrappedAccount struct {
	Account apitypes.Account
	Secrets []WrappedSecret
}

type Statefile struct {
	WrappedAccounts []WrappedAccount
	Folders         []apitypes.Folder
	AuditLog        []apitypes.AuditlogEntry
}

const maxAuditLogEntries = 10

func (s *Statefile) Audit(message string, meta *domain.EventMeta) {
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
	}
}
