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

func handleEvent(event domain.Event) error {
	switch e := event.(type) {
	case *domain.AccountCreated:
		handleAccountCreated(e, Inst)
	case *domain.AccountDeleted:
		handleAccountDeleted(e, Inst)
	case *domain.AccountRenamed:
		handleAccountRenamed(e, Inst)
	case *domain.AccountDescriptionChanged:
		handleAccountDescriptionChanged(e, Inst)
	case *domain.AccountOtpTokenAdded:
		handleAccountOtpTokenAdded(e, Inst)
	case *domain.AccountPasswordAdded:
		handleAccountPasswordAdded(e, Inst)
	case *domain.AccountSecretDeleted:
		handleAccountSecretDeleted(e, Inst)
	case *domain.AccountSecretUsed:
		handleAccountSecretUsed(e, Inst)
	case *domain.AccountSshKeyAdded:
		handleAccountSshKeyAdded(e, Inst)
	case *domain.AccountUsernameChanged:
		handleAccountUsernameChanged(e, Inst)
	case *domain.AccountFolderCreated:
		handleAccountFolderCreated(e, Inst)
	case *domain.AccountFolderMoved:
		handleAccountFolderMoved(e, Inst)
	case *domain.AccountFolderRenamed:
		handleAccountFolderRenamed(e, Inst)
	case *domain.DatabaseUnsealed:
		handleDatabaseUnsealed(e, Inst)
	case *domain.DatabaseMasterPasswordChanged:
		handleDatabaseMasterPasswordChanged(e, Inst)
	case *domain.DatabaseS3IntegrationConfigured:
		handleDatabaseS3IntegrationConfigured(e, Inst)
	default:
		panic(errors.New("unknown event: " + event.MetaType()))
	}

	return nil
}
