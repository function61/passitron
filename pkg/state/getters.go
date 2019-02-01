package state

import (
	"fmt"
	"github.com/function61/gokit/mac"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"time"
)

func (s *UserStorage) SubfoldersByParentId(id string) []apitypes.Folder {
	subFolders := []apitypes.Folder{}

	for _, f := range s.Folders {
		if f.ParentId != id {
			continue
		}

		subFolders = append(subFolders, f)
	}

	return subFolders
}

func (s *UserStorage) WrappedAccountsByFolder(id string) []WrappedAccount {
	accounts := []WrappedAccount{}

	for _, wacc := range s.WrappedAccounts {
		if wacc.Account.FolderId != id {
			continue
		}

		accounts = append(accounts, wacc)
	}

	return accounts
}

func (s *UserStorage) WrappedSecretById(accountId string, secretId string) *WrappedSecret {
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

func (s *UserStorage) WrappedAccountById(id string) *WrappedAccount {
	for _, wacc := range s.WrappedAccounts {
		if wacc.Account.Id == id {
			return &wacc
		}
	}

	return nil
}

func (s *UserStorage) FolderById(id string) *apitypes.Folder {
	for _, folder := range s.Folders {
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

func UnwrapSecrets(secrets []WrappedSecret, st *AppState) []apitypes.ExposedSecret {
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
			OtpProof:        otpProof,
			OtpProofTime:    otpProofTime,
			OtpKeyExportMac: mac.New(st.GetMacSigningKey(), psecret.Secret.Id).Sign(),
			Secret:          psecret.Secret,
		}

		ret = append(ret, es)
	}

	return ret
}

func (s *AppState) NextFreeUserId() string {
	// 1st user has ID of 2, so that's why we use + 2
	return fmt.Sprintf("%d", len(s.DB.UserScope)+2)
}
