package event

import (
	"github.com/function61/pi-security-module/util/eventbase"
	"log"
)

const (
	SecretUsedTypeSshSigning      = "SshSigning"
	SecretUsedTypePasswordExposed = "PasswordExposed"
)

type SecretUsed struct {
	eventbase.Event
	Secret string
	Type   string
}

func (e *SecretUsed) Apply() {
	log.Printf("Secret %s was used, type = %s", e.Secret, e.Type)
}
