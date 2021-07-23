package main

import (
	"testing"

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

			hashes, err := getNotesHashes(gitWrapper, dummyRef)

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
			getCommitHashesOutput: []string{"0", "1", "2", "3"},
			expectedNotes:         []string{},
		},
		{
			name:                 "Some notes, no commits",
			getNotesHashesOutput: []string{"0", "1", "2", "3"},
			expectedNotes:        []string{},
		},
		{
			name:                  "Note at depth 1",
			getCommitHashesOutput: []string{"0", "1"},
			getNotesHashesOutput:  []string{"1"},
			expectedNotes:         []string{"1"},
		},
		{
			name:                  "Note at depth 10",
			getCommitHashesOutput: []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			getNotesHashesOutput:  []string{"10"},
			expectedNotes:         []string{"10"},
		},
		{
			name:                  "Note at depth 0 and 1",
			getCommitHashesOutput: []string{"0", "1"},
			getNotesHashesOutput:  []string{"0", "1"},
			expectedNotes:         []string{"0", "1"},
		},
		{
			name:                  "Note at depth 0 and 1 (notes output reversed)",
			getCommitHashesOutput: []string{"0", "1"},
			getNotesHashesOutput:  []string{"1", "0"},
			expectedNotes:         []string{"0", "1"},
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
			notes, err := getRelevantNotes(gitWrapper, dummyRef)

			assert.NoError(t, err)
			assert.EqualValues(t, tc.expectedNotes, notes)
		})
	}
}
