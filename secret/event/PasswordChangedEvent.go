package event

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type PasswordChanged struct {
	eventbase.Event
	Id       string
	Password string
}

func (e *PasswordChanged) Apply() {
	for idx, s := range state.Inst.State.Secrets {
		if s.Id == e.Id {
			s.Password = e.Password
			state.Inst.State.Secrets[idx] = s
			return
		}
	}
}
