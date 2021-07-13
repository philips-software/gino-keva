package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCommand(t *testing.T) {

	testCases := []struct {
		name       string
		args       []string
		input      string
		wantOutput string
	}{
		{
			name:       "Get value of a key",
			args:       []string{"get", "MY_KEY"},
			input:      testDataKeyValue.inputNew,
			wantOutput: "value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCommand()
			ctx := ContextWithGitWrapper(context.Background(), &notesStub{
				logCommitsImplementation: responseStubArgsNone(simpleLogCommitsResponse),
				notesListImplementation:  responseStubArgsString(simpleNotesListResponse),
				notesShowImplementation:  responseStubArgsStringString(tc.input),
			})
			args := disableFetch(tc.args)
			gotOutput, err := executeCommandContext(ctx, root, args...)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantOutput, gotOutput)
		})
	}
}

func TestGetValue(t *testing.T) {
	maxDepth := uint(5)

	testCases := []struct {
		name      string
		key       string
		depth     uint
		input     string
		wantValue string
	}{
		{
			name:      "Get value of an existing key (old format)",
			key:       "MY_KEY",
			depth:     0,
			input:     testDataKeyValue.inputOld,
			wantValue: "value",
		},
		{
			name:      "Get value of a non-existing key",
			key:       "nonExistingKey",
			depth:     0,
			input:     testDataKeyValue.inputNew,
			wantValue: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gitWrapper := notesStub{
				logCommitsImplementation: responseStubArgsNone(simpleLogCommitsResponse),
				notesListImplementation:  responseStubArgsString(simpleNotesListResponse),
				notesShowImplementation:  responseStubArgsStringString(tc.input),
			}
			gotValue, err := getValue(&gitWrapper, dummyRef, tc.key, maxDepth)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantValue, gotValue)
		})
	}
}
