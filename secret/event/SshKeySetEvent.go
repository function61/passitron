package event

import (
	"github.com/function61/pi-security-module/state"
)

type SshKeySet struct {
	Id                     string
	SshPrivateKey          string
	SshPublicKeyAuthorized string
}

func (e *SshKeySet) Apply() {
	for idx, s := range state.Inst.State.Secrets {
		if s.Id == e.Id {
			s.SshPrivateKey = e.SshPrivateKey
			s.SshPublicKeyAuthorized = e.SshPublicKeyAuthorized
			state.Inst.State.Secrets[idx] = s
			return
		}
	}
}
