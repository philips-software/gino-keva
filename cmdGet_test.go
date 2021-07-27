package main

import (
	"context"
	"testing"

	"github.com/philips-software/gino-keva/internal/event"
	"github.com/stretchr/testify/assert"
)

func TestGetCommand(t *testing.T) {

	testCases := []struct {
		name       string
		args       []string
		start      []event.Event
		wantOutput string
	}{
		{
			name:       "Get value of a key",
			args:       []string{"get", "key"},
			start:      []event.Event{event.TestDataSetKeyValue},
			wantOutput: "value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eventsJson, _ := event.Marshal(&tc.start)

			root := NewRootCommand()
			ctx := ContextWithGitWrapper(context.Background(), &notesStub{
				logCommitsImplementation: responseStubArgsNone(simpleLogCommitsResponse),
				notesListImplementation:  responseStubArgsString(simpleNotesListResponse),
				notesShowImplementation:  responseStubArgsStringString(eventsJson),
			})
			args := disableFetch(tc.args)
			gotOutput, err := executeCommandContext(ctx, root, args...)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantOutput, gotOutput)
		})
	}
}

func TestGetValue(t *testing.T) {
	testCases := []struct {
		name      string
		key       string
		depth     uint
		start     []event.Event
		wantValue string
	}{
		{
			name:      "Get value of an existing key",
			key:       "key",
			depth:     0,
			start:     []event.Event{event.TestDataSetKeyValue},
			wantValue: "value",
		},
		{
			name:      "Get value of a non-existing key",
			key:       "nonExistingKey",
			depth:     0,
			start:     []event.Event{event.TestDataSetKeyValue},
			wantValue: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eventsJson, _ := event.Marshal(&[]event.Event{event.TestDataSetKeyValue})

			gitWrapper := notesStub{
				logCommitsImplementation: responseStubArgsNone(simpleLogCommitsResponse),
				notesListImplementation:  responseStubArgsString(simpleNotesListResponse),
				notesShowImplementation:  responseStubArgsStringString(eventsJson),
			}
			gotValue, err := getValue(&gitWrapper, TestDataDummyRef, tc.key)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantValue, gotValue)
		})
	}
}
