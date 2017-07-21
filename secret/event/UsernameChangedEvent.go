package event

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type UsernameChanged struct {
	eventbase.Event
	Id       string
	Username string
}

func (e *UsernameChanged) Apply() {
	for idx, s := range state.Inst.State.Secrets {
		if s.Id == e.Id {
			s.Username = e.Username
			state.Inst.State.Secrets[idx] = s
			return
		}
	}
}
