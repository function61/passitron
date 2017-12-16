package state

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/util/crypto"
	"github.com/function61/pi-security-module/util/eventapplicator"
	"io/ioutil"
	"log"
	"os"
)

const (
	statefilePath = "state.json"
	logfilePath   = "events.log"
)

type State struct {
	Password string
	State    *Statefile
	EventLog *eventapplicator.EventApplicator
}

func (s *Statefile) Save(password string) {
	jsonBytes, errJson := json.MarshalIndent(s, "", "    ")
	if errJson != nil {
		panic(errJson)
	}

	encryptedBytes, err := crypto.Encrypt(jsonBytes, password)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(statefilePath, encryptedBytes, 0644); err != nil {
		panic(err)
	}
}

func writeBlankStatefile(password string) {
	rootFolder := Folder{
		Id:       "root",
		ParentId: "",
		Name:     "root",
	}

	Inst.State = &Statefile{
		Accounts: []InsecureAccount{},
		Folders:  []Folder{rootFolder},
	}

	Inst.State.Save(password)
}

var Inst *State

func Initialize() {
	if Inst != nil {
		panic(errors.New("statefile: initialize called twice"))
	}

	ea := eventapplicator.NewEventApplicator(logfilePath)

	Inst = &State{
		Password: "",
		State:    nil,
		EventLog: ea,
	}
}

func (s *State) Close() {
	s.EventLog.Close()

	Inst = nil
}

func (s *State) Unseal(password string) error {
	state, err := ReadStatefile(password)
	if err != nil {
		return err
	}

	s.Password = password
	s.State = state

	return nil
}

func (s *State) ChangePassword(newPassword string) error {
	s.Password = newPassword

	return s.Save()
}

// for Keepass export
func (s *State) GetMasterPassword() string {
	return s.Password
}

func (s *State) Save() error {
	s.State.Save(s.Password)

	return nil
}

func (s *State) IsUnsealed() bool {
	return s.State != nil
}

func ReadStatefile(password string) (*Statefile, error) {
	var s Statefile

	if _, err := os.Stat(statefilePath); os.IsNotExist(err) {
		log.Printf("Statefile does not exist. Initializing %s", statefilePath)

		writeBlankStatefile(password)
	}

	encryptedBytes, err := ioutil.ReadFile(statefilePath)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := crypto.Decrypt(encryptedBytes, password)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonBytes, &s); err != nil {
		return nil, err
	}

	return &s, nil
}
