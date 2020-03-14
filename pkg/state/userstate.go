package state

import (
	"context"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehclient"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/eventhorizon/pkg/ehreader"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
	"sync"
	"time"
)

const (
	stream             = "/pism"
	maxAuditLogEntries = 30
)

type InternalAccount struct {
	Account apitypes.Account // exposed to UI - the rest are not
	Secrets []InternalSecret
}

type InternalSecret struct {
	Id                     string
	created                time.Time
	Title                  string
	SshPublicKeyAuthorized string
	externalTokenKind      *domain.ExternalTokenKind
	keylistKeyExample      string
	Kind                   domain.SecretKind
	Envelope               []byte // depending on Kind: password | secret note | OTP provisioning URL | SSH key
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
	User         apitypes.User // exposed to UI - the rest are not
	AccessToken  string        // stores only the latest. TODO: support multiple
	PasswordHash string
}

type S3ExportDetails struct {
	Bucket       string
	ApiKeyId     string
	ApiKeySecret string
}

// holds all state for one user
type UserStorage struct {
	cursor          ehclient.Cursor
	mu              sync.Mutex
	sUser           *SensitiveUser
	accounts        map[string]*InternalAccount
	folders         []*apitypes.Folder
	u2FTokens       []*U2FToken
	crypto          *cryptoThingie
	auditLog        []apitypes.AuditlogEntry
	s3ExportDetails *S3ExportDetails
}

func newUserStorage(tenant ehreader.Tenant) *UserStorage {
	crypto, err := newCryptoThingie()
	if err != nil {
		panic(err)
	}

	return &UserStorage{
		cursor:   ehclient.Beginning(tenant.Stream(stream)),
		accounts: map[string]*InternalAccount{},
		folders: []*apitypes.Folder{
			{
				Id:       domain.RootFolderId,
				ParentId: "", // root is the only one with no parent
				Name:     domain.RootFolderName,
			},
		},
		u2FTokens: []*U2FToken{},
		crypto:    crypto,
		auditLog:  []apitypes.AuditlogEntry{},
	}
}

func (l *UserStorage) GetEventTypes() ehevent.Allocators {
	return domain.Types
}

func (l *UserStorage) ProcessEvents(ctx context.Context, handle ehreader.EventProcessorHandler) error {
	l.mu.Lock()
	l.mu.Unlock()

	return handle(
		l.cursor,
		func(e ehevent.Event) error { return l.processEvent(e) },
		func(commit ehclient.Cursor) error {
			l.cursor = commit
			return nil
		})
}

func (l *UserStorage) processEvent(ev ehevent.Event) error {
	switch e := ev.(type) {
	case *domain.DatabaseS3IntegrationConfigured:
		l.s3ExportDetails = &S3ExportDetails{
			Bucket:       e.Bucket,
			ApiKeyId:     e.ApiKey,
			ApiKeySecret: e.Secret,
		}
	case *domain.DatabaseMasterPasswordChanged:
		l.audit("Changed the master key", ev.Meta())
	case *domain.DatabaseUnsealed:
		l.audit("Unlocked the decryption key", ev.Meta())
	case *domain.SessionSignedIn:
		l.audit(fmt.Sprintf("Signed in with IP %s with %s", e.IpAddress, e.UserAgent), ev.Meta())
	case *domain.UserCreated:
		l.sUser = &SensitiveUser{
			User: apitypes.User{
				Id:       e.Id,
				Created:  e.Meta().Timestamp,
				Username: e.Username,
			},
		}
	case *domain.UserPasswordUpdated:
		l.sUser.PasswordHash = e.Password

		// PasswordLastChanged only reflects actual password changes, not technical ones
		if !e.AutomaticUpgrade {
			l.sUser.User.PasswordLastChanged = e.Meta().Timestamp
		}
	case *domain.UserAccessTokenAdded:
		l.sUser.AccessToken = e.Token
	case *domain.UserU2FTokenRegistered:
		l.u2FTokens = append(l.u2FTokens, &U2FToken{
			Name:             e.Name,
			EnrolledAt:       e.Meta().Timestamp,
			KeyHandle:        e.KeyHandle,
			RegistrationData: e.RegistrationData,
			ClientData:       e.ClientData,
			Version:          e.Version,
			Counter:          0,
		})
	case *domain.UserU2FTokenUsed:
		for _, token := range l.u2FTokens {
			if token.KeyHandle == e.KeyHandle {
				token.Counter = uint32(e.Counter)
			}
		}
	case *domain.AccountFolderCreated:
		l.folders = append(l.folders, &apitypes.Folder{
			Id:       e.Id,
			ParentId: e.ParentId,
			Name:     e.Name,
		})
	case *domain.AccountFolderMoved:
		for _, folder := range l.folders {
			if folder.Id == e.Id {
				folder.ParentId = e.ParentId
				break
			}
		}
	case *domain.AccountFolderRenamed:
		for _, folder := range l.folders {
			if folder.Id == e.Id {
				folder.Name = e.Name
				break
			}
		}
	case *domain.AccountFolderDeleted:
		for idx, folder := range l.folders {
			if folder.Id == e.Id {
				l.folders = append(l.folders[:idx], l.folders[idx+1:]...)
				break
			}
		}
	case *domain.AccountSecretDeleted:
		acc := l.accounts[e.Account]

		for idx, secret := range acc.Secrets {
			if secret.Id == e.Secret {
				acc.Secrets = append(acc.Secrets[:idx], acc.Secrets[idx+1:]...)
				break
			}
		}
	case *domain.AccountCreated:
		l.accounts[e.Id] = &InternalAccount{
			Account: apitypes.Account{
				Id:       e.Id,
				Created:  e.Meta().Timestamp,
				FolderId: e.FolderId,
				Title:    e.Title,
			},
			Secrets: []InternalSecret{},
		}
	case *domain.AccountUsernameChanged:
		l.accounts[e.Id].Account.Username = e.Username
	case *domain.AccountUrlChanged:
		l.accounts[e.Id].Account.Url = e.Url
	case *domain.AccountRenamed:
		l.accounts[e.Id].Account.Title = e.Title
	case *domain.AccountDescriptionChanged:
		l.accounts[e.Id].Account.Description = e.Description
	case *domain.AccountMoved:
		l.accounts[e.Id].Account.FolderId = e.NewParentFolder
	case *domain.AccountDeleted:
		delete(l.accounts, e.Id)
	case *domain.AccountSecretNoteAdded:
		acc := l.accounts[e.Account]
		acc.Secrets = append(acc.Secrets, InternalSecret{
			Id:       e.Id,
			created:  e.Meta().Timestamp,
			Title:    e.Title,
			Kind:     domain.SecretKindNote,
			Envelope: e.Note,
		})
	case *domain.AccountPasswordAdded:
		acc := l.accounts[e.Account]
		acc.Secrets = append(acc.Secrets, InternalSecret{
			Id:       e.Id,
			created:  e.Meta().Timestamp,
			Kind:     domain.SecretKindPassword,
			Envelope: e.Password,
		})
	case *domain.AccountOtpTokenAdded:
		acc := l.accounts[e.Account]
		acc.Secrets = append(acc.Secrets, InternalSecret{
			Id:       e.Id,
			created:  e.Meta().Timestamp,
			Kind:     domain.SecretKindOtpToken,
			Envelope: e.OtpProvisioningUrl,
		})
	case *domain.AccountKeylistAdded:
		acc := l.accounts[e.Account]
		acc.Secrets = append(acc.Secrets, InternalSecret{
			Id:                e.Id,
			created:           e.Meta().Timestamp,
			Kind:              domain.SecretKindKeylist,
			Title:             e.Title,
			keylistKeyExample: e.KeyExample,
			Envelope:          e.Keys,
		})
	case *domain.AccountExternalTokenAdded:
		acc := l.accounts[e.Account]
		acc.Secrets = append(acc.Secrets, InternalSecret{
			Id:                e.Id,
			created:           e.Meta().Timestamp,
			Title:             e.Description,
			Kind:              domain.SecretKindExternalToken,
			externalTokenKind: &e.Kind,
		})
	case *domain.AccountSshKeyAdded:
		acc := l.accounts[e.Account]
		acc.Secrets = append(acc.Secrets, InternalSecret{
			Id:                     e.Id,
			created:                e.Meta().Timestamp,
			SshPublicKeyAuthorized: e.SshPublicKeyAuthorized,
			Kind:                   domain.SecretKindSshKey,
			Envelope:               e.SshPrivateKey,
		})
	case *domain.AccountSecretUsed:
		l.audit(fmt.Sprintf("Account %s secret %v - %s", e.Account, e.Secrets, e.Type), ev.Meta())
	default:
		return ehreader.UnsupportedEventTypeErr(ev)
	}

	return nil
}

func (l *UserStorage) audit(message string, meta *ehevent.EventMeta) {
	entry := apitypes.AuditlogEntry{
		Timestamp: meta.Timestamp,
		Message:   message,
	}

	high := len(l.auditLog)
	if high > maxAuditLogEntries-1 {
		high = maxAuditLogEntries - 1
	}

	l.auditLog = append(
		[]apitypes.AuditlogEntry{entry},
		l.auditLog[0:high]...)
}
