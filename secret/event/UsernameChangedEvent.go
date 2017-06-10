package event

import (
	"github.com/function61/pi-security-module/state"
)

type UsernameChanged struct {
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
