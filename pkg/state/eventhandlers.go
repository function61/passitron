package state

import (
	"errors"
	"fmt"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventkit/event"
)

var (
	errRecordNotFound = errors.New("record not found")
)

func ewrap(msg string, inner error) error {
	return errors.New(msg + ": " + inner.Error())
}

func (s *AppState) ApplyAccountCreated(e *domain.AccountCreated) error {
	wrappedAccount := WrappedAccount{
		Account: apitypes.Account{
			Id:       e.Id,
			Created:  e.Meta().Timestamp,
			FolderId: e.FolderId,
			Title:    e.Title,
		},
		Secrets: []WrappedSecret{},
	}

	s.DB.WrappedAccounts = append(s.DB.WrappedAccounts, wrappedAccount)

	return nil
}

func (s *AppState) ApplySessionSignedIn(e *domain.SessionSignedIn) error {
	s.DB.Audit(fmt.Sprintf("Signed in with IP %s with %s", e.IpAddress, e.UserAgent), e.Meta())

	return nil
}

func (s *AppState) ApplyAccountDeleted(e *domain.AccountDeleted) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			s.DB.WrappedAccounts = append(
				s.DB.WrappedAccounts[:idx],
				s.DB.WrappedAccounts[idx+1:]...)
			return nil
		}
	}

	return ewrap("ApplyAccountDeleted", errRecordNotFound)
}

func (s *AppState) ApplyAccountRenamed(e *domain.AccountRenamed) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Title = e.Title
			s.DB.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountRenamed", errRecordNotFound)
}

func (s *AppState) ApplyAccountMoved(e *domain.AccountMoved) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.FolderId = e.NewParentFolder
			s.DB.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountMoved", errRecordNotFound)
}

func (s *AppState) ApplyAccountDescriptionChanged(e *domain.AccountDescriptionChanged) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Description = e.Description
			s.DB.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountDescriptionChanged", errRecordNotFound)
}

func (s *AppState) ApplyAccountUrlChanged(e *domain.AccountUrlChanged) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Url = e.Url
			s.DB.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountUrlChanged", errRecordNotFound)
}

func (s *AppState) ApplyAccountSecretNoteAdded(e *domain.AccountSecretNoteAdded) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Account {
			secret := WrappedSecret{
				Secret: apitypes.Secret{
					Id:      e.Id,
					Kind:    domain.SecretKindNote,
					Created: e.Meta().Timestamp,
					Title:   e.Title,
					Note:    e.Note,
				},
			}

			wacc.Secrets = append(wacc.Secrets, secret)
			s.DB.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountSecretNoteAdded", errRecordNotFound)
}

func (s *AppState) ApplyAccountOtpTokenAdded(e *domain.AccountOtpTokenAdded) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Account {
			secret := WrappedSecret{
				Secret: apitypes.Secret{
					Id:      e.Id,
					Kind:    domain.SecretKindOtpToken,
					Created: e.Meta().Timestamp,
				},
				OtpProvisioningUrl: e.OtpProvisioningUrl,
			}

			wacc.Secrets = append(wacc.Secrets, secret)
			s.DB.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountOtpTokenAdded", errRecordNotFound)
}

func (s *AppState) ApplyAccountPasswordAdded(e *domain.AccountPasswordAdded) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Account {
			secret := WrappedSecret{
				Secret: apitypes.Secret{
					Id:       e.Id,
					Kind:     domain.SecretKindPassword,
					Created:  e.Meta().Timestamp,
					Password: e.Password,
				},
			}

			wacc.Secrets = append(wacc.Secrets, secret)
			s.DB.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountPasswordAdded", errRecordNotFound)
}

func (s *AppState) ApplyAccountKeylistAdded(e *domain.AccountKeylistAdded) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Account {
			keyItems := []apitypes.SecretKeylistKey{}

			keylistKeyExample := ""

			for _, key := range e.Keys {
				if keylistKeyExample == "" {
					keylistKeyExample = key.Key
				}

				keyItems = append(keyItems, apitypes.SecretKeylistKey{
					Key:   key.Key,
					Value: key.Value,
				})
			}

			secret := WrappedSecret{
				Secret: apitypes.Secret{
					Id:                e.Id,
					Kind:              domain.SecretKindKeylist,
					Title:             e.Title,
					Created:           e.Meta().Timestamp,
					KeylistKeyExample: keylistKeyExample,
				},
				KeylistKeys: keyItems,
			}

			wacc.Secrets = append(wacc.Secrets, secret)
			s.DB.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountKeylistAdded", errRecordNotFound)
}

func (s *AppState) ApplyAccountSecretDeleted(e *domain.AccountSecretDeleted) error {
	for accountIdx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Account {
			for secretIdx, secret := range wacc.Secrets {
				if secret.Secret.Id == e.Secret {
					wacc.Secrets = append(
						wacc.Secrets[:secretIdx],
						wacc.Secrets[secretIdx+1:]...)
				}
			}
			s.DB.WrappedAccounts[accountIdx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountSecretDeleted", errRecordNotFound)
}

func (s *AppState) ApplyAccountSecretUsed(e *domain.AccountSecretUsed) error {
	s.DB.Audit(fmt.Sprintf("Secret %s used (%s)", e.Account, e.Type), e.Meta())

	return nil
}

func (s *AppState) ApplyAccountSshKeyAdded(e *domain.AccountSshKeyAdded) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Account {
			secret := WrappedSecret{
				Secret: apitypes.Secret{
					Id:                     e.Id,
					Kind:                   domain.SecretKindSshKey,
					Created:                e.Meta().Timestamp,
					SshPublicKeyAuthorized: e.SshPublicKeyAuthorized,
				},
				SshPrivateKey: e.SshPrivateKey,
			}

			wacc.Secrets = append(wacc.Secrets, secret)
			s.DB.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountSshKeyAdded", errRecordNotFound)
}

func (s *AppState) ApplyAccountUsernameChanged(e *domain.AccountUsernameChanged) error {
	for idx, wacc := range s.DB.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Username = e.Username
			s.DB.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountUsernameChanged", errRecordNotFound)
}

func (s *AppState) ApplyAccountFolderCreated(e *domain.AccountFolderCreated) error {
	newFolder := apitypes.Folder{
		Id:       e.Id,
		ParentId: e.ParentId,
		Name:     e.Name,
	}

	s.DB.Folders = append(s.DB.Folders, newFolder)
	return nil
}

func (s *AppState) ApplyAccountFolderDeleted(e *domain.AccountFolderDeleted) error {
	for idx, folder := range s.DB.Folders {
		if folder.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			s.DB.Folders = append(
				s.DB.Folders[:idx],
				s.DB.Folders[idx+1:]...)
			return nil
		}
	}

	return ewrap("ApplyAccountFolderDeleted", errRecordNotFound)
}

func (s *AppState) ApplyAccountFolderMoved(e *domain.AccountFolderMoved) error {
	for idx, acc := range s.DB.Folders {
		if acc.Id == e.Id {
			acc.ParentId = e.ParentId
			s.DB.Folders[idx] = acc
			return nil
		}
	}

	return ewrap("ApplyAccountFolderMoved", errRecordNotFound)
}

func (s *AppState) ApplyAccountFolderRenamed(e *domain.AccountFolderRenamed) error {
	for idx, acc := range s.DB.Folders {
		if acc.Id == e.Id {
			acc.Name = e.Name
			s.DB.Folders[idx] = acc
			return nil
		}
	}

	return ewrap("ApplyAccountFolderRenamed", errRecordNotFound)
}

func (s *AppState) ApplyDatabaseUnsealed(e *domain.DatabaseUnsealed) error {
	// no-op

	return nil
}

func (s *AppState) ApplyDatabaseMasterPasswordChanged(e *domain.DatabaseMasterPasswordChanged) error {
	s.SetMasterPassword(e.Password)

	return nil
}

func (s *AppState) ApplyDatabaseS3IntegrationConfigured(e *domain.DatabaseS3IntegrationConfigured) error {
	s.S3ExportBucket = e.Bucket
	s.S3ExportApiKey = e.ApiKey
	s.S3ExportSecret = e.Secret

	return nil
}

func (s *AppState) ApplyUserCreated(e *domain.UserCreated) error {
	s.DB.Users[e.Id] = SensitiveUser{
		User: apitypes.User{
			Id:       e.Id,
			Created:  e.Meta().Timestamp,
			Username: e.Username,
		},
	}

	return nil
}

func (s *AppState) ApplyUserPasswordUpdated(e *domain.UserPasswordUpdated) error {
	// PasswordLastChanged only reflects actual password changes, not technical ones
	if !e.AutomaticUpgrade {
		u := s.DB.Users[e.User]
		u.PasswordHash = e.Password
		u.User.PasswordLastChanged = e.Meta().Timestamp
		s.DB.Users[e.User] = u
	}

	return nil
}

func (s *AppState) ApplyUserU2FTokenRegistered(e *domain.UserU2FTokenRegistered) error {
	s.DB.U2FTokens[e.KeyHandle] = &U2FToken{
		Name:             e.Name,
		UserId:           e.Meta().UserId,
		EnrolledAt:       e.Meta().Timestamp,
		KeyHandle:        e.KeyHandle,
		RegistrationData: e.RegistrationData,
		ClientData:       e.ClientData,
		Version:          e.Version,
		Counter:          0,
	}

	return nil
}

func (s *AppState) ApplyUserU2FTokenUsed(e *domain.UserU2FTokenUsed) error {
	token := s.DB.U2FTokens[e.KeyHandle]

	token.Counter = uint32(e.Counter)

	return nil
}

func (s *AppState) HandleUnknownEvent(e event.Event) error {
	return errors.New("unknown event: " + e.MetaType())
}
