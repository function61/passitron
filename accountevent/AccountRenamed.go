package accountevent

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type AccountRenamed struct {
	eventbase.Event
	Id    string
	Title string
}

func (e *AccountRenamed) Apply() {
	for idx, s := range state.Inst.State.Accounts {
		if s.Id == e.Id {
			s.Title = e.Title
			state.Inst.State.Accounts[idx] = s
			return
		}
	}
}
