package main

import (
	"context"
	"testing"

	"github.com/philips-software/gino-keva/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestSetCommand(t *testing.T) {
	testCases := []struct {
		name   string
		args   []string
		start  string
		source string
		wanted string
	}{
		{
			name:   "Start empty, set key=value (default ref)",
			start:  testDataEmpty.input,
			args:   []string{"set", "key", "value"},
			source: "01234567",
			wanted: testDataKeyValue.outputJSON,
		},
		{
			name:   "Start empty, set key=value (non-default ref)",
			start:  testDataEmpty.input,
			args:   []string{"set", "key", "value", "--ref", "non_default"},
			source: "01234567",
			wanted: testDataKeyValue.outputJSON,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCommand()
			notesAccess := &notesAddSpy{
				revParseHeadResponse: tc.source,
				showResponse:         tc.start,
			}
			ctx := git.ContextWithNotes(context.Background(), notesAccess)

			_, err := executeCommandContext(ctx, root, tc.args...)

			assert.NoError(t, err)
			assert.Equal(t, tc.wanted, notesAccess.AddResult)
		})
	}
}

func TestInvalidKeys(t *testing.T) {
	testCases := []struct {
		name       string
		invalidKey string
	}{
		{
			name:       "Key cannot be empty",
			invalidKey: "",
		},
		{
			name:       "Key contains an invalid character",
			invalidKey: "foo!",
		},
		{
			name:       "First character of key is not a letter",
			invalidKey: "2BeOrNot2Be",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notesAccess := &notesStub{
				addImplementation:          dummyStubInputsStringString,
				revParseHeadImplementation: dummyStubInputsNone,
				showImplementation:         dummyStubInputsStringString,
			}

			err := set(notesAccess, "dummyRef", tc.invalidKey, "dummyValue", 0)
			if assert.Error(t, err) {
				assert.IsType(t, &InvalidKey{}, err)
			}
		})
	}
}

func TestSet(t *testing.T) {
	testCases := []struct {
		name   string
		start  string
		key    string
		value  Value
		wanted string
	}{
		{
			name:   "Start empty, set key=value",
			start:  testDataEmpty.input,
			key:    "key",
			value:  Value{Data: "value", Source: "01234567"},
			wanted: testDataKeyValue.outputJSON,
		},
		{
			name:   "Start key=value, set foo=bar",
			start:  testDataKeyValue.input,
			key:    "foo",
			value:  Value{Data: "bar", Source: "abcd1234"},
			wanted: testDataKeyValueFooBar.outputJSON,
		},
		{
			name:   "Source hash is cut off at 8 characters",
			start:  testDataEmpty.input,
			key:    "key",
			value:  Value{Data: "value", Source: "01234567_and_the_remainder"},
			wanted: testDataKeyValue.outputJSON,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notesAccess := &notesAddSpy{
				revParseHeadResponse: tc.value.Source,
				showResponse:         tc.start,
			}

			err := set(notesAccess, "dummyRef", tc.key, tc.value.Data, 0)
			assert.NoError(t, err)
			assert.Equal(t, tc.wanted, notesAccess.AddResult)
		})
	}
}
