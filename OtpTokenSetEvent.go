package main

type OtpTokenSet struct {
	Id                 string
	OtpProvisioningUrl string
}

func (e *OtpTokenSet) Apply() {
	for idx, s := range state.Secrets {
		if s.Id == e.Id {
			s.OtpProvisioningUrl = e.OtpProvisioningUrl
			state.Secrets[idx] = s
			return
		}
	}
}
