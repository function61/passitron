package util

import (
	"github.com/function61/pi-security-module/util/eventbase"
	"log"
)

func ApplyEvent(event eventbase.EventInterface) {
	log.Printf("Applying %s", event.Serialize())

	event.Apply()
}

func ApplyEvents(events []eventbase.EventInterface) {
	for _, event := range events {
		ApplyEvent(event)
	}
}
