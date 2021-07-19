package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsetCommand(t *testing.T) {
	testCases := []struct {
		name   string
		args   []string
		start  string
		source string
		wanted string
	}{
		{
			name:   "Unset foo",
			start:  testDataKeyValueFooBar.inputNew,
			args:   []string{"unset", "foo"},
			wanted: testDataKeyValue.outputRaw,
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

func TestUnsetValue(t *testing.T) {
	testCases := []struct {
		name   string
		start  string
		key    string
		value  Value
		wanted string
	}{
		{
			name:   "Unset non-existing key has no effect",
			start:  testDataEmpty.inputNew,
			key:    "non_existing_key",
			wanted: testDataEmpty.outputRaw,
		},
		{
			name:   "Unset foo doesn't affect other key/value",
			start:  testDataKeyValueFooBar.inputNew,
			key:    "foo",
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

			err := unset(gitWrapper, dummyRef, tc.key, 0)
			assert.NoError(t, err)
			assert.Equal(t, tc.wanted, notesAddArgMsg)
		})
	}
}
