package eventapplicator

import (
	"github.com/function61/pi-security-module/util/eventbase"
	"log"
	"os"
)

var logHandle *os.File

func InitStreamLog(filename string) error {
	log.Printf("Opening stream log from %s", filename)

	var err error
	logHandle, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	return err
}

func CloseStreamLog() {
	log.Printf("Closing stream log")

	if err := logHandle.Close(); err != nil {
		log.Printf("Error closing stream log: %s", err.Error())
	}
}

func ApplyEvent(event eventbase.EventInterface) {
	if _, err := logHandle.Write([]byte(event.Serialize() + "\n")); err != nil {
		panic(err)
	}

	event.Apply()
}

func ApplyEvents(events []eventbase.EventInterface) {
	for _, event := range events {
		ApplyEvent(event)
	}
}
