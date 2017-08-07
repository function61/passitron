package accountevent

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type OtpTokenAdded struct {
	eventbase.Event
	Account            string
	Id                 string
	OtpProvisioningUrl string
}

func (e *OtpTokenAdded) Apply() {
	for idx, account := range state.Inst.State.Accounts {
		if account.Id == e.Account {
			secret := state.Secret{
				Id:                 e.Id,
				Kind:               state.SecretKindOtpToken,
				Created:            e.Timestamp,
				OtpProvisioningUrl: e.OtpProvisioningUrl,
			}

			account.Secrets = append(account.Secrets, secret)
			state.Inst.State.Accounts[idx] = account
			return
		}
	}
}
