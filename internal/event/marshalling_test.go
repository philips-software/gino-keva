package event

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var eventsWrapper = "{\"events\":[%v]}\n"

var foo = "foo"
var bar = "bar"
var key = "key"
var value = "value"

var eventSetFooBar = "{\"type\":\"set\",\"key\":\"foo\",\"value\":\"bar\"}"
var eventSetKeyValue = "{\"type\":\"set\",\"key\":\"key\",\"value\":\"value\"}"
var eventUnsetKey = "{\"type\":\"unset\",\"key\":\"key\"}"

var eventUnknown = "{\"type\":\"unknown\"}"
var eventSetFooMissingValue = "{\"type\":\"set\",\"key\":\"foo\"}"
var eventSetMissingKeyValueBar = "{\"type\":\"set\",\"value\":\"bar\"}"
var eventUnsetMissingKey = "{\"type\":\"unset\"}"

func TestMarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  *[]Event
		wanted []string
	}{
		{
			name:   "nil input",
			input:  nil,
			wanted: []string{},
		},
		{
			name:   "no events",
			input:  &[]Event{},
			wanted: []string{},
		},
		{
			name: "set foo=bar",
			input: &[]Event{
				{
					EventType: Set,
					Key:       foo,
					Value:     &bar,
				},
			},
			wanted: []string{eventSetFooBar},
		},
		{
			name: "set foo=bar, set key=value",
			input: &[]Event{
				{
					EventType: Set,
					Key:       foo,
					Value:     &bar,
				},
				{
					EventType: Set,
					Key:       key,
					Value:     &value,
				},
			},
			wanted: []string{eventSetFooBar, eventSetKeyValue},
		},
		{
			name: "set key=value, unset key",
			input: &[]Event{
				{
					EventType: Set,
					Key:       key,
					Value:     &value,
				},
				{
					EventType: Unset,
					Key:       key,
				},
			},
			wanted: []string{eventSetKeyValue, eventUnsetKey},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wanted := fmt.Sprintf(eventsWrapper, strings.Join(tc.wanted, ","))
			got, err := Marshal(tc.input)

			assert.NoError(t, err)
			assert.Equal(t, wanted, got)
		})
	}
}

func TestUnMarshalInvalidFormat(t *testing.T) {
	testCases := []struct {
		name            string
		input           []string
		wantedErrorType error
	}{
		{
			name:            "unknown event",
			input:           []string{eventUnknown},
			wantedErrorType: &UnknownType{},
		},
		{
			name:            "Set with missing value",
			input:           []string{eventSetFooMissingValue},
			wantedErrorType: &ValueMissing{},
		},
		{
			name:            "Set with missing key",
			input:           []string{eventSetMissingKeyValueBar},
			wantedErrorType: &KeyMissing{},
		},
		{
			name:            "Unset with missing key",
			input:           []string{eventUnsetMissingKey},
			wantedErrorType: &KeyMissing{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := fmt.Sprintf(eventsWrapper, strings.Join(tc.input, ","))
			var got []Event
			err := Unmarshal(input, &got)

			if assert.Error(t, err) {
				assert.IsType(t, tc.wantedErrorType, err)
			}
		})
	}
}
func TestUnMarshal(t *testing.T) {
	testCases := []struct {
		name   string
		input  []string
		wanted []Event
	}{
		{
			name:   "nil input",
			input:  nil,
			wanted: []Event{},
		},
		{
			name:   "no events",
			input:  []string{},
			wanted: []Event{},
		},
		{
			name:  "set foo=bar",
			input: []string{eventSetFooBar},
			wanted: []Event{
				{
					EventType: Set,
					Key:       foo,
					Value:     &bar,
				},
			},
		},
		{
			name:  "set foo=bar, set key=value",
			input: []string{eventSetFooBar, eventSetKeyValue},
			wanted: []Event{
				{
					EventType: Set,
					Key:       foo,
					Value:     &bar,
				},
				{
					EventType: Set,
					Key:       key,
					Value:     &value,
				},
			},
		},
		{
			name:  "set key=value, unset key",
			input: []string{eventSetKeyValue, eventUnsetKey},
			wanted: []Event{
				{
					EventType: Set,
					Key:       key,
					Value:     &value,
				},
				{
					EventType: Unset,
					Key:       key,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := fmt.Sprintf(eventsWrapper, strings.Join(tc.input, ","))
			var got []Event
			err := Unmarshal(input, &got)

			assert.NoError(t, err)
			assert.Equal(t, tc.wanted, got)
		})
	}
}
