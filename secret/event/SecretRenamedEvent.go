package event

import (
	"github.com/function61/pi-security-module/state"
)

type SecretRenamed struct {
	Id    string
	Title string
}

func (e *SecretRenamed) Apply() {
	for idx, s := range state.Data.Secrets {
		if s.Id == e.Id {
			s.Title = e.Title
			state.Data.Secrets[idx] = s
			return
		}
	}
}
