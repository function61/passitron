package state

import (
	"errors"
	"github.com/function61/pi-security-module/domain"
)

func handleAccountCreated(e *domain.AccountCreated, state *State) {
	account := InsecureAccount{
		Id:       e.Id,
		FolderId: e.FolderId,
		Title:    e.Title,
		Created:  e.Meta().Timestamp,
	}

	state.State.Accounts = append(state.State.Accounts, account)
}

func handleAccountDeleted(e *domain.AccountDeleted, state *State) {
	for idx, s := range state.State.Accounts {
		if s.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			state.State.Accounts = append(
				state.State.Accounts[:idx],
				state.State.Accounts[idx+1:]...)
			return
		}
	}
}

func handleAccountRenamed(e *domain.AccountRenamed, state *State) {
	for idx, s := range state.State.Accounts {
		if s.Id == e.Id {
			s.Title = e.Title
			state.State.Accounts[idx] = s
			return
		}
	}
}

func handleAccountDescriptionChanged(e *domain.AccountDescriptionChanged, state *State) {
	for idx, s := range state.State.Accounts {
		if s.Id == e.Id {
			s.Description = e.Description
			state.State.Accounts[idx] = s
			return
		}
	}
}

func handleAccountOtpTokenAdded(e *domain.AccountOtpTokenAdded, state *State) {
	for idx, account := range state.State.Accounts {
		if account.Id == e.Account {
			secret := Secret{
				Id:                 e.Id,
				Kind:               SecretKindOtpToken,
				Created:            e.Meta().Timestamp,
				OtpProvisioningUrl: e.OtpProvisioningUrl,
			}

			account.Secrets = append(account.Secrets, secret)
			state.State.Accounts[idx] = account
			return
		}
	}
}

func handleAccountPasswordAdded(e *domain.AccountPasswordAdded, state *State) {
	for idx, account := range state.State.Accounts {
		if account.Id == e.Account {
			secret := Secret{
				Id:       e.Id,
				Kind:     SecretKindPassword,
				Created:  e.Meta().Timestamp,
				Password: e.Password,
			}

			account.Secrets = append(account.Secrets, secret)
			state.State.Accounts[idx] = account
			return
		}
	}
}

func handleAccountSecretDeleted(e *domain.AccountSecretDeleted, state *State) {
	for accountIdx, account := range state.State.Accounts {
		if account.Id == e.Account {
			for secretIdx, secret := range account.Secrets {
				if secret.Id == e.Secret {
					account.Secrets = append(
						account.Secrets[:secretIdx],
						account.Secrets[secretIdx+1:]...)
				}
			}
			state.State.Accounts[accountIdx] = account
			return
		}
	}
}

func handleAccountSecretUsed(e *domain.AccountSecretUsed, state *State) {
	// no-op
}

func handleAccountSshKeyAdded(e *domain.AccountSshKeyAdded, state *State) {
	for idx, account := range state.State.Accounts {
		if account.Id == e.Account {
			secret := Secret{
				Id:                     e.Id,
				Kind:                   SecretKindSshKey,
				Created:                e.Meta().Timestamp,
				SshPrivateKey:          e.SshPrivateKey,
				SshPublicKeyAuthorized: e.SshPublicKeyAuthorized,
			}

			account.Secrets = append(account.Secrets, secret)
			state.State.Accounts[idx] = account
			return
		}
	}
}

func handleAccountUsernameChanged(e *domain.AccountUsernameChanged, state *State) {
	for idx, s := range state.State.Accounts {
		if s.Id == e.Id {
			s.Username = e.Username
			state.State.Accounts[idx] = s
			return
		}
	}
}

func handleAccountFolderCreated(e *domain.AccountFolderCreated, state *State) {
	newFolder := Folder{
		Id:       e.Id,
		ParentId: e.ParentId,
		Name:     e.Name,
	}

	state.State.Folders = append(state.State.Folders, newFolder)
}

func handleAccountFolderMoved(e *domain.AccountFolderMoved, state *State) {
	for idx, s := range state.State.Folders {
		if s.Id == e.Id {
			s.ParentId = e.ParentId
			state.State.Folders[idx] = s
			return
		}
	}
}

func handleAccountFolderRenamed(e *domain.AccountFolderRenamed, state *State) {
	for idx, s := range state.State.Folders {
		if s.Id == e.Id {
			s.Name = e.Name
			state.State.Folders[idx] = s
			return
		}
	}
}

func handleDatabaseUnsealed(e *domain.DatabaseUnsealed, state *State) {
	// no-op
}

func handleDatabaseMasterPasswordChanged(e *domain.DatabaseMasterPasswordChanged, state *State) {
	state.SetMasterPassword(e.Password)
}

func handleDatabaseS3IntegrationConfigured(e *domain.DatabaseS3IntegrationConfigured, state *State) {
	state.S3ExportBucket = e.Bucket
	state.S3ExportApiKey = e.ApiKey
	state.S3ExportSecret = e.Secret
}

func handleEvent(event domain.Event, state *State) error {
	switch e := event.(type) {
	case *domain.AccountCreated:
		handleAccountCreated(e, state)
	case *domain.AccountDeleted:
		handleAccountDeleted(e, state)
	case *domain.AccountRenamed:
		handleAccountRenamed(e, state)
	case *domain.AccountDescriptionChanged:
		handleAccountDescriptionChanged(e, state)
	case *domain.AccountOtpTokenAdded:
		handleAccountOtpTokenAdded(e, state)
	case *domain.AccountPasswordAdded:
		handleAccountPasswordAdded(e, state)
	case *domain.AccountSecretDeleted:
		handleAccountSecretDeleted(e, state)
	case *domain.AccountSecretUsed:
		handleAccountSecretUsed(e, state)
	case *domain.AccountSshKeyAdded:
		handleAccountSshKeyAdded(e, state)
	case *domain.AccountUsernameChanged:
		handleAccountUsernameChanged(e, state)
	case *domain.AccountFolderCreated:
		handleAccountFolderCreated(e, state)
	case *domain.AccountFolderMoved:
		handleAccountFolderMoved(e, state)
	case *domain.AccountFolderRenamed:
		handleAccountFolderRenamed(e, state)
	case *domain.DatabaseUnsealed:
		handleDatabaseUnsealed(e, state)
	case *domain.DatabaseMasterPasswordChanged:
		handleDatabaseMasterPasswordChanged(e, state)
	case *domain.DatabaseS3IntegrationConfigured:
		handleDatabaseS3IntegrationConfigured(e, state)
	default:
		panic(errors.New("unknown event: " + event.MetaType()))
	}

	return nil
}
