package accountevent

import (
	"encoding/json"
	"github.com/function61/pi-security-module/util/eventbase"
)

const (
	SecretUsedTypeSshSigning      = "SshSigning"
	SecretUsedTypePasswordExposed = "PasswordExposed"
)

type SecretUsed struct {
	eventbase.Event
	Account string
	Type    string
}

func (e SecretUsed) Serialize() string {
	asJson, _ := json.Marshal(e)

	return "SecretUsed " + string(asJson)
}

func (e SecretUsed) Apply() {
	// noop
}
