package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"
)

// table (keyed by MetaType()) of empty event struct allocators
type AllocatorMap map[string]func() Event

func (e *EventMeta) Serialize(payload Event) string {
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return payload.MetaType() + " " + e.Timestamp.Format(time.RFC3339Nano) + " " + e.UserId + " " + string(payloadJson)
}

// only the user ID is optional
var deserializeRe = regexp.MustCompile("^([^ ]+) ([^ ]+) ([^ ]*) (.+)$")

func Deserialize(serialized string, allocators AllocatorMap) (Event, error) {
	match := deserializeRe.FindStringSubmatch(serialized)
	if len(match) != 5 {
		return nil, errors.New("parsing failed")
	}

	eventType := match[1]
	eventDate, errDate := time.Parse(time.RFC3339Nano, match[2])
	if errDate != nil {
		return nil, errDate
	}

	eventUser := match[3]
	eventPayload := match[4]

	allocator, eventKnown := allocators[eventType]
	if !eventKnown {
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}

	event := allocator()

	meta := event.Meta()
	meta.Timestamp = eventDate
	meta.UserId = eventUser

	if err := json.Unmarshal([]byte(eventPayload), event); err != nil {
		return nil, err
	}

	return event, nil
}
