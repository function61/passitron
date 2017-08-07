package accountevent

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type DescriptionChanged struct {
	eventbase.Event
	Id          string
	Description string
}

func (e *DescriptionChanged) Apply() {
	for idx, s := range state.Inst.State.Accounts {
		if s.Id == e.Id {
			s.Description = e.Description
			state.Inst.State.Accounts[idx] = s
			return
		}
	}
}
