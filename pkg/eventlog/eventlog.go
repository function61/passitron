package eventlog

import (
	"bufio"
	"fmt"
	"github.com/function61/pi-security-module/pkg/domain"
	"log"
	"os"
)

func readOldEvents(filename string, eventAdded func(domain.Event) error) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatal("events.log not present. Either create it or resolve the problem.")
	}

	logFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer logFile.Close()

	logLineScanner := bufio.NewScanner(logFile)
	logLineScanner.Split(bufio.ScanLines)

	eventsRead := 0

	for logLineScanner.Scan() {
		event, err := domain.Deserialize(logLineScanner.Text())
		if err != nil {
			return fmt.Errorf("Failed to parse event: %s", err.Error())
		}

		if err := eventAdded(event); err != nil {
			return err
		}

		eventsRead++
	}

	if logLineScanner.Err() != nil {
		log.Fatalf("logLineScanner error: %s", logLineScanner.Err().Error())
	}

	log.Printf("ReadOldEvents(): read %d event(s)", eventsRead)

	return nil
}

type EventLog struct {
	logHandle  *os.File
	eventAdded func(domain.Event) error
}

func New(filename string, eventAdded func(domain.Event) error) *EventLog {
	log.Printf("Opening stream log from %s", filename)

	if err := readOldEvents(filename, eventAdded); err != nil {
		panic(err)
	}

	logHandle, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("InitStreamLog: failure: %s", err.Error())
	}

	return &EventLog{
		logHandle:  logHandle,
		eventAdded: eventAdded,
	}
}

func (e *EventLog) Close() {
	log.Printf("Closing stream log")

	if err := e.logHandle.Close(); err != nil {
		log.Printf("Error closing stream log: %s", err.Error())
	}
}
func (e *EventLog) Append(event domain.Event) {
	if _, err := e.logHandle.Write([]byte(event.Serialize() + "\n")); err != nil {
		log.Fatalf("Append: failure: %s", err.Error())
	}

	if err := e.eventAdded(event); err != nil {
		panic(err)
	}
}

func (e *EventLog) AppendBatch(events []domain.Event) {
	for _, event := range events {
		e.Append(event)
	}
}
