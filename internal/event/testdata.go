package event

// TestData string constants
var (
	// TestDataFoo
	TestDataFoo = "foo"
	// TestDataBar
	TestDataBar = "bar"
	// TestDataKey
	TestDataKey = "key"
	// TestDataValue
	TestDataValue = "value"
	// TestDataOtherValue
	TestDataOtherValue = "otherValue"
)

// TestDataSetFooBar is an event which sets foo=bar
var TestDataSetFooBar = Event{
	EventType: Set,
	Key:       TestDataFoo,
	Value:     &TestDataBar,
}

// TestDataSetKeyValue is an event which sets key=value
var TestDataSetKeyValue = Event{
	EventType: Set,
	Key:       TestDataKey,
	Value:     &TestDataValue,
}

// TestDataSetKeyOtherValue is an event which sets key=otherValue
var TestDataSetKeyOtherValue = Event{
	EventType: Set,
	Key:       TestDataKey,
	Value:     &TestDataOtherValue,
}

// TestDataUnsetKey is an event which unsets key
var TestDataUnsetKey = Event{
	EventType: Unset,
	Key:       TestDataKey,
}
