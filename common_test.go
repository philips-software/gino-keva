package main

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/philips-software/gino-keva/internal/event"
	"github.com/stretchr/testify/assert"
)

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

			hashes, err := getCommitHashes(gitWrapper)

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

			hashes, err := getNotesHashes(gitWrapper, TestDataDummyRef)

			assert.NoError(t, err)
			assert.EqualValues(t, tc.wantedNotesCommits, hashes)
		})
	}
}

func TestFindNoteText(t *testing.T) {
	testCases := []struct {
		name                  string
		getCommitHashesOutput []string
		getNotesHashesOutput  []string
		expectedNotes         []string
	}{
		{
			name:          "No commits, no notes",
			expectedNotes: []string{},
		},
		{
			name:                  "Some commits, no notes",
			getCommitHashesOutput: generateIncrementingNumbersListOfLength(4),
			expectedNotes:         []string{},
		},
		{
			name:                 "Some notes, no commits",
			getNotesHashesOutput: generateIncrementingNumbersListOfLength(4),
			expectedNotes:        []string{},
		},
		{
			name:                  "Note at depth 1",
			getCommitHashesOutput: generateIncrementingNumbersListOfLength(2),
			getNotesHashesOutput:  []string{"1"},
			expectedNotes:         []string{"1"},
		},
		{
			name:                  "Note at depth 10",
			getCommitHashesOutput: generateIncrementingNumbersListOfLength(11),
			getNotesHashesOutput:  []string{"10"},
			expectedNotes:         []string{"10"},
		},
		{
			name:                  "Note at depth 0 and 1",
			getCommitHashesOutput: generateIncrementingNumbersListOfLength(2),
			getNotesHashesOutput:  generateIncrementingNumbersListOfLength(2),
			expectedNotes:         generateIncrementingNumbersListOfLength(2),
		},
		{
			name:                  "Note at depth 0 and 1 (notes output reversed)",
			getCommitHashesOutput: generateIncrementingNumbersListOfLength(2),
			getNotesHashesOutput:  []string{"1", "0"},
			expectedNotes:         generateIncrementingNumbersListOfLength(2),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getCommitHashes = func(GitWrapper) ([]string, error) {
				return tc.getCommitHashesOutput, nil
			}

			getNotesHashes = func(GitWrapper, string) ([]string, error) {
				return tc.getNotesHashesOutput, nil
			}

			var notesShowCalled bool
			var hashArg string
			gitWrapper := &notesStub{
				notesShowImplementation: spyArgsStringString(&notesShowCalled, nil, &hashArg),
			}
			notes, err := getRelevantNotes(gitWrapper, TestDataDummyRef)

			assert.NoError(t, err)
			assert.EqualValues(t, tc.expectedNotes, notes)
		})
	}
}

func TestCalculateKeyValues(t *testing.T) {
	testCases := []struct {
		name   string
		events [][]event.Event
		wanted map[string]Value
	}{
		{
			name:   "No notes",
			events: [][]event.Event{{}},
			wanted: map[string]Value{},
		},
		{
			name: "Note with single set event",
			events: [][]event.Event{
				{event.TestDataSetFooBar},
			},
			wanted: map[string]Value{
				event.TestDataFoo: Value(event.TestDataBar),
			},
		},
		{
			name: "Note with 2 set event",
			events: [][]event.Event{
				{event.TestDataSetFooBar, event.TestDataSetKeyValue},
			},
			wanted: map[string]Value{
				event.TestDataFoo: Value(event.TestDataBar),
				event.TestDataKey: Value(event.TestDataValue),
			},
		},
		{
			name: "Overwrite value in same note",
			events: [][]event.Event{
				{event.TestDataSetKeyOtherValue, event.TestDataSetKeyValue},
			},
			wanted: map[string]Value{
				event.TestDataKey: Value(event.TestDataOtherValue),
			},
		},
		{
			name: "Set and unset value in same note",
			events: [][]event.Event{
				{event.TestDataUnsetKey, event.TestDataSetKeyValue},
			},
			wanted: map[string]Value{},
		},
		{
			name: "Set events across 2 notes",
			events: [][]event.Event{
				{event.TestDataSetFooBar},
				{event.TestDataSetKeyValue},
			},
			wanted: map[string]Value{
				event.TestDataFoo: Value(event.TestDataBar),
				event.TestDataKey: Value(event.TestDataValue),
			},
		},
		{
			name: "Overwrite value across 2 notes",
			events: [][]event.Event{
				{event.TestDataSetKeyOtherValue},
				{event.TestDataSetKeyValue},
			},
			wanted: map[string]Value{
				event.TestDataKey: Value(event.TestDataOtherValue),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getCommitHashes = func(GitWrapper) ([]string, error) {
				return generateIncrementingNumbersListOfLength(len(tc.events)), nil
			}

			getNotesHashes = func(GitWrapper, string) ([]string, error) {
				return generateIncrementingNumbersListOfLength(len(tc.events)), nil
			}

			gitWrapper := &notesStub{
				logCommitsImplementation: dummyStubArgsNone,
				notesListImplementation:  dummyStubArgsString,
				notesShowImplementation: func(string, hash string) (string, error) {
					i, _ := strconv.Atoi(hash)
					return event.Marshal(&tc.events[i])
				},
			}

			got, err := calculateKeyValues(gitWrapper, TestDataDummyRef)

			assert.NoError(t, err)
			assert.Truef(t, reflect.DeepEqual(got.values, tc.wanted), "Got %v, wanted %v", got.values, tc.wanted)
		})
	}
}
