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

			hashes, err := getNotesHashes(gitWrapper, dummyRef)

			assert.NoError(t, err)
			assert.EqualValues(t, tc.wantedNotesCommits, hashes)
		})
	}
}

func TestFindNoteText(t *testing.T) {
	testCases := []struct {
		name                    string
		getCommitHashesOutput   []string
		getNotesHashesOutput    []string
		maxDepth                uint
		expectedNotesShowCalled bool
		expectedHashArg         string
	}{
		{
			name:                    "No commits, no notes",
			maxDepth:                1000,
			expectedNotesShowCalled: false,
		},
		{
			name:                    "Some commits, no notes",
			getCommitHashesOutput:   []string{"0", "1", "2", "3"},
			maxDepth:                1000,
			expectedNotesShowCalled: false,
		},
		{
			name:                    "Some notes, no commits",
			getNotesHashesOutput:    []string{"0", "1", "2", "3"},
			maxDepth:                1000,
			expectedNotesShowCalled: false,
		},
		{
			name:                    "One commit with note at depth 0 (HEAD)",
			getCommitHashesOutput:   []string{"0"},
			getNotesHashesOutput:    []string{"0"},
			maxDepth:                0,
			expectedNotesShowCalled: true,
			expectedHashArg:         "0",
		},
		{
			name:                    "Note at depth 1, maxDepth=0",
			getCommitHashesOutput:   []string{"0", "1"},
			getNotesHashesOutput:    []string{"1"},
			maxDepth:                0,
			expectedNotesShowCalled: false,
		},
		{
			name:                    "Note at depth 1, maxDepth=1",
			getCommitHashesOutput:   []string{"0", "1"},
			getNotesHashesOutput:    []string{"1"},
			maxDepth:                1,
			expectedNotesShowCalled: true,
			expectedHashArg:         "1",
		},
		{
			name:                    "Note at depth 10, maxDepth=10",
			getCommitHashesOutput:   []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			getNotesHashesOutput:    []string{"10"},
			maxDepth:                10,
			expectedNotesShowCalled: true,
			expectedHashArg:         "10",
		},
		{
			name:                    "Note at depth 0 and 1",
			getCommitHashesOutput:   []string{"0", "1"},
			getNotesHashesOutput:    []string{"0", "1"},
			maxDepth:                1000,
			expectedNotesShowCalled: true,
			expectedHashArg:         "0",
		},
		{
			name:                    "Note at depth 0 and 1 (notes output reversed)",
			getCommitHashesOutput:   []string{"0", "1"},
			getNotesHashesOutput:    []string{"1", "0"},
			maxDepth:                1000,
			expectedNotesShowCalled: true,
			expectedHashArg:         "0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getCommitHashes = func(GitWrapper, uint) ([]string, error) {
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
			_, err := findNoteText(gitWrapper, dummyRef, tc.maxDepth)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedNotesShowCalled, notesShowCalled)
			if notesShowCalled {
				assert.Equal(t, tc.expectedHashArg, hashArg)
			}
		})
	}
}
