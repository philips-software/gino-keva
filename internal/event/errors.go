package event

import "fmt"

// UnknownType error indicates Gino keva ran into an unknown event type stored in Git Notes
type UnknownType struct {
	EventType string
}

func (u UnknownType) Error() string {
	return fmt.Sprintf("Unknown event type: %s", u.EventType)
}

// KeyMissing error indicates Gino keva ran into an event with a missing or empty key
type KeyMissing struct {
	event Event
}

func (k KeyMissing) Error() string {
	return fmt.Sprintf("Key missing from event: %v", k.event)
}

// ValueMissing error indicates Gino keva ran into an event with a missing value
type ValueMissing struct {
	event Event
}

func (v ValueMissing) Error() string {
	return fmt.Sprintf("Value missing from event: %v", v.event)
}

// InvalidKey error indicates the key is not valid
type InvalidKey struct {
	msg string
}

func (i InvalidKey) Error() string {
	return fmt.Sprintf("Invalid key: %v", i.msg)
}
