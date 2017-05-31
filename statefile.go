package main

import (
	"encoding/json"
	"io/ioutil"
)

const (
	statefilePath = "state.json"
)

func (s *Secret) GetPassword() *ExposedPassword {
	return &ExposedPassword{s.password}
}

type ExposedPassword struct {
	Password string
}

// "secure secret": contains all fields from InsecureSecret *except* password
// is unexported
type Secret struct {
	Id          string
	FolderId    string
	Title       string
	Username    string
	password    string
	Description string
	// created
	// password last changed
}

type InsecureSecret struct {
	Id          string
	FolderId    string
	Title       string
	Username    string
	Password    string
	Description string
	// created
	// password last changed
}

func (i *InsecureSecret) ToSecureSecret() Secret {
	return Secret{
		Id:          i.Id,
		FolderId:    i.FolderId,
		Title:       i.Title,
		Username:    i.Username,
		password:    i.Password,
		Description: i.Description,
	}
}

func (s *Secret) ToInsecureSecret() InsecureSecret {
	return InsecureSecret{
		Id:          s.Id,
		FolderId:    s.FolderId,
		Title:       s.Title,
		Username:    s.Username,
		Password:    s.password,
		Description: s.Description,
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
