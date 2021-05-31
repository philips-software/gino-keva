package main

import (
	"context"
	"testing"

	"github.com/philips-internal/gino-keva/internal/git"
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
			start:  testDataKeyValueFooBar.input,
			args:   []string{"unset", "foo"},
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
			start:  testDataEmpty.input,
			key:    "non_existing_key",
			wanted: testDataEmpty.outputJSON,
		},
		{
			name:   "Unset foo doesn't affect other key/value",
			start:  testDataKeyValueFooBar.input,
			key:    "foo",
			wanted: testDataKeyValue.outputJSON,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notesAccess := &notesAddSpy{
				revParseHeadResponse: tc.value.Source,
				showResponse:         tc.start,
			}

			err := unset(notesAccess, "dummyRef", tc.key, 0)
			assert.NoError(t, err)
			assert.Equal(t, tc.wanted, notesAccess.AddResult)
		})
	}
}
