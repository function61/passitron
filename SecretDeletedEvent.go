package main

import (
	"github.com/function61/pi-security-module/state"
)

type SecretDeleted struct {
	Id string
}

func (e *SecretDeleted) Apply() {
	for idx, s := range state.Data.Secrets {
		if s.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			state.Data.Secrets = append(state.Data.Secrets[:idx], state.Data.Secrets[idx+1:]...)
			return
		}
	}
}
