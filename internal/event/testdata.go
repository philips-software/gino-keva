package event

var (
	// TestData string constants
	TestDataFoo        = "foo"
	TestDataBar        = "bar"
	TestDataKey        = "key"
	TestDataValue      = "value"
	TestDataOtherValue = "otherValue"
)

var TestDataSetFooBar = Event{
	EventType: Set,
	Key:       TestDataFoo,
	Value:     &TestDataBar,
}

var TestDataSetKeyValue = Event{
	EventType: Set,
	Key:       TestDataKey,
	Value:     &TestDataValue,
}

var TestDataSetKeyOtherValue = Event{
	EventType: Set,
	Key:       TestDataKey,
	Value:     &TestDataOtherValue,
}

var TestDataUnsetKey = Event{
	EventType: Unset,
	Key:       TestDataKey,
}
