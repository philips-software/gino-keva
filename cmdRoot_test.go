package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/philips-software/gino-keva/internal/git"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type notesStub struct {
	fetchNotesImplementation   func(string) (string, error)
	logCommitsImplementation   func() (string, error)
	notesAddImplementation     func(string, string) (string, error)
	notesListImplementation    func(string) (string, error)
	notesShowImplementation    func(string, string) (string, error)
	pushNotesImplementation    func(string) (string, error)
	revParseHeadImplementation func() (string, error)
}

// Provide simple implementations for logCommits and notesList for testing purposes
var (
	simpleLogCommitsResponse = "COMMIT_REFERENCE\n"
	simpleNotesListResponse  = "NOTES_OBJECT_ID COMMIT_REFERENCE\n"
)

// FetchNotes test-double
func (n notesStub) FetchNotes(notesRef string, force bool) (string, error) {
	return n.fetchNotesImplementation(notesRef)
}

// LogCommits test-double
func (n notesStub) LogCommits(uint) (string, error) {
	return n.logCommitsImplementation()
}

// NotesAdd test-double
func (n notesStub) NotesAdd(notesRef string, msg string) (string, error) {
	return n.notesAddImplementation(notesRef, msg)
}

// NotesList test-double
func (n notesStub) NotesList(notesRef string) (string, error) {
	return n.notesListImplementation(notesRef)
}

// NotesPrune dummy
func (notesStub) NotesPrune(string) (string, error) {
	return "", nil
}

//NotesShow test-double calls the stub implementation
func (n *notesStub) NotesShow(notesRef, hash string) (response string, err error) {
	return n.notesShowImplementation(notesRef, hash)
}

// PushNotes test-double
func (n notesStub) PushNotes(notesRef string) (string, error) {
	return n.pushNotesImplementation(notesRef)
}

// RevParseHead test-double
func (n notesStub) RevParseHead() (string, error) {
	return n.revParseHeadImplementation()
}

var dummyStubArgsNone = func() (string, error) { return "", nil }
var dummyStubArgsString = func(string) (string, error) { return "", nil }
var dummyStubArgsStringString = func(string, string) (string, error) { return "", nil }

var responseStubArgsNone = func(expectedResponse string) func() (string, error) {
	return func() (string, error) {
		return expectedResponse, nil
	}
}

var responseStubArgsString = func(expectedResponse string) func(string) (string, error) {
	return func(string) (string, error) {
		return expectedResponse, nil
	}
}

var responseStubArgsStringString = func(expectedResponse string) func(string, string) (string, error) {
	return func(string, string) (string, error) {
		return expectedResponse, nil
	}
}

var spyArgsString = func(isCalled *bool, arg1 *string) func(string) (string, error) {
	return func(a1 string) (string, error) {
		if isCalled != nil {
			*isCalled = true
		}
		if arg1 != nil {
			*arg1 = a1
		}
		return "", nil
	}
}

var spyArgsStringString = func(isCalled *bool, arg1, arg2 *string) func(string, string) (string, error) {
	return func(a1, a2 string) (string, error) {
		if isCalled != nil {
			*isCalled = true
		}
		if arg1 != nil {
			*arg1 = a1
		}
		if arg2 != nil {
			*arg2 = a2
		}
		return "", nil
	}
}

type testData struct {
	input       string
	outputRaw   string
	outputPlain string
	outputJSON  string
}

var testDataEmpty = testData{
	input:       `{}`,
	outputRaw:   "{}\n",
	outputPlain: "",
	outputJSON:  "{}\n",
}

var testDataKeyValue = testData{
	input:       `{"MY_KEY": {"data":"value", "source": "01234567"}}`,
	outputRaw:   "{\"MY_KEY\":{\"data\":\"value\",\"source\":\"01234567\"}}\n",
	outputPlain: "MY_KEY=value\n",
	outputJSON:  "{\n  \"MY_KEY\": \"value\"\n}\n",
}

var testDataKeyValueFooBar = testData{
	input:       `{"FOO": {"data":"bar", "source": "abcd1234"},"MY_KEY": {"data":"value", "source": "01234567"}}`,
	outputRaw:   "{\"FOO\":{\"data\":\"bar\",\"source\":\"abcd1234\"},\"MY_KEY\":{\"data\":\"value\",\"source\":\"01234567\"}}\n",
	outputPlain: "MY_KEY=value\nFOO=bar\n",
	outputJSON:  "{\n  \"FOO\": \"bar\",\n  \"MY_KEY\": \"value\"\n}\n",
}

func executeCommandContext(ctx context.Context, root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)

	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.ExecuteContext(ctx)
	return buf.String(), err
}

func disableFetch(args []string) []string {
	return append(args, "--fetch=false")
}

func enablePush(args []string) []string {
	return append(args, "--push")
}

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

			ctx := git.ContextWithGitWrapper(context.Background(), &notesStub{})
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
			ctx := git.ContextWithGitWrapper(context.Background(), gitWrapper)

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
				revParseHeadImplementation: dummyStubArgsNone,
				logCommitsImplementation:   dummyStubArgsNone,
				notesAddImplementation:     dummyStubArgsStringString,
				notesListImplementation:    dummyStubArgsString,
				notesShowImplementation:    dummyStubArgsStringString,
			}
			ctx := git.ContextWithGitWrapper(context.Background(), gitWrapper)

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
		// Mimic error when repository is exhausted without encountering a note
		err = errors.New("exit status 128")
		response = "fatal: couldn't find remote ref refs/notes/FOO"

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
		ctx := git.ContextWithGitWrapper(context.Background(), gitWrapper)

		args := []string{"list"}
		gotOutput, err := executeCommandContext(ctx, root, args...)

		assert.NoError(t, err)
		assert.Equal(t, testDataEmpty.outputPlain, gotOutput)
	})
}

func TestGetCommitHashes(t *testing.T) {
	testCases := []struct {
		name                string
		gitLogCommitsOutput string
		wantedGitCommits    []string
	}{
		{
			name:                "Get commit hashes - no history",
			gitLogCommitsOutput: "",
			wantedGitCommits:    []string{},
		},
		{
			name:                "Get commit hashes - 1 commit history",
			gitLogCommitsOutput: "COMMIT_HASH\n",
			wantedGitCommits:    []string{"COMMIT_HASH"},
		},
		{
			name:                "Get commit hashes - 2 commit history",
			gitLogCommitsOutput: "1234567\n890abcd\n",
			wantedGitCommits:    []string{"1234567", "890abcd"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gitWrapper := &notesStub{
				logCommitsImplementation: func() (response string, err error) {
					return tc.gitLogCommitsOutput, nil
				},
			}

			hashes, err := getCommitHashes(gitWrapper, 10000)

			assert.NoError(t, err)
			assert.EqualValues(t, tc.wantedGitCommits, hashes)
		})
	}
}

func TestGetNotesHashes(t *testing.T) {
	testCases := []struct {
		name               string
		gitNotesListOutput string
		wantedNotesCommits []string
	}{
		{
			name:               "Get notes hashes - no notes",
			gitNotesListOutput: "",
			wantedNotesCommits: []string{},
		},
		{
			name:               "Get notes hashes - 1 note",
			gitNotesListOutput: "NOTE_OBJECT_HASH ANNOTATED_OBJECT_HASH\n",
			wantedNotesCommits: []string{"ANNOTATED_OBJECT_HASH"},
		},
		{
			name:               "Get notes hashes - 1 note",
			gitNotesListOutput: "NOTE_OBJECT_HASH ANNOTATED_OBJECT_HASH\n01234567 890abcd\n",
			wantedNotesCommits: []string{"ANNOTATED_OBJECT_HASH", "890abcd"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gitWrapper := &notesStub{
				notesListImplementation: func(string) (response string, err error) {
					return tc.gitNotesListOutput, nil
				},
			}

			hashes, err := getNotesHashes(gitWrapper, "DUMMY_NOTES_REF")

			assert.NoError(t, err)
			assert.EqualValues(t, tc.wantedNotesCommits, hashes)
		})
	}
}
