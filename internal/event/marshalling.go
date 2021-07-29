package event

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Marshal a list of Event objects into a string
func Marshal(events *[]Event) (string, error) {
	if events == nil {
		events = &[]Event{}
	}

	wrappedEvents := map[string]*[]Event{"events": events}
	result, err := json.Marshal(wrappedEvents)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n", result), nil
}

// Unmarshal a string into a list of Event objects
func Unmarshal(s string, events *[]Event) error {
	r := make(map[string]json.RawMessage)
	if err := json.Unmarshal([]byte(s), &r); err != nil {
		return err
	}

	if eventsJson, ok := r["events"]; ok {
		if err := json.Unmarshal(eventsJson, &events); err != nil {
			return err
		}
	} else {
		log.Debug("Ignoring since cannot find events key in JSON. Old syntax?")
		*events = []Event{}
		return nil
	}

	for _, e := range *events {
		switch e.EventType {
		case Set:
			if e.Value == nil {
				return &ValueMissing{e}
			}
			fallthrough
		case Unset:
			if e.Key == "" {
				return &KeyMissing{e}
			}
		default:
			log.Fatal("Fatal: Unknown event type encountered")
		}
	}
	return nil
}
