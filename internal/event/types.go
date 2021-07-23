package event

// Event represents an event stored in git notes
type Event struct {
	EventType EventType `json:"type"`
	Key       string    `json:"key"`
	Value     *string   `json:"value,omitempty"`
}
