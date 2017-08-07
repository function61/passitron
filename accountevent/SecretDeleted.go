package accountevent

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type SecretDeleted struct {
	eventbase.Event
	Account string
	Secret  string
}

func (e *SecretDeleted) Apply() {
	for accountIdx, account := range state.Inst.State.Accounts {
		if account.Id == e.Account {
			for secretIdx, secret := range account.Secrets {
				if secret.Id == e.Secret {
					account.Secrets = append(
						account.Secrets[:secretIdx],
						account.Secrets[secretIdx+1:]...)
				}
			}
			state.Inst.State.Accounts[accountIdx] = account
			return
		}
	}
}
