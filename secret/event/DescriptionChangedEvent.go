package event

import (
	"github.com/function61/pi-security-module/state"
)

type DescriptionChanged struct {
	Id          string
	Description string
}

func (e *DescriptionChanged) Apply() {
	for idx, s := range state.Inst.State.Secrets {
		if s.Id == e.Id {
			s.Description = e.Description
			state.Inst.State.Secrets[idx] = s
			return
		}
	}
}
