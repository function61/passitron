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

func (e DatabaseUnsealed) Apply() {
	// noop
}
