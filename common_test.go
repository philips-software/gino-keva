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

			hashes, err := getNotesHashes(gitWrapper, "DUMMY_NOTES_REF")

			assert.NoError(t, err)
			assert.EqualValues(t, tc.wantedNotesCommits, hashes)
		})
	}
}
