package state

import (
	"errors"
	"fmt"
	"github.com/function61/pi-security-module/pkg/domain"
)

var (
	errRecordNotFound = errors.New("record not found")
)

func (s *State) ApplyAccountCreated(e *domain.AccountCreated) error {
	account := InsecureAccount{
		Id:       e.Id,
		FolderId: e.FolderId,
		Title:    e.Title,
		Created:  e.Meta().Timestamp,
	}

	s.State.Accounts = append(s.State.Accounts, account)

	return nil
}

func (s *State) ApplyAccountDeleted(e *domain.AccountDeleted) error {
	for idx, acc := range s.State.Accounts {
		if acc.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			s.State.Accounts = append(
				s.State.Accounts[:idx],
				s.State.Accounts[idx+1:]...)
			return nil
		}
	}

	return errRecordNotFound
}

func (s *State) ApplyAccountRenamed(e *domain.AccountRenamed) error {
	for idx, acc := range s.State.Accounts {
		if acc.Id == e.Id {
			acc.Title = e.Title
			s.State.Accounts[idx] = acc
			return nil
		}
	}

	return errRecordNotFound
}

func (s *State) ApplyAccountDescriptionChanged(e *domain.AccountDescriptionChanged) error {
	for idx, acc := range s.State.Accounts {
		if acc.Id == e.Id {
			acc.Description = e.Description
			s.State.Accounts[idx] = acc
			return nil
		}
	}

	return errRecordNotFound
}

func (s *State) ApplyAccountOtpTokenAdded(e *domain.AccountOtpTokenAdded) error {
	for idx, account := range s.State.Accounts {
		if account.Id == e.Account {
			secret := Secret{
				Id:                 e.Id,
				Kind:               SecretKindOtpToken,
				Created:            e.Meta().Timestamp,
				OtpProvisioningUrl: e.OtpProvisioningUrl,
			}

			account.Secrets = append(account.Secrets, secret)
			s.State.Accounts[idx] = account
			return nil
		}
	}

	return errRecordNotFound
}

func (s *State) ApplyAccountPasswordAdded(e *domain.AccountPasswordAdded) error {
	for idx, account := range s.State.Accounts {
		if account.Id == e.Account {
			secret := Secret{
				Id:       e.Id,
				Kind:     SecretKindPassword,
				Created:  e.Meta().Timestamp,
				Password: e.Password,
			}

			account.Secrets = append(account.Secrets, secret)
			s.State.Accounts[idx] = account
			return nil
		}
	}

	return errRecordNotFound
}

func (s *State) ApplyAccountKeylistAdded(e *domain.AccountKeylistAdded) error {
	for idx, account := range s.State.Accounts {
		if account.Id == e.Account {
			keyItems := []KeylistKey{}

			for _, key := range e.Keys {
				keyItems = append(keyItems, KeylistKey{
					Key:   key.Key,
					Value: key.Value,
				})
			}

			secret := Secret{
				Id:          e.Id,
				Kind:        SecretKindKeylist,
				Title:       e.Title,
				Created:     e.Meta().Timestamp,
				KeylistKeys: keyItems,
			}

			account.Secrets = append(account.Secrets, secret)
			s.State.Accounts[idx] = account
			return nil
		}
	}

	return errRecordNotFound
}

func (s *State) ApplyAccountSecretDeleted(e *domain.AccountSecretDeleted) error {
	for accountIdx, account := range s.State.Accounts {
		if account.Id == e.Account {
			for secretIdx, secret := range account.Secrets {
				if secret.Id == e.Secret {
					account.Secrets = append(
						account.Secrets[:secretIdx],
						account.Secrets[secretIdx+1:]...)
				}
			}
			s.State.Accounts[accountIdx] = account
			return nil
		}
	}

	return errRecordNotFound
}

func (s *State) ApplyAccountSecretUsed(e *domain.AccountSecretUsed) error {
	s.State.Audit(fmt.Sprintf("Secret %s used (%s)", e.Account, e.Type), e.Meta())

	return nil
}

func (s *State) ApplyAccountSshKeyAdded(e *domain.AccountSshKeyAdded) error {
	for idx, account := range s.State.Accounts {
		if account.Id == e.Account {
			secret := Secret{
				Id:                     e.Id,
				Kind:                   SecretKindSshKey,
				Created:                e.Meta().Timestamp,
				SshPrivateKey:          e.SshPrivateKey,
				SshPublicKeyAuthorized: e.SshPublicKeyAuthorized,
			}

			account.Secrets = append(account.Secrets, secret)
			s.State.Accounts[idx] = account
			return nil
		}
	}

	return errRecordNotFound
}

func (s *State) ApplyAccountUsernameChanged(e *domain.AccountUsernameChanged) error {
	for idx, acc := range s.State.Accounts {
		if acc.Id == e.Id {
			acc.Username = e.Username
			s.State.Accounts[idx] = acc
			return nil
		}
	}

	return errRecordNotFound
}

func (s *State) ApplyAccountFolderCreated(e *domain.AccountFolderCreated) error {
	newFolder := Folder{
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

	return errRecordNotFound
}

func (s *State) ApplyAccountFolderRenamed(e *domain.AccountFolderRenamed) error {
	for idx, acc := range s.State.Folders {
		if acc.Id == e.Id {
			acc.Name = e.Name
			s.State.Folders[idx] = acc
			return nil
		}
	}

	return errRecordNotFound
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

func (s *State) HandleUnknownEvent(e domain.Event) error {
	return errors.New("unknown event: " + e.MetaType())
}
