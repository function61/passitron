package accountevent

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type AccountDeleted struct {
	eventbase.Event
	Id string
}

func (e *AccountDeleted) Apply() {
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
