package event

import (
	"regexp"
)

// AddNewEvent to start of list
func AddNewEvent(events *[]Event, e *Event) []Event {
	return append([]Event{*e}, *events...)
}

// NewSetEvent will create a new event of type Set
func NewSetEvent(key string, value string) (*Event, error) {
	err := validateKey(key)
	if err != nil {
		return nil, err
	}

	return &Event{
		EventType: Set,
		Key:       key,
		Value:     &value,
	}, nil
}

// NewUnsetEvent will create a new event of type Unset
func NewUnsetEvent(key string) (*Event, error) {
	err := validateKey(key)
	if err != nil {
		return nil, err
	}

	return &Event{
		EventType: Unset,
		Key:       key,
	}, nil
}

func validateKey(key string) error {
	if key == "" {
		return &InvalidKey{msg: "key cannot be empty"}
	}

	{
		pattern := `[^A-Za-z0-9_-]`
		matched, err := regexp.Match(pattern, []byte(key))
		if err != nil {
			return err
		}

		if matched {
			return &InvalidKey{msg: "key contains invalid characters"}
		}
	}

	{
		pattern := `^[^A-Za-z]`
		matched, err := regexp.Match(pattern, []byte(key))
		if err != nil {
			return err
		}

		if matched {
			return &InvalidKey{msg: "first character is not a letter"}
		}
	}

	{
		pattern := `[^A-Za-z0-9]$`
		matched, err := regexp.Match(pattern, []byte(key))
		if err != nil {
			return err
		}

		if matched {
			return &InvalidKey{msg: "last character is not a letter or number"}
		}
	}

	return nil
}
