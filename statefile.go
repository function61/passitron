package main

import (
	"encoding/json"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"io/ioutil"
	"time"
)

const (
	statefilePath = "state.json"
)

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

// "secure secret": contains all fields from InsecureSecret *except* password
// is unexported
type Secret struct {
	Id                 string
	FolderId           string
	Title              string
	Username           string
	password           string
	otpProvisioningUrl string
	Description        string
	// created
	// password last changed
}

type InsecureSecret struct {
	Id                 string
	FolderId           string
	Title              string
	Username           string
	Password           string
	OtpProvisioningUrl string
	Description        string
	// created
	// password last changed
}

func (i *InsecureSecret) ToSecureSecret() Secret {
	return Secret{
		Id:                 i.Id,
		FolderId:           i.FolderId,
		Title:              i.Title,
		Username:           i.Username,
		password:           i.Password,
		otpProvisioningUrl: i.OtpProvisioningUrl,
		Description:        i.Description,
	}
}

func (s *Secret) ToInsecureSecret() InsecureSecret {
	return InsecureSecret{
		Id:                 s.Id,
		FolderId:           s.FolderId,
		Title:              s.Title,
		Username:           s.Username,
		Password:           s.password,
		OtpProvisioningUrl: s.otpProvisioningUrl,
		Description:        s.Description,
	}
}

type Folder struct {
	Id       string
	ParentId string
	Name     string
}

type Statefile struct {
	Secrets []InsecureSecret
	Folders []Folder
}

func (s *Statefile) Save() {
	jsonBytes, errJson := json.MarshalIndent(s, "", "    ")
	if errJson != nil {
		panic(errJson)
	}

	err := ioutil.WriteFile(statefilePath, jsonBytes, 0644)

	if err != nil {
		panic(err)
	}
}

func ReadStatefile() (*Statefile, error) {
	var s Statefile

	jsonBytes, err := ioutil.ReadFile(statefilePath)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(jsonBytes, &s); err != nil {
		panic(err)
	}

	return &s, nil
}
