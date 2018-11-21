package eventlog

import (
	"bufio"
	"fmt"
	"github.com/function61/pi-security-module/pkg/domain"
	"io"
	"log"
)

func readOldEvents(logReader io.Reader, eventAdded func(domain.Event) error) error {
	logLineScanner := bufio.NewScanner(logReader)
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
	logWriter  io.Writer
	eventAdded func(domain.Event) error
}

func New(logReader io.Reader, logWriter io.Writer, eventAdded func(domain.Event) error) *EventLog {
	if err := readOldEvents(logReader, eventAdded); err != nil {
		log.Fatalf("readOldEvents: %s", err.Error())
	}

	return &EventLog{
		logWriter:  logWriter,
		eventAdded: eventAdded,
	}
}

func (e *EventLog) Append(event domain.Event) {
	if _, err := e.logWriter.Write([]byte(event.Serialize() + "\n")); err != nil {
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
