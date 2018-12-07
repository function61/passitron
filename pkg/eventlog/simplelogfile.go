package eventlog

import (
	"bufio"
	"fmt"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/event"
	"io"
	"log"
	"strings"
)

func readOldEvents(logReader io.Reader, eventAdded func(event.Event) error) (int, error) {
	logLineScanner := bufio.NewScanner(logReader)
	logLineScanner.Split(bufio.ScanLines)

	eventsRead := 0

	for logLineScanner.Scan() {
		event, err := domain.Deserialize(logLineScanner.Text())
		if err != nil {
			return eventsRead, fmt.Errorf("domain.Deserialize: %s", err.Error())
		}

		if err := eventAdded(event); err != nil {
			return eventsRead, err
		}

		eventsRead++
	}

	if logLineScanner.Err() != nil {
		return eventsRead, logLineScanner.Err()
	}

	return eventsRead, nil
}

type SimpleLogFile struct {
	logWriter  io.Writer
	eventAdded func(event.Event) error
}

func NewSimpleLogFile(logReader io.Reader, logWriter io.Writer, eventAdded func(event.Event) error) (Log, error) {
	eventsRead, err := readOldEvents(logReader, eventAdded)
	if err != nil {
		return nil, fmt.Errorf("readOldEvents: %s", err.Error())
	}

	log.Printf("readOldEvents: succesfully read %d event(s)", eventsRead)

	return &SimpleLogFile{
		logWriter:  logWriter,
		eventAdded: eventAdded,
	}, nil
}

func (e *SimpleLogFile) Append(events []event.Event) error {
	// the below algorithm does not work with 0 events
	if len(events) == 0 {
		return nil
	}

	serialized := []string{}

	for _, event := range events {
		serialized = append(serialized, event.Serialize())
	}

	writeBatch := strings.Join(serialized, "\n") + "\n"

	if _, err := e.logWriter.Write([]byte(writeBatch)); err != nil {
		return fmt.Errorf("Write(): %v", err)
	}

	// TODO: fsync()?

	for _, event := range events {
		if err := e.eventAdded(event); err != nil {
			return fmt.Errorf("eventAdded(): %v", err)
		}
	}

	return nil
}
