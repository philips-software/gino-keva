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
			ctx := git.ContextWithGitWrapper(context.Background(), &notesStub{
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
			gitWrapper := notesStub{
				logCommitsImplementation: dummyStubArgsNone,
				notesListImplementation:  dummyStubArgsString,
				notesShowImplementation:  responseStubArgsStringString(input),
			}
			gotOutput, err := getListOutput(&gitWrapper, dummyRef, 0, tc.outputFormat)

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
		ctx := git.ContextWithGitWrapper(context.Background(), gitWrapper)

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

		_, err := getListOutput(&gitWrapper, dummyRef, 0, "invalid format")
		if assert.Error(t, err) {
			assert.IsType(t, &InvalidOutputFormat{}, err)
		}
	})
}
