package accountevent

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type AccountCreated struct {
	eventbase.Event
	Id       string
	FolderId string
	Title    string
}

func (e AccountCreated) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "AccountCreated " + string(asJson)
}

func (e AccountCreated) Apply() {
	account := state.InsecureAccount{
		Id:       e.Id,
		FolderId: e.FolderId,
		Title:    e.Title,
		Created:  e.Timestamp,
	}

	state.Inst.State.Accounts = append(state.Inst.State.Accounts, account)
}
