package main

import (
	"context"
	"testing"

	"github.com/philips-software/gino-keva/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestGetCommand(t *testing.T) {
	input := testDataKeyValue.input

	testCases := []struct {
		name       string
		args       []string
		input      string
		wantOutput string
	}{
		{
			name:       "Get value of a key",
			args:       []string{"get", "MY_KEY"},
			wantOutput: "value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCommand()
			ctx := git.ContextWithGitWrapper(context.Background(), &notesStub{
				fetchNotesImplementation: dummyStubInputsString,
				pushNotesImplementation:  dummyStubInputsString,
				showNoteImplementation:   showStubReturnResponseAtDepth(input, 0),
			})
			gotOutput, err := executeCommandContext(ctx, root, tc.args...)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantOutput, gotOutput)
		})
	}
}

func TestGetValue(t *testing.T) {
	input := testDataKeyValue.input
	maxDepth := 5

	testCases := []struct {
		name      string
		key       string
		depth     int
		wantValue string
	}{
		{
			name:      "Get value of an existing key",
			key:       "MY_KEY",
			depth:     0,
			wantValue: "value",
		},
		{
			name:      "Get value of a non-existing key",
			key:       "nonExistingKey",
			depth:     0,
			wantValue: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gitWrapper := notesStub{
				fetchNotesImplementation: dummyStubInputsString,
				pushNotesImplementation:  dummyStubInputsString,
				showNoteImplementation:   showStubReturnResponseAtDepth(input, tc.depth),
			}
			gotValue, err := getValue(&gitWrapper, "dummyRef", tc.key, maxDepth)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantValue, gotValue)
		})
	}
}
