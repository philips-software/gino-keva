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

type notesDummy struct {
}

// AddNote test-double
func (n notesDummy) AddNote(string, string) (string, error) {
	panic(errors.New("unexpected call to dummy method"))
}

// FetchNotes test-double
func (n notesDummy) FetchNotes(string, bool) (string, error) {
	return "", nil
}

// GetCommitHashes test-double
func (n notesDummy) GetCommitHashes() (string, error) {
	return "", nil
}

// PushNotes test-double
func (n notesDummy) PushNotes(string) (string, error) {
	return "", nil
}

// RevParseHead test-double
func (n notesDummy) RevParseHead() (string, error) {
	panic(errors.New("unexpected call to dummy method"))
}

//ShowNote test-double
func (n *notesDummy) ShowNote(string, string) (response string, err error) {
	panic(errors.New("unexpected call to dummy method"))
}

type notesAddSpy struct {
	AddResult            string
	revParseHeadResponse string
	showResponse         string
}

// AddNote test-double
func (n *notesAddSpy) AddNote(_ string, msg string) (string, error) {
	n.AddResult = msg // Store input to Add function for test inspection
	return "", nil
}

// FetchNotes test-double
func (n notesAddSpy) FetchNotes(string, bool) (string, error) {
	return "", nil
}

// GetCommitHashes test-double
func (n notesAddSpy) GetCommitHashes() (string, error) {
	return "", nil
}

// PushNotes test-double
func (n notesAddSpy) PushNotes(string) (string, error) {
	return "", nil
}

// RevParse test-double
func (n notesAddSpy) RevParseHead() (string, error) {
	return n.revParseHeadResponse, nil
}

//ShowNote test-double
func (n *notesAddSpy) ShowNote(string, string) (string, error) {
	return n.showResponse, nil
}

type notesStub struct {
	addNoteImplementation         func(string, string) (string, error)
	fetchNotesImplementation      func(string) (string, error)
	getCommitHashesImplementation func() (string, error)
	pushNotesImplementation       func(string) (string, error)
	revParseHeadImplementation    func() (string, error)
	showNoteImplementation        func(string, string) (string, error)
}

// AddNote test-double
func (n notesStub) AddNote(notesRef string, msg string) (string, error) {
	return n.addNoteImplementation(notesRef, msg)
}

// FetchNotes test-double
func (n notesStub) FetchNotes(notesRef string, force bool) (string, error) {
	return n.fetchNotesImplementation(notesRef)
}

// GetCommitHashes test-double
func (n notesStub) GetCommitHashes() (string, error) {
	return n.getCommitHashesImplementation()
}

// PushNotes test-double
func (n notesStub) PushNotes(notesRef string) (string, error) {
	return n.pushNotesImplementation(notesRef)
}

// RevParseHead test-double
func (n notesStub) RevParseHead() (string, error) {
	return n.revParseHeadImplementation()
}

//ShowNote test-double calls the stub implementation
func (n *notesStub) ShowNote(notesRef, hash string) (response string, err error) {
	return n.showNoteImplementation(notesRef, hash)
}

var dummyStubInputsNone = func() (string, error) { return "", nil }
var dummyStubInputsString = func(string) (string, error) { return "", nil }
var dummyStubInputsStringString = func(string, string) (string, error) { return "", nil }

var spyInputsString = func(isCalled *bool) func(string) (string, error) {
	*isCalled = false
	return func(string) (string, error) {
		*isCalled = true
		return "", nil
	}
}

var showStubReturnResponseAtDepth = func(expectedResponse string, depth int) func(string, string) (response string, err error) {
	return func(string, string) (response string, err error) {
		switch {
		case depth < 0:
			err = errors.New("search continued too deep")
		case depth == 0:
			response = expectedResponse // Note found
		default:
			// No note at this level; mimic expected error and response
			err = errors.New("exit status 1")
			response = "error: no note found for object 0123456789abcdef.\n"
		}

		depth--

		return response, err
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

			ctx := git.ContextWithGitWrapper(context.Background(), &notesDummy{})
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
			name:            "Fetch is called when flag is set (default)",
			args:            []string{"show-flag", "ref"},
			wantFetchCalled: true,
		},
		{
			name:            "Fetch is NOT called when flag is unset",
			args:            []string{"show-flag", "ref", "--fetch=false"},
			wantFetchCalled: false,
		}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var fetchCalled bool
			gitWrapper := &notesStub{
				fetchNotesImplementation: spyInputsString(&fetchCalled),
				pushNotesImplementation:  dummyStubInputsString,
				showNoteImplementation:   dummyStubInputsStringString,
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
			args:           []string{"set", "foo", "bar", "--push"},
			wantPushCalled: true,
		}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var pushCalled bool
			gitWrapper := &notesStub{
				addNoteImplementation:      dummyStubInputsStringString,
				fetchNotesImplementation:   dummyStubInputsString,
				pushNotesImplementation:    spyInputsString(&pushCalled),
				revParseHeadImplementation: dummyStubInputsNone,
				showNoteImplementation:     dummyStubInputsStringString,
			}
			ctx := git.ContextWithGitWrapper(context.Background(), gitWrapper)

			root := NewRootCommand()
			_, err := executeCommandContext(ctx, root, tc.args...)

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
			pushNotesImplementation:  dummyStubInputsString,
			showNoteImplementation:   dummyStubInputsStringString,
		}
		ctx := git.ContextWithGitWrapper(context.Background(), gitWrapper)

		args := []string{"list"}
		gotOutput, err := executeCommandContext(ctx, root, args...)

		assert.NoError(t, err)
		assert.Equal(t, testDataEmpty.outputPlain, gotOutput)
	})
}
