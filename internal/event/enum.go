package event

import (
	"bytes"
	"encoding/json"
)

// EventType represents the type of an event as stored by Gino Keva
type EventType int

const (
	// Invalid represents an invalid eventtype
	Invalid EventType = iota
	// Set represents the value for a key to be set or overwritten if already set
	Set
	// Unset represents a key to be unset if present
	Unset
)

func (t EventType) String() string {
	return toString[t]
}

var toString = map[EventType]string{
	Set:   "set",
	Unset: "unset",
}

var toID = map[string]EventType{
	"set":   Set,
	"unset": Unset,
}

// MarshalJSON marshals the enum as a quoted json string
func (s EventType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *EventType) UnmarshalJSON(b []byte) error {
	var j string

	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	// Note that if the string cannot be found then it will be set to the zero value
	*s = toID[j]

	if *s == Invalid {
		return &UnknownType{EventType: j}
	}

	return nil
}
