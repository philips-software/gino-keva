package main

// Values represents a collection of values
type Values struct {
	values map[string]Value
}

// Add a key/value to the collection
func (v *Values) Add(key string, value Value) {
	v.values[key] = value
}

// Count returns number of items in collection
func (v Values) Count() int {
	return len(v.values)
}

// Get returns a single value from the collection
func (v Values) Get(key string) Value {
	return v.values[key]
}

// HasKey indicates if key exists in collection
func (v Values) HasKey(key string) bool {
	_, ok := v.values[key]
	return ok
}

// Iterate the collection value data
func (v Values) Iterate() map[string]Value {
	return v.values
}

// Remove a key from the collection
func (v *Values) Remove(key string) {
	delete(v.values, key)
}

// NewValues returns a new values map
func NewValues() *Values {
	return &Values{
		values: make(map[string]Value),
	}
}

// Value represents the parsed value as stored in git notes
type Value string
