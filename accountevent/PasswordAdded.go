package accountevent

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type PasswordAdded struct {
	eventbase.Event
	Account  string
	Id       string
	Password string
}

func (e PasswordAdded) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "PasswordAdded " + string(asJson)
}

func PasswordAddedFromSerialized(payload []byte) *PasswordAdded {
	var e PasswordAdded
	if err := json.Unmarshal(payload, &e); err != nil {
		panic(err)
	}
	return &e
}

func (e PasswordAdded) Apply() {
	for idx, account := range state.Inst.State.Accounts {
		if account.Id == e.Account {
			secret := state.Secret{
				Id:       e.Id,
				Kind:     state.SecretKindPassword,
				Created:  e.Timestamp,
				Password: e.Password,
			}

			account.Secrets = append(account.Secrets, secret)
			state.Inst.State.Accounts[idx] = account
			return
		}
	}
}
