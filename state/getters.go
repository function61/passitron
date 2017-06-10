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

func SecretsByFolder(id string) []Secret {
	secrets := []Secret{}

	for _, s := range Inst.State.Secrets {
		if s.FolderId != id {
			continue
		}

		secrets = append(secrets, s.ToSecureSecret())
	}

	return secrets
}

func SecretById(id string) *Secret {
	for _, s := range Inst.State.Secrets {
		if s.Id == id {
			secret := s.ToSecureSecret()
			return &secret
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

func (s *Secret) GetPassword() *ExposedPassword {
	otpProof := ""

	if s.otpProvisioningUrl != "" {
		key, err := otp.NewKeyFromURL(s.otpProvisioningUrl)
		if err != nil {
			panic(err)
		}

		otpProof, err = totp.GenerateCode(key.Secret(), time.Now())
		if err != nil {
			panic(err)
		}
	}

	return &ExposedPassword{s.password, otpProof}
}

type ExposedPassword struct {
	Password string
	OtpProof string
}
