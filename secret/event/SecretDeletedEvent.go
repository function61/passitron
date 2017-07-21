package event

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type SecretDeleted struct {
	eventbase.Event
	Id string
}

func (e *SecretDeleted) Apply() {
	for idx, s := range state.Inst.State.Secrets {
		if s.Id == e.Id {
			// https://github.com/golang/go/wiki/SliceTricks
			state.Inst.State.Secrets = append(
				state.Inst.State.Secrets[:idx],
				state.Inst.State.Secrets[idx+1:]...)
			return
		}
	}
}
