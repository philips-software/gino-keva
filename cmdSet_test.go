package main

import (
	"context"
	"errors"
	"testing"

	"github.com/philips-software/gino-keva/internal/event"
	"github.com/stretchr/testify/assert"
)

func TestSetCommand(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		startEvents  []event.Event
		wantedEvents []event.Event
	}{
		{
			name:         "Start empty, set key=value (default ref)",
			startEvents:  []event.Event{},
			args:         []string{"set", "key", "value"},
			wantedEvents: []event.Event{event.TestDataSetKeyValue},
		},
		{
			name:         "Start key=value, set foo=bar (non-default ref)",
			startEvents:  []event.Event{event.TestDataSetKeyValue},
			args:         []string{"set", "foo", "bar", "--ref", "non_default"},
			wantedEvents: []event.Event{event.TestDataSetKeyValue, event.TestDataSetFooBar},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			startEventsJson, _ := event.Marshal(&tc.startEvents)
			wantedEventsJson, _ := event.Marshal(&tc.wantedEvents)

			root := NewRootCommand()
			var notesAddArgMsg string
			gitWrapper := &notesStub{
				notesAddImplementation:     spyArgsStringString(nil, nil, &notesAddArgMsg),
				notesShowImplementation:    responseStubArgsStringString(startEventsJson),
				revParseHeadImplementation: responseStubArgsNone(TestDataDummyHash),
			}
			ctx := ContextWithGitWrapper(context.Background(), gitWrapper)

			args := disableFetch(tc.args)
			_, err := executeCommandContext(ctx, root, args...)

			assert.NoError(t, err)
			assert.Equal(t, wantedEventsJson, notesAddArgMsg)
		})
	}
}

func TestSetInvalidKey(t *testing.T) {
	gitWrapper := &notesStub{}

	t.Run("Key cannot be empty", func(t *testing.T) {
		err := set(gitWrapper, TestDataDummyRef, TestDataEmptyString, TestDataDummyValue)
		if assert.Error(t, err) {
			assert.IsType(t, &event.InvalidKey{}, err)
		}
	})
}

func TestSetWithoutHeadEvents(t *testing.T) {
	var showStubNoNote = func(string, string) (response string, err error) {
		// Mimic error when ther's no note available
		err = errors.New("exit status 128")
		response = "error: no note found for object DUMMY_HASH.\n"

		return response, err
	}

	gitWrapper := &notesStub{
		notesShowImplementation:    showStubNoNote,
		revParseHeadImplementation: responseStubArgsNone(TestDataDummyHash),
		notesAddImplementation:     dummyStubArgsStringString,
	}

	t.Run("Set without prior events on HEAD commit doesn't fail", func(t *testing.T) {
		err := set(gitWrapper, TestDataDummyRef, event.TestDataFoo, event.TestDataBar)
		assert.NoError(t, err)
	})
}
