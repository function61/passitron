package accountevent

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util/eventbase"
)

type UsernameChanged struct {
	eventbase.Event
	Id       string
	Username string
}

func (e UsernameChanged) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "UsernameChanged " + string(asJson)
}

func UsernameChangedFromSerialized(payload []byte) *UsernameChanged {
	var e UsernameChanged
	if err := json.Unmarshal(payload, &e); err != nil {
		panic(err)
	}
	return &e
}

func (e UsernameChanged) Apply() {
	for idx, s := range state.Inst.State.Accounts {
		if s.Id == e.Id {
			s.Username = e.Username
			state.Inst.State.Accounts[idx] = s
			return
		}
	}
}
