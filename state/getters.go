package state

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"time"
)

func SubfoldersById(id string) []Folder {
	subFolders := []Folder{}

	for _, f := range Inst.State.Folders {
		if f.ParentId != id {
			continue
		}

		subFolders = append(subFolders, f)
	}

	return subFolders
}

func AccountsByFolder(id string) []SecureAccount {
	accounts := []SecureAccount{}

	for _, s := range Inst.State.Accounts {
		if s.FolderId != id {
			continue
		}

		accounts = append(accounts, s.ToSecureAccount())
	}

	return accounts
}

func AccountById(id string) *SecureAccount {
	for _, s := range Inst.State.Accounts {
		if s.Id == id {
			account := s.ToSecureAccount()
			return &account
		}
	}

	return nil
}

func FolderById(id string) *Folder {
	for _, f := range Inst.State.Folders {
		if f.Id == id {
			return &f
		}
	}

	return nil
}

func FolderByName(name string) *Folder {
	for _, f := range Inst.State.Folders {
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
			Created:                secret.Created,
			Password:               secret.Password,
			OtpProof:               otpProof,
			OtpProofTime:           otpProofTime,
			SshPublicKeyAuthorized: secret.SshPublicKeyAuthorized,
		})
	}

	return secrets
}

type ExposedSecret struct {
	Id                     string
	Kind                   string
	Created                time.Time
	Password               string
	OtpProof               string
	OtpProofTime           time.Time
	SshPublicKeyAuthorized string
}
