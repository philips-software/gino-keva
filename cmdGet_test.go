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
			input:      testDataKeyValue.input,
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
	testCases := []struct {
		name      string
		key       string
		depth     uint
		input     string
		wantValue string
	}{
		{
			name:      "Get value of an existing key",
			key:       "MY_KEY",
			depth:     0,
			input:     testDataKeyValue.input,
			wantValue: "value",
		},
		{
			name:      "Get value of a non-existing key",
			key:       "nonExistingKey",
			depth:     0,
			input:     testDataKeyValue.input,
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
			gotValue, err := getValue(&gitWrapper, dummyRef, tc.key)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantValue, gotValue)
		})
	}
}
