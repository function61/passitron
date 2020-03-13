package state

import (
	"github.com/function61/gokit/mac"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"strings"
	"time"
)

func (s *UserStorage) SensitiveUser() SensitiveUser {
	s.mu.Lock()
	defer s.mu.Unlock()

	return *s.sUser
}

func (s *UserStorage) SubfoldersByParentId(id string) []apitypes.Folder {
	s.mu.Lock()
	defer s.mu.Unlock()

	subFolders := []apitypes.Folder{}

	for _, f := range s.folders {
		if f.ParentId != id {
			continue
		}

		subFolders = append(subFolders, *f)
	}

	return subFolders
}

func (s *UserStorage) WrappedAccountsByFolder(id string) []WrappedAccount {
	s.mu.Lock()
	defer s.mu.Unlock()

	accounts := []WrappedAccount{}

	for _, acc := range s.accounts {
		if acc.WrappedAccount.Account.FolderId != id {
			continue
		}

		accounts = append(accounts, *acc.WrappedAccount)
	}

	return accounts
}

func (s *UserStorage) U2FTokens() []*U2FToken {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokens := []*U2FToken{}

	for _, token := range s.u2FTokens {
		tokens = append(tokens, token)
	}

	return tokens
}

func (s *UserStorage) WrappedAccounts() []WrappedAccount {
	s.mu.Lock()
	defer s.mu.Unlock()

	waccs := []WrappedAccount{}

	for _, acc := range s.accounts {
		waccs = append(waccs, *acc.WrappedAccount)
	}

	return waccs
}

func (s *UserStorage) SearchAccounts(query string) []apitypes.Account {
	queryLowercased := strings.ToLower(query)

	matches := []apitypes.Account{}

	for _, acc := range s.accounts {
		if !strings.Contains(strings.ToLower(acc.WrappedAccount.Account.Title), queryLowercased) {
			continue
		}

		matches = append(matches, acc.WrappedAccount.Account)
	}

	return matches
}

func (s *UserStorage) WrappedSecretById(accountId string, secretId string) *WrappedSecret {
	// WrappedAccountById() does locking
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
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, acc := range s.accounts {
		if acc.WrappedAccount.Account.Id == id {
			return acc.WrappedAccount
		}
	}

	return nil
}

func (s *UserStorage) FolderById(id string) *apitypes.Folder {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, folder := range s.folders {
		if folder.Id == id {
			return folder
		}
	}

	return nil
}

func (s *UserStorage) SearchFolders(query string) []apitypes.Folder {
	s.mu.Lock()
	defer s.mu.Unlock()

	queryLowercased := strings.ToLower(query)

	matches := []apitypes.Folder{}

	for _, folder := range s.folders {
		if !strings.Contains(strings.ToLower(folder.Name), queryLowercased) {
			continue
		}

		matches = append(matches, *folder)
	}

	return matches
}

func UnwrapAccounts(waccs []WrappedAccount) []apitypes.Account {
	ret := []apitypes.Account{}

	for _, wacc := range waccs {
		ret = append(ret, wacc.Account)
	}

	return ret
}

func UnwrapSecrets(secrets []WrappedSecret, st *AppState) ([]apitypes.ExposedSecret, error) {
	exposed := []apitypes.ExposedSecret{}

	otpProofTime := time.Now()

	for _, psecret := range secrets {
		otpProof := ""

		if psecret.OtpProvisioningUrl != "" {
			key, err := otp.NewKeyFromURL(psecret.OtpProvisioningUrl)
			if err != nil {
				return nil, err
			}

			otpProof, err = totp.GenerateCode(key.Secret(), otpProofTime)
			if err != nil {
				return nil, err
			}
		}

		exposed = append(exposed, apitypes.ExposedSecret{
			OtpProof:        otpProof,
			OtpProofTime:    otpProofTime,
			OtpKeyExportMac: mac.New(st.GetMacSigningKey(), psecret.Secret.Id).Sign(),
			Secret:          psecret.Secret,
		})
	}

	return exposed, nil
}

func (s *UserStorage) AuditLog() []apitypes.AuditlogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.auditLog
}

func (s *UserStorage) S3ExportDetails() *S3ExportDetails {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.s3ExportDetails
}
