package event

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// Correct events
	rawEventSetFooBar   = "{\"type\":\"set\",\"key\":\"foo\",\"value\":\"bar\"}"
	rawEventSetKeyValue = "{\"type\":\"set\",\"key\":\"key\",\"value\":\"value\"}"
	rawEventUnsetKey    = "{\"type\":\"unset\",\"key\":\"key\"}"

	// Incorrect events
	rawEventTypeUnknown           = "{\"type\":\"unknown\"}"
	rawEventSetFooMissingValue    = "{\"type\":\"set\",\"key\":\"foo\"}"
	rawEventSetMissingKeyValueBar = "{\"type\":\"set\",\"value\":\"bar\"}"
	rawEventUnsetMissingKey       = "{\"type\":\"unset\"}"
)

func wrapEvents(events ...string) string {
	return fmt.Sprintf("{\"events\":[%v]}\n", strings.Join(events, ","))
}

func TestMarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  *[]Event
		wanted string
	}{
		{
			name:   "nil input",
			input:  nil,
			wanted: wrapEvents(),
		},
		{
			name:   "no events",
			input:  &[]Event{},
			wanted: wrapEvents(),
		},
		{
			name:   "set foo=bar",
			input:  &[]Event{TestDataSetFooBar},
			wanted: wrapEvents(rawEventSetFooBar),
		},
		{
			name:   "set foo=bar, set key=value",
			input:  &[]Event{TestDataSetFooBar, TestDataSetKeyValue},
			wanted: wrapEvents(rawEventSetFooBar, rawEventSetKeyValue),
		},
		{
			name:   "set key=value, unset key",
			input:  &[]Event{TestDataSetKeyValue, TestDataUnsetKey},
			wanted: wrapEvents(rawEventSetKeyValue, rawEventUnsetKey),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Marshal(tc.input)

			assert.NoError(t, err)
			assert.Equal(t, tc.wanted, got)
		})
	}
}

func TestUnMarshalInvalidFormat(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		wantedErrorType error
	}{
		{
			name:            "unknown event",
			input:           wrapEvents(rawEventTypeUnknown),
			wantedErrorType: &UnknownType{},
		},
		{
			name:            "Set with missing value",
			input:           wrapEvents(rawEventSetFooMissingValue),
			wantedErrorType: &ValueMissing{},
		},
		{
			name:            "Set with missing key",
			input:           wrapEvents(rawEventSetMissingKeyValueBar),
			wantedErrorType: &KeyMissing{},
		},
		{
			name:            "Unset with missing key",
			input:           wrapEvents(rawEventUnsetMissingKey),
			wantedErrorType: &KeyMissing{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got []Event
			err := Unmarshal(tc.input, &got)

			if assert.Error(t, err) {
				assert.IsType(t, tc.wantedErrorType, err)
			}
		})
	}
}
func TestUnMarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		wanted []Event
	}{
		{
			name:   "nil input",
			input:  wrapEvents(),
			wanted: []Event{},
		},
		{
			name:   "no events",
			input:  wrapEvents(),
			wanted: []Event{},
		},
		{
			name:  "set foo=bar",
			input: wrapEvents(rawEventSetFooBar),
			wanted: []Event{
				{
					EventType: Set,
					Key:       TestDataFoo,
					Value:     &TestDataBar,
				},
			},
		},
		{
			name:   "set foo=bar, set key=value",
			input:  wrapEvents(rawEventSetFooBar, rawEventSetKeyValue),
			wanted: []Event{TestDataSetFooBar, TestDataSetKeyValue},
		},
		{
			name:   "set key=value, unset key",
			input:  wrapEvents(rawEventSetKeyValue, rawEventUnsetKey),
			wanted: []Event{TestDataSetKeyValue, TestDataUnsetKey},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got []Event
			err := Unmarshal(tc.input, &got)

			assert.NoError(t, err)
			assert.Equal(t, tc.wanted, got)
		})
	}
}
