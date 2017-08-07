package accountevent

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type AccountCreated struct {
	eventbase.Event
	Id       string
	FolderId string
	Title    string
}

func (e *AccountCreated) Apply() {
	account := state.InsecureAccount{
		Id:       e.Id,
		FolderId: e.FolderId,
		Title:    e.Title,
		Created:  e.Timestamp,
	}

	state.Inst.State.Accounts = append(state.Inst.State.Accounts, account)
}
