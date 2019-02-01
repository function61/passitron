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

	s.DB.UserScope[e.Meta().UserId].WrappedAccounts = append(s.DB.UserScope[e.Meta().UserId].WrappedAccounts, wrappedAccount)

	return nil
}

func (s *AppState) ApplySessionSignedIn(e *domain.SessionSignedIn) error {
	s.DB.Audit(fmt.Sprintf("Signed in with IP %s with %s", e.IpAddress, e.UserAgent), e.Meta())

	return nil
}

func (s *AppState) ApplyAccountDeleted(e *domain.AccountDeleted) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			us.WrappedAccounts = append(
				us.WrappedAccounts[:idx],
				us.WrappedAccounts[idx+1:]...)
			return nil
		}
	}

	return ewrap("ApplyAccountDeleted", errRecordNotFound)
}

func (s *AppState) ApplyAccountRenamed(e *domain.AccountRenamed) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Title = e.Title
			us.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountRenamed", errRecordNotFound)
}

func (s *AppState) ApplyAccountMoved(e *domain.AccountMoved) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.FolderId = e.NewParentFolder
			us.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountMoved", errRecordNotFound)
}

func (s *AppState) ApplyAccountDescriptionChanged(e *domain.AccountDescriptionChanged) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Description = e.Description
			us.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountDescriptionChanged", errRecordNotFound)
}

func (s *AppState) ApplyAccountUrlChanged(e *domain.AccountUrlChanged) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Url = e.Url
			us.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountUrlChanged", errRecordNotFound)
}

func (s *AppState) ApplyAccountSecretNoteAdded(e *domain.AccountSecretNoteAdded) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
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
			us.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountSecretNoteAdded", errRecordNotFound)
}

func (s *AppState) ApplyAccountOtpTokenAdded(e *domain.AccountOtpTokenAdded) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
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
			us.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountOtpTokenAdded", errRecordNotFound)
}

func (s *AppState) ApplyAccountPasswordAdded(e *domain.AccountPasswordAdded) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
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
			us.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountPasswordAdded", errRecordNotFound)
}

func (s *AppState) ApplyAccountKeylistAdded(e *domain.AccountKeylistAdded) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
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
			us.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountKeylistAdded", errRecordNotFound)
}

func (s *AppState) ApplyAccountSecretDeleted(e *domain.AccountSecretDeleted) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for accountIdx, wacc := range us.WrappedAccounts {
		if wacc.Account.Id == e.Account {
			for secretIdx, secret := range wacc.Secrets {
				if secret.Secret.Id == e.Secret {
					wacc.Secrets = append(
						wacc.Secrets[:secretIdx],
						wacc.Secrets[secretIdx+1:]...)
				}
			}
			us.WrappedAccounts[accountIdx] = wacc
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
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
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
			us.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountSshKeyAdded", errRecordNotFound)
}

func (s *AppState) ApplyAccountUsernameChanged(e *domain.AccountUsernameChanged) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, wacc := range us.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Username = e.Username
			us.WrappedAccounts[idx] = wacc
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

	us := s.DB.UserScope[e.Meta().UserId]
	us.Folders = append(us.Folders, newFolder)
	return nil
}

func (s *AppState) ApplyAccountFolderDeleted(e *domain.AccountFolderDeleted) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, folder := range us.Folders {
		if folder.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			us.Folders = append(
				us.Folders[:idx],
				us.Folders[idx+1:]...)
			return nil
		}
	}

	return ewrap("ApplyAccountFolderDeleted", errRecordNotFound)
}

func (s *AppState) ApplyAccountFolderMoved(e *domain.AccountFolderMoved) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, acc := range us.Folders {
		if acc.Id == e.Id {
			acc.ParentId = e.ParentId
			us.Folders[idx] = acc
			return nil
		}
	}

	return ewrap("ApplyAccountFolderMoved", errRecordNotFound)
}

func (s *AppState) ApplyAccountFolderRenamed(e *domain.AccountFolderRenamed) error {
	us := s.DB.UserScope[e.Meta().UserId]
	for idx, acc := range us.Folders {
		if acc.Id == e.Id {
			acc.Name = e.Name
			us.Folders[idx] = acc
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
	s.DB.UserScope[e.Id] = NewUserStorage(SensitiveUser{
		User: apitypes.User{
			Id:       e.Id,
			Created:  e.Meta().Timestamp,
			Username: e.Username,
		},
	})

	return nil
}

func (s *AppState) ApplyUserPasswordUpdated(e *domain.UserPasswordUpdated) error {
	us := s.DB.UserScope[e.Meta().UserId]
	// PasswordLastChanged only reflects actual password changes, not technical ones
	if !e.AutomaticUpgrade {
		su := us.SensitiveUser
		su.PasswordHash = e.Password
		su.User.PasswordLastChanged = e.Meta().Timestamp
		us.SensitiveUser = su
	}

	return nil
}

func (s *AppState) ApplyUserAccessTokenAdded(e *domain.UserAccessTokenAdded) error {
	us := s.DB.UserScope[e.Meta().UserId]
	us.SensitiveUser.AccessToken = e.Token

	return nil
}

func (s *AppState) ApplyUserU2FTokenRegistered(e *domain.UserU2FTokenRegistered) error {
	us := s.DB.UserScope[e.Meta().UserId]
	us.U2FTokens[e.KeyHandle] = &U2FToken{
		Name:             e.Name,
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
	us := s.DB.UserScope[e.Meta().UserId]

	token := us.U2FTokens[e.KeyHandle]
	token.Counter = uint32(e.Counter)

	return nil
}

func (s *AppState) HandleUnknownEvent(e event.Event) error {
	return errors.New("unknown event: " + e.MetaType())
}
