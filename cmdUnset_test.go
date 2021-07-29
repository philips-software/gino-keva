package main

import (
	"context"
	"testing"

	"github.com/philips-software/gino-keva/internal/event"
	"github.com/stretchr/testify/assert"
)

func TestUnsetCommand(t *testing.T) {
	testCases := []struct {
		name   string
		args   []string
		start  []event.Event
		wanted []event.Event
	}{
		{
			name:   "Unset key",
			start:  []event.Event{},
			args:   []string{"unset", "key"},
			wanted: []event.Event{event.TestDataUnsetKey},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			starteventsJSON, _ := event.Marshal(&tc.start)
			wantedeventsJSON, _ := event.Marshal(&tc.wanted)
			root := NewRootCommand()

			var notesAddArgMsg string
			gitWrapper := &notesStub{
				logCommitsImplementation:   responseStubArgsNone(simpleLogCommitsResponse),
				notesListImplementation:    responseStubArgsString(simpleNotesListResponse),
				notesAddImplementation:     spyArgsStringString(nil, nil, &notesAddArgMsg),
				notesShowImplementation:    responseStubArgsStringString(starteventsJSON),
				revParseHeadImplementation: responseStubArgsNone(TestDataDummyHash),
			}

			ctx := ContextWithGitWrapper(context.Background(), gitWrapper)

			args := disableFetch(tc.args)
			_, err := executeCommandContext(ctx, root, args...)

			assert.NoError(t, err)
			assert.Equal(t, wantedeventsJSON, notesAddArgMsg)
		})
	}
}

func TestUnsetInvalidKey(t *testing.T) {
	gitWrapper := &notesStub{
		notesAddImplementation:     dummyStubArgsStringString,
		revParseHeadImplementation: responseStubArgsNone(TestDataDummyHash),
		logCommitsImplementation:   dummyStubArgsNone,
		notesListImplementation:    dummyStubArgsString,
		notesShowImplementation:    dummyStubArgsStringString,
	}

	t.Run("Key cannot be empty", func(t *testing.T) {
		err := unset(gitWrapper, TestDataDummyRef, TestDataEmptyString)
		if assert.Error(t, err) {
			assert.IsType(t, &event.InvalidKey{}, err)
		}
	})
}
