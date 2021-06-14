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
			args:   []string{"set", "MY_key", "value"},
			source: "01234567",
			wanted: testDataKeyValue.outputRaw,
		},
		{
			name:   "Start empty, set key=value (non-default ref)",
			start:  testDataEmpty.input,
			args:   []string{"set", "MY_KEY", "value", "--ref", "non_default"},
			source: "01234567",
			wanted: testDataKeyValue.outputRaw,
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
		name  string
		key   string
		valid bool
	}{
		{
			name:  "Key can container underscores and dashes",
			key:   "This-key_is-valid",
			valid: true,
		},
		{
			name:  "Key cannot be empty",
			key:   "",
			valid: false,
		},
		{
			name:  "Key contains an invalid character",
			key:   "foo!",
			valid: false,
		},
		{
			name:  "First character of key is not a letter",
			key:   "2BeOrNot2Be",
			valid: false,
		},
		{
			name:  "Last character of key is not a letter",
			key:   "invalid-",
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notesAccess := &notesStub{
				addImplementation:          dummyStubInputsStringString,
				revParseHeadImplementation: dummyStubInputsNone,
				showImplementation:         dummyStubInputsStringString,
			}

			err := set(notesAccess, "dummyRef", tc.key, "dummyValue", 0)
			if tc.valid {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) {
					assert.IsType(t, &InvalidKey{}, err)
				}
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
			name:   "Start empty, set MY_KEY=value",
			start:  testDataEmpty.input,
			key:    "my-key",
			value:  Value{Data: "value", Source: "01234567"},
			wanted: testDataKeyValue.outputRaw,
		},
		{
			name:   "Start MY_KEY=value, set foo=bar",
			start:  testDataKeyValue.input,
			key:    "foo",
			value:  Value{Data: "bar", Source: "abcd1234"},
			wanted: testDataKeyValueFooBar.outputRaw,
		},
		{
			name:   "Source hash is cut off at 8 characters",
			start:  testDataEmpty.input,
			key:    "MY_KEY",
			value:  Value{Data: "value", Source: "01234567_and_the_remainder"},
			wanted: testDataKeyValue.outputRaw,
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
