package main

import (
	"context"
	"errors"
	"testing"

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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCommand()
			ctx := ContextWithGitWrapper(context.Background(), &notesStub{
				logCommitsImplementation: responseStubArgsNone(simpleLogCommitsResponse),
				notesListImplementation:  responseStubArgsString(simpleNotesListResponse),
				notesShowImplementation:  responseStubArgsStringString(input),
			})

			args := disableFetch(tc.args)
			gotOutput, err := executeCommandContext(ctx, root, args...)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantOutput, gotOutput)
		})
	}
}

func TestGetListOutputTestDataEmpty(t *testing.T) {
	td := testDataEmpty

	testCases := []struct {
		name         string
		outputFormat string
		input        string
		wantText     string
	}{
		{
			name:         "Empty note (plain)",
			outputFormat: "plain",
			input:        td.input,
			wantText:     td.outputPlain,
		},
		{
			name:         "Empty note (json)",
			outputFormat: "json",
			input:        td.input,
			wantText:     td.outputJSON,
		},
		{
			name:         "Empty note (plain)",
			outputFormat: "plain",
			input:        td.input,
			wantText:     td.outputPlain,
		},
		{
			name:         "Empty note (json)",
			outputFormat: "json",
			input:        td.input,
			wantText:     td.outputJSON,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gitWrapper := notesStub{
				logCommitsImplementation: dummyStubArgsNone,
				notesListImplementation:  dummyStubArgsString,
				notesShowImplementation:  responseStubArgsStringString(tc.input),
			}
			gotOutput, err := getListOutput(&gitWrapper, dummyRef, tc.outputFormat)

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
		gitWrapper := &notesStub{
			logCommitsImplementation: dummyStubArgsNone,
			notesListImplementation:  dummyStubArgsString,
			notesShowImplementation:  showStubExhaustedRepo,
		}
		ctx := ContextWithGitWrapper(context.Background(), gitWrapper)

		args := disableFetch([]string{"list"})
		gotOutput, err := executeCommandContext(ctx, root, args...)

		assert.NoError(t, err)
		assert.Equal(t, testDataEmpty.outputPlain, gotOutput)
	})
}

func TestInvalidOutputFormat(t *testing.T) {
	t.Run("InvalidOutputFormat error raised when specifying invalid output format", func(t *testing.T) {
		gitWrapper := notesStub{
			logCommitsImplementation: dummyStubArgsNone,
			notesListImplementation:  dummyStubArgsString,
			notesShowImplementation:  dummyStubArgsStringString,
		}

		_, err := getListOutput(&gitWrapper, dummyRef, "invalid format")
		if assert.Error(t, err) {
			assert.IsType(t, &InvalidOutputFormat{}, err)
		}
	})
}
