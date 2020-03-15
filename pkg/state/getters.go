package state

import (
	"encoding/json"
	"errors"
	"github.com/function61/gokit/mac"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"strings"
	"time"
)

func (l *UserStorage) MacKey() []byte {
	if len(l.macKey) == 0 {
		panic("empty macKey")
	}

	return l.macKey
}

func (s *UserStorage) Crypto() *cryptoThingie {
	return s.crypto
}

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

func (s *UserStorage) WrappedAccountsByFolder(id string) []InternalAccount {
	s.mu.Lock()
	defer s.mu.Unlock()

	accounts := []InternalAccount{}

	for _, acc := range s.accounts {
		if acc.Account.FolderId != id {
			continue
		}

		accounts = append(accounts, *acc)
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

func (s *UserStorage) WrappedAccounts() []InternalAccount {
	s.mu.Lock()
	defer s.mu.Unlock()

	accounts := []InternalAccount{}

	for _, acc := range s.accounts {
		accounts = append(accounts, *acc)
	}

	return accounts
}

func (s *UserStorage) SearchAccounts(query string) []apitypes.Account {
	queryLowercased := strings.ToLower(query)

	matches := []apitypes.Account{}

	for _, acc := range s.accounts {
		if !strings.Contains(strings.ToLower(acc.Account.Title), queryLowercased) {
			continue
		}

		matches = append(matches, acc.Account)
	}

	return matches
}

func (s *UserStorage) InternalSecretById(accountId string, secretId string) *InternalSecret {
	// WrappedAccountById() does locking
	acc := s.WrappedAccountById(accountId)
	if acc == nil {
		return nil
	}

	for _, secret := range acc.Secrets {
		if secret.Id == secretId {
			return &secret
		}
	}

	return nil
}

func (s *UserStorage) WrappedAccountById(id string) *InternalAccount {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, acc := range s.accounts {
		if acc.Account.Id == id {
			return acc
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

func (s *UserStorage) DecryptOtpProvisioningUrl(secret InternalSecret) (string, error) {
	// could be dangerous to expose other secret material as "otp provisioning URL"
	if secret.Kind != domain.SecretKindOtpToken {
		return "", errors.New("DecryptOtpProvisioningUrl with invalid kind")
	}

	otpProvisioningUrl, err := s.crypto.Decrypt(secret.Envelope)
	if err != nil {
		return "", err
	}

	return string(otpProvisioningUrl), nil
}

func (s *UserStorage) DecryptKeylist(secret InternalSecret) ([]domain.AccountKeylistAddedKeysItem, error) {
	keylistJson, err := s.crypto.Decrypt(secret.Envelope)
	if err != nil {
		return nil, err
	}

	keys := []domain.AccountKeylistAddedKeysItem{}
	if err := json.Unmarshal(keylistJson, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}

func (s *UserStorage) DecryptSecrets(
	secrets []InternalSecret,
) ([]apitypes.ExposedSecret, error) {
	exposed := []apitypes.ExposedSecret{}

	otpProofTime := time.Now()

	for _, internalSecret := range secrets {
		otpProof := ""
		otpKeyExportMac := ""
		note := []byte{}
		password := []byte{}

		var err error

		switch domain.SecretKindExhaustive97ac5d(internalSecret.Kind) {
		case domain.SecretKindNote:
			note, err = s.crypto.Decrypt(internalSecret.Envelope)
			if err != nil {
				return nil, err
			}
		case domain.SecretKindPassword:
			password, err = s.crypto.Decrypt(internalSecret.Envelope)
			if err != nil {
				return nil, err
			}
		case domain.SecretKindOtpToken:
			otpProvisioningUrl, err := s.DecryptOtpProvisioningUrl(internalSecret)
			if err != nil {
				return nil, err
			}

			key, err := otp.NewKeyFromURL(otpProvisioningUrl)
			if err != nil {
				return nil, err
			}

			otpProof, err = totp.GenerateCode(key.Secret(), otpProofTime)
			if err != nil {
				return nil, err
			}

			otpKeyExportMac = s.OtpKeyExportMac(&internalSecret).Sign()
		case domain.SecretKindSshKey:
			// special handling elsewhere, never exposed to UI
		case domain.SecretKindKeylist:
			// special handling elsewhere
		case domain.SecretKindExternalToken:
			// informational - there's no secret
		}

		exposed = append(exposed, apitypes.ExposedSecret{
			OtpProof:        otpProof,
			OtpProofTime:    otpProofTime,
			OtpKeyExportMac: otpKeyExportMac,
			Secret: apitypes.Secret{
				Id:                     internalSecret.Id,
				Kind:                   internalSecret.Kind,
				Created:                internalSecret.created,
				Title:                  internalSecret.Title,
				ExternalTokenKind:      internalSecret.externalTokenKind,
				KeylistKeyExample:      internalSecret.keylistKeyExample,
				SshPublicKeyAuthorized: internalSecret.SshPublicKeyAuthorized,
				Note:                   string(note),
				Password:               string(password),
			},
		})
	}

	return exposed, nil
}

func (s *UserStorage) OtpKeyExportMac(secret *InternalSecret) *mac.Mac {
	return mac.New(string(s.MacKey()), secret.Id)
}

func UnwrapAccounts(iaccounts []InternalAccount) []apitypes.Account {
	accounts := []apitypes.Account{}

	for _, acccount := range iaccounts {
		accounts = append(accounts, acccount.Account)
	}

	return accounts
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
