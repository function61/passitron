package event

import (
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type SecretCreated struct {
	eventbase.Event
	Id       string
	FolderId string
	Title    string
}

func (e *SecretCreated) Apply() {
	secret := state.InsecureSecret{
		Id:       e.Id,
		FolderId: e.FolderId,
		Title:    e.Title,
		Created:  e.Timestamp,
	}

	state.Inst.State.Secrets = append(state.Inst.State.Secrets, secret)
}
