package event

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type OtpTokenSet struct {
	eventbase.Event
	Id                 string
	OtpProvisioningUrl string
}

func (e *OtpTokenSet) Apply() {
	for idx, s := range state.Inst.State.Secrets {
		if s.Id == e.Id {
			s.OtpProvisioningUrl = e.OtpProvisioningUrl
			state.Inst.State.Secrets[idx] = s
			return
		}
	}
}
