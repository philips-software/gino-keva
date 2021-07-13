package main

import (
	"context"
	"testing"

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
			name:   "Start empty, set MY_KEY=value (default ref)",
			start:  testDataEmpty.inputOld,
			args:   []string{"set", "my_key", "value"},
			source: "01234567",
			wanted: testDataKeyValue.outputRaw,
		},
		{
			name:   "Start MY_KEY=value, set foo=bar (non-default ref)",
			start:  testDataKeyValue.inputNew,
			args:   []string{"set", "foo", "bar", "--ref", "non_default"},
			source: "abcd1234",
			wanted: testDataKeyValueFooBar.outputRaw,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCommand()
			var notesAddArgMsg string
			gitWrapper := &notesStub{
				logCommitsImplementation:   responseStubArgsNone(simpleLogCommitsResponse),
				notesListImplementation:    responseStubArgsString(simpleNotesListResponse),
				notesAddImplementation:     spyArgsStringString(nil, nil, &notesAddArgMsg),
				notesShowImplementation:    responseStubArgsStringString(tc.start),
				revParseHeadImplementation: responseStubArgsNone(tc.source),
			}
			ctx := ContextWithGitWrapper(context.Background(), gitWrapper)

			args := disableFetch(tc.args)
			_, err := executeCommandContext(ctx, root, args...)

			assert.NoError(t, err)
			assert.Equal(t, tc.wanted, notesAddArgMsg)
		})
	}
}

func TestKeyValidation(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		valid bool
	}{
		{
			name:  "Key can end on number",
			key:   "TheAnswerIs42",
			valid: true,
		},
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
			name:  "Last character of key is not a letter of number",
			key:   "invalid-",
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gitWrapper := &notesStub{
				notesAddImplementation:     dummyStubArgsStringString,
				revParseHeadImplementation: dummyStubArgsNone,
				logCommitsImplementation:   dummyStubArgsNone,
				notesListImplementation:    dummyStubArgsString,
				notesShowImplementation:    dummyStubArgsStringString,
			}

			err := set(gitWrapper, dummyRef, tc.key, dummyValue, 0)
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
			start:  testDataEmpty.inputOld,
			key:    "my-key",
			value:  Value{Data: "value", Source: "01234567"},
			wanted: testDataKeyValue.outputRaw,
		},
		{
			name:   "Start MY_KEY=value, set foo=bar",
			start:  testDataKeyValue.inputNew,
			key:    "foo",
			value:  Value{Data: "bar", Source: "abcd1234"},
			wanted: testDataKeyValueFooBar.outputRaw,
		},
		{
			name:   "Source hash is cut off at 8 characters",
			start:  testDataEmpty.inputNew,
			key:    "MY_KEY",
			value:  Value{Data: "value", Source: "01234567_and_the_remainder"},
			wanted: testDataKeyValue.outputRaw,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var notesAddArgMsg string
			gitWrapper := &notesStub{
				logCommitsImplementation:   responseStubArgsNone(simpleLogCommitsResponse),
				notesListImplementation:    responseStubArgsString(simpleNotesListResponse),
				notesAddImplementation:     spyArgsStringString(nil, nil, &notesAddArgMsg),
				notesShowImplementation:    responseStubArgsStringString(tc.start),
				revParseHeadImplementation: responseStubArgsNone(tc.value.Source),
			}

			err := set(gitWrapper, dummyRef, tc.key, tc.value.Data, 0)
			assert.NoError(t, err)
			assert.Equal(t, tc.wanted, notesAddArgMsg)
		})
	}
}
