package event

import (
	"encoding/json"
	"github.com/function61/pi-security-module/util/eventbase"
)

type MasterPasswordChanged struct {
	eventbase.Event
}

func (e MasterPasswordChanged) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "MasterPasswordChanged " + string(asJson)
}

func MasterPasswordChangedFromSerialized(payload []byte) *MasterPasswordChanged {
	var e MasterPasswordChanged
	if err := json.Unmarshal(payload, &e); err != nil {
		panic(err)
	}
	return &e
}

func (e MasterPasswordChanged) Apply() {
	// noop
}
