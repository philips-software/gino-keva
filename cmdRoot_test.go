package main

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlagResolution(t *testing.T) {
	testCases := []struct {
		name       string
		envVar     string
		flagArgs   []string
		wantOutput string
	}{
		{
			name:       "If no flag value is provided, the default value (gino_keva) should be used",
			envVar:     "",
			flagArgs:   []string{},
			wantOutput: "gino_keva",
		},
		{
			name:       "Set notes ref flag via command line",
			envVar:     "",
			flagArgs:   []string{"--ref", "From_Command_Line"},
			wantOutput: "From_Command_Line",
		},
		{
			name:       "Set notes ref flag with an environment variable",
			envVar:     "From_Environment",
			flagArgs:   []string{},
			wantOutput: "From_Environment",
		},
		{
			name:       "Overwrite notes ref flag set in environment via command line",
			envVar:     "From_Environment",
			flagArgs:   []string{"--ref", "From_Command_Line"},
			wantOutput: "From_Command_Line",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCommand()

			os.Setenv("GINO_KEVA_REF", tc.envVar)
			defer os.Unsetenv("GINO_KEVA_REF")

			listFlagArgs := []string{"show-flag", "ref"}
			args := append(listFlagArgs, tc.flagArgs...)

			ctx := ContextWithGitWrapper(context.Background(), &notesStub{})
			gotOutput, err := executeCommandContext(ctx, root, args...)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantOutput, gotOutput)
		})
	}
}

func TestFetchFlag(t *testing.T) {
	testCases := []struct {
		name            string
		args            []string
		wantFetchCalled bool
	}{
		{
			name:            "Fetch is called when calling list (default)",
			args:            []string{"list"},
			wantFetchCalled: true,
		},
		{
			name:            "Fetch is NOT called if set to false",
			args:            disableFetch([]string{"list"}),
			wantFetchCalled: false,
		}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var fetchCalled bool
			gitWrapper := &notesStub{
				fetchNotesImplementation: spyArgsString(&fetchCalled, nil),
				logCommitsImplementation: dummyStubArgsNone,
				notesListImplementation:  dummyStubArgsString,
				notesShowImplementation:  dummyStubArgsStringString,
			}
			ctx := ContextWithGitWrapper(context.Background(), gitWrapper)

			root := NewRootCommand()
			_, err := executeCommandContext(ctx, root, tc.args...)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantFetchCalled, fetchCalled)
		})
	}
}

func TestPushFlag(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		wantPushCalled bool
	}{
		{
			name:           "Push is NOT called when flag is unset (default)",
			args:           []string{"set", "foo", "bar"},
			wantPushCalled: false,
		},
		{
			name:           "Push is called when flag is set",
			args:           enablePush([]string{"set", "foo", "bar"}),
			wantPushCalled: true,
		}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var pushCalled bool
			gitWrapper := &notesStub{
				pushNotesImplementation:    spyArgsString(&pushCalled, nil),
				revParseHeadImplementation: responseStubArgsNone(TestDataDummyHash),
				logCommitsImplementation:   dummyStubArgsNone,
				notesAddImplementation:     dummyStubArgsStringString,
				notesListImplementation:    dummyStubArgsString,
				notesShowImplementation:    dummyStubArgsStringString,
			}
			ctx := ContextWithGitWrapper(context.Background(), gitWrapper)

			root := NewRootCommand()

			args := disableFetch(tc.args)
			_, err := executeCommandContext(ctx, root, args...)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantPushCalled, pushCalled)
		})
	}
}

func TestFetchNoUpstreamRef(t *testing.T) {
	var fetchStubNoUpstreamRef = func(string) (response string, err error) {
		// Mimic error when there's no upstream ref present yet
		err = errors.New("exit status 128")
		response = "fatal: couldn't find remote ref refs/notes/foo"

		return response, err
	}

	t.Run("Fetch without upstream notesref doesn't result in error", func(t *testing.T) {
		root := NewRootCommand()
		gitWrapper := &notesStub{
			fetchNotesImplementation: fetchStubNoUpstreamRef,
			logCommitsImplementation: dummyStubArgsNone,
			notesListImplementation:  dummyStubArgsString,
			notesShowImplementation:  dummyStubArgsStringString,
		}
		ctx := ContextWithGitWrapper(context.Background(), gitWrapper)

		args := []string{"list"}
		gotOutput, err := executeCommandContext(ctx, root, args...)

		assert.NoError(t, err)
		assert.Equal(t, TestDataEmptyString, gotOutput)
	})
}
