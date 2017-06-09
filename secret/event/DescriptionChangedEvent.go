package event

import (
	"github.com/function61/pi-security-module/state"
)

type DescriptionChanged struct {
	Id          string
	Description string
}

func (e *DescriptionChanged) Apply() {
	for idx, s := range state.Data.Secrets {
		if s.Id == e.Id {
			s.Description = e.Description
			state.Data.Secrets[idx] = s
			return
		}
	}
}
