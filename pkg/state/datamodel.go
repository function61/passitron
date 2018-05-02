package state

import (
	"github.com/function61/pi-security-module/pkg/domain"
	"time"
)

const (
	SecretKindPassword = "password"
	SecretKindOtpToken = "otp_token"
	SecretKindSshKey   = "ssh_key"
	SecretKindKeylist  = "keylist"
)

type Secret struct {
	Id                     string
	Kind                   string
	Title                  string
	Created                time.Time
	Password               string
	SshPrivateKey          string
	SshPublicKeyAuthorized string
	OtpProvisioningUrl     string
	KeylistKeys            []KeylistKey
}

type KeylistKey struct {
	Key   string
	Value string
}

// insecure account = secure account + secrets as public
type SecureAccount struct {
	Id          string
	FolderId    string
	Title       string
	Username    string
	Description string
	Created     time.Time
	secrets     []Secret
}

type InsecureAccount struct {
	Id          string
	FolderId    string
	Title       string
	Username    string
	Description string
	Created     time.Time
	Secrets     []Secret
}

func (i *InsecureAccount) ToSecureAccount() SecureAccount {
	return SecureAccount{
		Id:          i.Id,
		FolderId:    i.FolderId,
		Title:       i.Title,
		Username:    i.Username,
		Description: i.Description,
		Created:     i.Created,
		secrets:     i.Secrets,
	}
}

func (s *SecureAccount) ToInsecureAccount() InsecureAccount {
	return InsecureAccount{
		Id:          s.Id,
		FolderId:    s.FolderId,
		Title:       s.Title,
		Username:    s.Username,
		Description: s.Description,
		Created:     s.Created,
		Secrets:     s.secrets,
	}
}

type Folder struct {
	Id       string
	ParentId string
	Name     string
}

type Statefile struct {
	Accounts []InsecureAccount
	Folders  []Folder
	AuditLog []AuditLogEntry
}

const maxAuditLogEntries = 10

func (s *Statefile) Audit(message string, meta *domain.EventMeta) {
	entry := AuditLogEntry{meta.Timestamp, message}

	high := len(s.AuditLog)
	if high > maxAuditLogEntries-1 {
		high = maxAuditLogEntries - 1
	}

	s.AuditLog = append(
		[]AuditLogEntry{entry},
		s.AuditLog[0:high]...)
}

type AuditLogEntry struct {
	Timestamp time.Time
	Message   string
}

func NewStatefile() *Statefile {
	rootFolder := Folder{
		Id:       "root",
		ParentId: "",
		Name:     "root",
	}

	return &Statefile{
		Accounts: []InsecureAccount{},
		Folders:  []Folder{rootFolder},
		AuditLog: []AuditLogEntry{},
	}
}
