package main

import (
	"bytes"
	"context"

	"github.com/spf13/cobra"
)

const (
	dummyRef   = "DUMMY_REF"
	dummyValue = "DUMMY_VALUE"
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

// Simple dummy responses for logCommits and notesList
var (
	simpleLogCommitsResponse = "COMMIT_REFERENCE\n"
	simpleNotesListResponse  = "NOTES_OBJECT_ID COMMIT_REFERENCE\n"
)

type testData struct {
	inputOld    string
	inputNew    string
	outputPlain string
	outputJSON  string
	outputRaw   string
}

var testDataEmpty = testData{
	inputOld:    `{}`,
	inputNew:    `{"snapshot":{}}`,
	outputPlain: "",
	outputJSON:  "{}\n",
	outputRaw:   "{\"snapshot\":{}}\n",
}

var testDataKeyValue = testData{
	inputOld:    `{"MY_KEY": {"data":"value", "source": "01234567"}}`,
	inputNew:    `{"snapshot":{"MY_KEY": {"data":"value", "source": "01234567"}}}`,
	outputPlain: "MY_KEY=value\n",
	outputJSON:  "{\n  \"MY_KEY\": \"value\"\n}\n",
	outputRaw:   "{\"snapshot\":{\"MY_KEY\":{\"data\":\"value\",\"source\":\"01234567\"}}}\n",
}

var testDataKeyValueFooBar = testData{
	inputOld:    `{"FOO": {"data":"bar", "source": "abcd1234"},"MY_KEY": {"data":"value", "source": "01234567"}}`,
	inputNew:    `{"snapshot":{"FOO": {"data":"bar", "source": "abcd1234"},"MY_KEY": {"data":"value", "source": "01234567"}}}`,
	outputPlain: "MY_KEY=value\nFOO=bar\n",
	outputJSON:  "{\n  \"FOO\": \"bar\",\n  \"MY_KEY\": \"value\"\n  }\n}",
	outputRaw:   "{\"snapshot\":{\"FOO\":{\"data\":\"bar\",\"source\":\"abcd1234\"},\"MY_KEY\":{\"data\":\"value\",\"source\":\"01234567\"}}}\n",
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
