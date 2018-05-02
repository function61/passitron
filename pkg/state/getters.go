package state

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"time"
)

func (s *State) SubfoldersById(id string) []Folder {
	subFolders := []Folder{}

	for _, f := range s.State.Folders {
		if f.ParentId != id {
			continue
		}

		subFolders = append(subFolders, f)
	}

	return subFolders
}

func (s *State) AccountsByFolder(id string) []SecureAccount {
	accounts := []SecureAccount{}

	for _, s := range s.State.Accounts {
		if s.FolderId != id {
			continue
		}

		accounts = append(accounts, s.ToSecureAccount())
	}

	return accounts
}

func (s *State) AccountById(id string) *SecureAccount {
	for _, s := range s.State.Accounts {
		if s.Id == id {
			account := s.ToSecureAccount()
			return &account
		}
	}

	return nil
}

func (s *State) FolderById(id string) *Folder {
	for _, f := range s.State.Folders {
		if f.Id == id {
			return &f
		}
	}

	return nil
}

func (s *State) FolderByName(name string) *Folder {
	for _, f := range s.State.Folders {
		if f.Name == name {
			return &f
		}
	}

	return nil
}

func (s *SecureAccount) GetSecrets() []ExposedSecret {
	secrets := []ExposedSecret{}

	otpProofTime := time.Now()

	for _, secret := range s.secrets {
		otpProof := ""

		if secret.OtpProvisioningUrl != "" {
			key, err := otp.NewKeyFromURL(secret.OtpProvisioningUrl)
			if err != nil {
				panic(err)
			}

			otpProof, err = totp.GenerateCode(key.Secret(), otpProofTime)
			if err != nil {
				panic(err)
			}
		}

		secrets = append(secrets, ExposedSecret{
			Id:                     secret.Id,
			Kind:                   secret.Kind,
			Title:                  secret.Title,
			Created:                secret.Created,
			Password:               secret.Password,
			OtpProof:               otpProof,
			OtpProofTime:           otpProofTime,
			KeylistKeys:            secret.KeylistKeys,
			SshPublicKeyAuthorized: secret.SshPublicKeyAuthorized,
		})
	}

	return secrets
}

type ExposedSecret struct {
	Id                     string
	Kind                   string
	Title                  string
	Created                time.Time
	Password               string
	OtpProof               string
	OtpProofTime           time.Time
	KeylistKeys            []KeylistKey
	SshPublicKeyAuthorized string
}
