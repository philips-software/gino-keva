package main

import (
	"context"
	"testing"

	"github.com/philips-software/gino-keva/internal/event"
	"github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
	td := []event.Event{event.TestDataSetKeyValue}

	testCases := []struct {
		name       string
		args       []string
		start      []event.Event
		input      string
		wantOutput string
	}{
		{
			name:       "List all notes (plain output)",
			args:       []string{"list"},
			start:      td,
			wantOutput: "key=value\n",
		},
		{
			name:       "List all notes (json output)",
			args:       []string{"list", "--output", "json"},
			start:      td,
			wantOutput: "{\n  \"key\": \"value\"\n}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eventsJSON, _ := event.Marshal(&tc.start)

			root := NewRootCommand()
			ctx := ContextWithGitWrapper(context.Background(), &notesStub{
				logCommitsImplementation: responseStubArgsNone(simpleLogCommitsResponse),
				notesListImplementation:  responseStubArgsString(simpleNotesListResponse),
				notesShowImplementation:  responseStubArgsStringString(eventsJSON),
			})

			args := disableFetch(tc.args)
			gotOutput, err := executeCommandContext(ctx, root, args...)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantOutput, gotOutput)
		})
	}
}

func TestGetListOutputTestDataEmpty(t *testing.T) {
	td := []event.Event{}

	testCases := []struct {
		name         string
		outputFormat string
		start        []event.Event
		wantText     string
	}{
		{
			name:         "Empty note (plain)",
			outputFormat: "plain",
			start:        td,
			wantText:     TestDataEmptyString,
		},
		{
			name:         "Empty note (json)",
			outputFormat: "json",
			start:        td,
			wantText:     "{}\n",
		},
		{
			name:         "Empty note (plain)",
			outputFormat: "plain",
			start:        td,
			wantText:     TestDataEmptyString,
		},
		{
			name:         "Empty note (json)",
			outputFormat: "json",
			start:        td,
			wantText:     "{}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eventsJSON, _ := event.Marshal(&tc.start)

			gitWrapper := notesStub{
				logCommitsImplementation: dummyStubArgsNone,
				notesListImplementation:  dummyStubArgsString,
				notesShowImplementation:  responseStubArgsStringString(eventsJSON),
			}
			gotOutput, err := getListOutput(&gitWrapper, TestDataDummyRef, tc.outputFormat)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantText, gotOutput)
		})
	}
}

func TestInvalidOutputFormat(t *testing.T) {
	t.Run("InvalidOutputFormat error raised when specifying invalid output format", func(t *testing.T) {
		gitWrapper := notesStub{
			logCommitsImplementation: dummyStubArgsNone,
			notesListImplementation:  dummyStubArgsString,
			notesShowImplementation:  dummyStubArgsStringString,
		}

		_, err := getListOutput(&gitWrapper, TestDataDummyRef, "invalid format")
		if assert.Error(t, err) {
			assert.IsType(t, &InvalidOutputFormat{}, err)
		}
	})
}
