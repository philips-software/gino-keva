package main

import (
	"context"
	"errors"
	"testing"

	"github.com/philips-software/gino-keva/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
	td := testDataKeyValue
	input := td.input

	testCases := []struct {
		name       string
		args       []string
		input      string
		wantOutput string
	}{
		{
			name:       "List all notes (plain output)",
			args:       []string{"list"},
			wantOutput: td.outputPlain,
		},
		{
			name:       "List all notes (json output)",
			args:       []string{"list", "--output", "json"},
			wantOutput: td.outputJSON,
		},
		{
			name:       "List all notes (raw output)",
			args:       []string{"list", "--output", "raw"},
			wantOutput: td.outputRaw,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCommand()
			ctx := git.ContextWithNotes(context.Background(), &notesStub{
				addImplementation:          panicStubInputsStringString,
				fetchImplementation:        dummyStubInputsString,
				pushImplementation:         dummyStubInputsString,
				revParseHeadImplementation: panicStubInputsNone,
				showImplementation:         showStubReturnResponseAtDepth(input, 0),
			})
			gotOutput, err := executeCommandContext(ctx, root, tc.args...)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantOutput, gotOutput)
		})
	}
}

func TestGetListOutputTestDataEmpty(t *testing.T) {
	td := testDataEmpty
	input := td.input

	testCases := []struct {
		name         string
		outputFormat string
		wantText     string
	}{
		{
			name:         "Empty note (plain)",
			outputFormat: "plain",
			wantText:     td.outputPlain,
		},
		{
			name:         "Empty note (json)",
			outputFormat: "json",
			wantText:     td.outputJSON,
		},
		{
			name:         "Empty note (raw)",
			outputFormat: "raw",
			wantText:     td.outputRaw,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notesAccess := notesStub{
				addImplementation:          panicStubInputsStringString,
				fetchImplementation:        dummyStubInputsString,
				pushImplementation:         dummyStubInputsString,
				revParseHeadImplementation: panicStubInputsNone,
				showImplementation:         showStubReturnResponseAtDepth(input, 0),
			}
			gotOutput, err := getListOutput(&notesAccess, "dummyRef", 0, tc.outputFormat)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantText, gotOutput)
		})
	}
}

func TestNoNotesLimitedRepoDepth(t *testing.T) {
	var showStubExhaustedRepo = func(string, string) (response string, err error) {
		// Mimic error when repository is exhausted without encountering a note
		err = errors.New("exit status 128")
		response = "fatal: failed to resolve 'FOO' as a valid ref."

		return response, err
	}

	t.Run("Small repository without prior notes doesn't result in error", func(t *testing.T) {
		root := NewRootCommand()
		notesAccess := &notesStub{
			addImplementation:          panicStubInputsStringString,
			fetchImplementation:        dummyStubInputsString,
			pushImplementation:         dummyStubInputsString,
			revParseHeadImplementation: panicStubInputsNone,
			showImplementation:         showStubExhaustedRepo,
		}
		ctx := git.ContextWithNotes(context.Background(), notesAccess)

		args := []string{"list"}
		gotOutput, err := executeCommandContext(ctx, root, args...)

		assert.NoError(t, err)
		assert.Equal(t, testDataEmpty.outputPlain, gotOutput)
	})
}

func TestInvalidOutputFormat(t *testing.T) {
	t.Run("InvalidOutputFormat error raised when specifying invalid output format", func(t *testing.T) {
		notesAccess := notesStub{
			showImplementation: dummyStubInputsStringString,
		}

		_, err := getListOutput(&notesAccess, "dummyRef", 0, "invalid format")
		if assert.Error(t, err) {
			assert.IsType(t, &InvalidOutputFormat{}, err)
		}
	})
}

func TestGetListOutputTestDataKeyValue(t *testing.T) {
	td := testDataKeyValue
	input := td.input
	maxDepth := 5

	testCases := []struct {
		name     string
		depth    int
		wantText string
	}{
		{
			name:     "Note on current commit",
			depth:    0,
			wantText: td.outputPlain,
		},
		{
			name:     "Note on parent commit",
			depth:    1,
			wantText: td.outputPlain,
		},
		{
			name:     "Note at max search depth",
			depth:    5,
			wantText: td.outputPlain,
		},
		{
			name:     "Note beyond max search depth",
			depth:    6,
			wantText: testDataEmpty.outputPlain,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			notesAccess := notesStub{
				addImplementation:          panicStubInputsStringString,
				fetchImplementation:        dummyStubInputsString,
				pushImplementation:         dummyStubInputsString,
				revParseHeadImplementation: panicStubInputsNone,
				showImplementation:         showStubReturnResponseAtDepth(input, tc.depth),
			}
			gotOutput, err := getListOutput(&notesAccess, "dummyRef", maxDepth, "plain")

			assert.NoError(t, err)
			assert.Equal(t, tc.wantText, gotOutput)
		})
	}
}
