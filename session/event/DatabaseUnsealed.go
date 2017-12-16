package event

import (
	"encoding/json"
	"github.com/function61/pi-security-module/util/eventbase"
)

type DatabaseUnsealed struct {
	eventbase.Event
}

func (e DatabaseUnsealed) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "DatabaseUnsealed " + string(asJson)
}

func DatabaseUnsealedFromSerialized(payload []byte) *DatabaseUnsealed {
	var e DatabaseUnsealed
	if err := json.Unmarshal(payload, &e); err != nil {
		panic(err)
	}
	return &e
}

func (e DatabaseUnsealed) Apply() {
	// noop
}
