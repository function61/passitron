package eventapplicator

import (
	"github.com/function61/pi-security-module/util/eventbase"
	"log"
	"os"
)

type EventApplicator struct {
	logHandle *os.File
}

func NewEventApplicator(filename string) *EventApplicator {
	log.Printf("Opening stream log from %s", filename)

	var err error
	logHandle, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("InitStreamLog: failure: %s", err.Error())
	}

	return &EventApplicator{
		logHandle: logHandle,
	}
}

func (e *EventApplicator) Close() {
	log.Printf("Closing stream log")

	if err := e.logHandle.Close(); err != nil {
		log.Printf("Error closing stream log: %s", err.Error())
	}
}

func (e *EventApplicator) Append(event eventbase.EventInterface) {
	if _, err := e.logHandle.Write([]byte(event.Serialize() + "\n")); err != nil {
		log.Fatalf("Append: failure: %s", err.Error())
	}

	event.Apply()
}

func (e *EventApplicator) AppendBatch(events []eventbase.EventInterface) {
	for _, event := range events {
		e.Append(event)
	}
}
