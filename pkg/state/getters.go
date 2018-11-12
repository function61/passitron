package state

import (
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"time"
)

func (s *State) SubfoldersByParentId(id string) []apitypes.Folder {
	subFolders := []apitypes.Folder{}

	for _, f := range s.State.Folders {
		if f.ParentId != id {
			continue
		}

		subFolders = append(subFolders, f)
	}

	return subFolders
}

func (s *State) WrappedAccountsByFolder(id string) []WrappedAccount {
	accounts := []WrappedAccount{}

	for _, wacc := range s.State.WrappedAccounts {
		if wacc.Account.FolderId != id {
			continue
		}

		accounts = append(accounts, wacc)
	}

	return accounts
}

func (s *State) WrappedSecretById(accountId string, secretId string) *WrappedSecret {
	wacc := s.WrappedAccountById(accountId)
	if wacc == nil {
		return nil
	}

	for _, wrappedSecret := range wacc.Secrets {
		if wrappedSecret.Secret.Id == secretId {
			return &wrappedSecret
		}
	}

	return nil
}

func (s *State) WrappedAccountById(id string) *WrappedAccount {
	for _, wacc := range s.State.WrappedAccounts {
		if wacc.Account.Id == id {
			return &wacc
		}
	}

	return nil
}

func (s *State) FolderById(id string) *apitypes.Folder {
	for _, folder := range s.State.Folders {
		if folder.Id == id {
			return &folder
		}
	}

	return nil
}

func UnwrapAccounts(waccs []WrappedAccount) []apitypes.Account {
	ret := []apitypes.Account{}

	for _, wacc := range waccs {
		ret = append(ret, wacc.Account)
	}

	return ret
}

func UnwrapSecrets(secrets []WrappedSecret) []apitypes.ExposedSecret {
	ret := []apitypes.ExposedSecret{}

	otpProofTime := time.Now()

	for _, psecret := range secrets {
		otpProof := ""

		if psecret.OtpProvisioningUrl != "" {
			key, err := otp.NewKeyFromURL(psecret.OtpProvisioningUrl)
			if err != nil {
				panic(err)
			}

			otpProof, err = totp.GenerateCode(key.Secret(), otpProofTime)
			if err != nil {
				panic(err)
			}
		}

		es := apitypes.ExposedSecret{
			OtpProof:     otpProof,
			OtpProofTime: otpProofTime,
			Secret:       psecret.Secret,
		}

		ret = append(ret, es)
	}

	return ret
}
