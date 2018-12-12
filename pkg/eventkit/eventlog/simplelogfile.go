package eventlog

import (
	"bufio"
	"fmt"
	"github.com/function61/pi-security-module/pkg/eventkit/event"
	"io"
	"log"
	"strings"
)

type eventDeserializer func(serialized string) (event.Event, error)
type eventHandler func(ev event.Event) error

func readOldEvents(
	logReader io.Reader,
	handleEvent eventHandler,
	deserializeEvent eventDeserializer,
) (int, error) {
	logLineScanner := bufio.NewScanner(logReader)
	logLineScanner.Split(bufio.ScanLines)

	eventsRead := 0

	for logLineScanner.Scan() {
		event, err := deserializeEvent(logLineScanner.Text())
		if err != nil {
			return eventsRead, fmt.Errorf("deserializeEvent: %s", err.Error())
		}

		if err := handleEvent(event); err != nil {
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
	logWriter   io.Writer
	handleEvent eventHandler
}

func NewSimpleLogFile(
	logReader io.Reader,
	logWriter io.Writer,
	handleEvent eventHandler,
	deserializeEvent eventDeserializer,
) (Log, error) {
	eventsRead, err := readOldEvents(logReader, handleEvent, deserializeEvent)
	if err != nil {
		return nil, fmt.Errorf("readOldEvents: %s", err.Error())
	}

	log.Printf("readOldEvents: succesfully read %d event(s)", eventsRead)

	return &SimpleLogFile{
		logWriter:   logWriter,
		handleEvent: handleEvent,
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
		if err := e.handleEvent(event); err != nil {
			return fmt.Errorf("handleEvent(): %v", err)
		}
	}

	return nil
}
