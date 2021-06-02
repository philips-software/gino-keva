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
			args:       []string{"get", "key"},
			wantOutput: "value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRootCommand()
			ctx := git.ContextWithNotes(context.Background(), &notesStub{
				addImplementation:          panicStubInputsStringString,
				fetchImplementation:        dummyStubInputsString,
				pushImplementation:         dummyStubInputsString,
				revParseHeadImplementation: panicStubInputsNone,
				showImplementation:         showStubReturnResponseAtDepth(input, 0),
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
			key:       "key",
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
			notesAccess := notesStub{
				addImplementation:          panicStubInputsStringString,
				fetchImplementation:        dummyStubInputsString,
				pushImplementation:         dummyStubInputsString,
				revParseHeadImplementation: panicStubInputsNone,
				showImplementation:         showStubReturnResponseAtDepth(input, tc.depth),
			}
			gotValue, err := getValue(&notesAccess, "dummyRef", tc.key, maxDepth)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantValue, gotValue)
		})
	}
}
