package accountevent

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type AccountDeleted struct {
	eventbase.Event
	Id string
}

func (e AccountDeleted) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "AccountDeleted " + string(asJson)
}

func AccountDeletedFromSerialized(payload []byte) *AccountDeleted {
	var e AccountDeleted
	if err := json.Unmarshal(payload, &e); err != nil {
		panic(err)
	}
	return &e
}

func (e AccountDeleted) Apply() {
	for idx, s := range state.Inst.State.Accounts {
		if s.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			state.Inst.State.Accounts = append(
				state.Inst.State.Accounts[:idx],
				state.Inst.State.Accounts[idx+1:]...)
			return
		}
	}
}
