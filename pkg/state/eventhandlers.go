package state

import (
	"errors"
	"fmt"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
)

var (
	errRecordNotFound = errors.New("record not found")
)

func ewrap(msg string, inner error) error {
	return errors.New(msg + ": " + inner.Error())
}

func (s *State) ApplyAccountCreated(e *domain.AccountCreated) error {
	wrappedAccount := WrappedAccount{
		Account: apitypes.Account{
			Id:       e.Id,
			Created:  e.Meta().Timestamp,
			FolderId: e.FolderId,
			Title:    e.Title,
		},
		Secrets: []WrappedSecret{},
	}

	s.State.WrappedAccounts = append(s.State.WrappedAccounts, wrappedAccount)

	return nil
}

func (s *State) ApplyAccountDeleted(e *domain.AccountDeleted) error {
	for idx, wacc := range s.State.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			s.State.WrappedAccounts = append(
				s.State.WrappedAccounts[:idx],
				s.State.WrappedAccounts[idx+1:]...)
			return nil
		}
	}

	return ewrap("ApplyAccountDeleted", errRecordNotFound)
}

func (s *State) ApplyAccountRenamed(e *domain.AccountRenamed) error {
	for idx, wacc := range s.State.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Title = e.Title
			s.State.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountRenamed", errRecordNotFound)
}

func (s *State) ApplyAccountDescriptionChanged(e *domain.AccountDescriptionChanged) error {
	for idx, wacc := range s.State.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Description = e.Description
			s.State.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountDescriptionChanged", errRecordNotFound)
}

func (s *State) ApplyAccountOtpTokenAdded(e *domain.AccountOtpTokenAdded) error {
	for idx, wacc := range s.State.WrappedAccounts {
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
			s.State.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountOtpTokenAdded", errRecordNotFound)
}

func (s *State) ApplyAccountPasswordAdded(e *domain.AccountPasswordAdded) error {
	for idx, wacc := range s.State.WrappedAccounts {
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
			s.State.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountPasswordAdded", errRecordNotFound)
}

func (s *State) ApplyAccountKeylistAdded(e *domain.AccountKeylistAdded) error {
	for idx, wacc := range s.State.WrappedAccounts {
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
			s.State.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountKeylistAdded", errRecordNotFound)
}

func (s *State) ApplyAccountSecretDeleted(e *domain.AccountSecretDeleted) error {
	for accountIdx, wacc := range s.State.WrappedAccounts {
		if wacc.Account.Id == e.Account {
			for secretIdx, secret := range wacc.Secrets {
				if secret.Secret.Id == e.Secret {
					wacc.Secrets = append(
						wacc.Secrets[:secretIdx],
						wacc.Secrets[secretIdx+1:]...)
				}
			}
			s.State.WrappedAccounts[accountIdx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountSecretDeleted", errRecordNotFound)
}

func (s *State) ApplyAccountSecretUsed(e *domain.AccountSecretUsed) error {
	s.State.Audit(fmt.Sprintf("Secret %s used (%s)", e.Account, e.Type), e.Meta())

	return nil
}

func (s *State) ApplyAccountSshKeyAdded(e *domain.AccountSshKeyAdded) error {
	for idx, wacc := range s.State.WrappedAccounts {
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
			s.State.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountSshKeyAdded", errRecordNotFound)
}

func (s *State) ApplyAccountUsernameChanged(e *domain.AccountUsernameChanged) error {
	for idx, wacc := range s.State.WrappedAccounts {
		if wacc.Account.Id == e.Id {
			wacc.Account.Username = e.Username
			s.State.WrappedAccounts[idx] = wacc
			return nil
		}
	}

	return ewrap("ApplyAccountUsernameChanged", errRecordNotFound)
}

func (s *State) ApplyAccountFolderCreated(e *domain.AccountFolderCreated) error {
	newFolder := apitypes.Folder{
		Id:       e.Id,
		ParentId: e.ParentId,
		Name:     e.Name,
	}

	s.State.Folders = append(s.State.Folders, newFolder)
	return nil
}

func (s *State) ApplyAccountFolderMoved(e *domain.AccountFolderMoved) error {
	for idx, acc := range s.State.Folders {
		if acc.Id == e.Id {
			acc.ParentId = e.ParentId
			s.State.Folders[idx] = acc
			return nil
		}
	}

	return ewrap("ApplyAccountFolderMoved", errRecordNotFound)
}

func (s *State) ApplyAccountFolderRenamed(e *domain.AccountFolderRenamed) error {
	for idx, acc := range s.State.Folders {
		if acc.Id == e.Id {
			acc.Name = e.Name
			s.State.Folders[idx] = acc
			return nil
		}
	}

	return ewrap("ApplyAccountFolderRenamed", errRecordNotFound)
}

func (s *State) ApplyDatabaseUnsealed(e *domain.DatabaseUnsealed) error {
	// no-op

	return nil
}

func (s *State) ApplyDatabaseMasterPasswordChanged(e *domain.DatabaseMasterPasswordChanged) error {
	s.SetMasterPassword(e.Password)

	return nil
}

func (s *State) ApplyDatabaseS3IntegrationConfigured(e *domain.DatabaseS3IntegrationConfigured) error {
	s.S3ExportBucket = e.Bucket
	s.S3ExportApiKey = e.ApiKey
	s.S3ExportSecret = e.Secret

	return nil
}

func (s *State) ApplyUserU2FTokenRegistered(e *domain.UserU2FTokenRegistered) error {
	s.State.U2FTokens[e.KeyHandle] = &U2FToken{
		KeyHandle:        e.KeyHandle,
		RegistrationData: e.RegistrationData,
		ClientData:       e.ClientData,
		Version:          e.Version,
		Counter:          0,
	}

	return nil
}

func (s *State) ApplyUserU2FTokenUsed(e *domain.UserU2FTokenUsed) error {
	token := s.State.U2FTokens[e.KeyHandle]

	token.Counter = uint32(e.Counter)

	return nil
}

func (s *State) HandleUnknownEvent(e domain.Event) error {
	return errors.New("unknown event: " + e.MetaType())
}
