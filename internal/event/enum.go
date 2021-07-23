package event

import (
	"bytes"
	"encoding/json"
)

// Type represents the type of an event as stored by Gino Keva
type Type int

const (
	// Invalid represents an invalid eventtype
	Invalid Type = iota
	// Set represents the value for a key to be set or overwritten if already set
	Set
	// Unset represents a key to be unset if present
	Unset
)

func (t Type) String() string {
	return toString[t]
}

var toString = map[Type]string{
	Set:   "set",
	Unset: "unset",
}

var toID = map[string]Type{
	"set":   Set,
	"unset": Unset,
}

// MarshalJSON marshals the enum as a quoted json string
func (t Type) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[t])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (t *Type) UnmarshalJSON(b []byte) error {
	var j string

	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	// Note that if the string cannot be found then it will be set to the zero value
	*t = toID[j]

	if *t == Invalid {
		return &UnknownType{EventType: j}
	}

	return nil
}
