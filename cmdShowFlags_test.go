package main

import (
	"context"
	"testing"

	"github.com/philips-software/gino-keva/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestShowFlags(t *testing.T) {
	t.Run("Error when trying to show unknown flag", func(t *testing.T) {
		root := NewRootCommand()
		args := []string{"show-flag", "Unknown_Flag"}

		ctx := git.ContextWithGitWrapper(context.Background(), &notesStub{})

		_, err := executeCommandContext(ctx, root, args...)

		assert.Error(t, err)
	})
	t.Run("Able to retrieve integer flags", func(t *testing.T) {
		root := NewRootCommand()
		args := []string{"show-flag", "max-depth", "--max-depth", "42"}
		ctx := git.ContextWithGitWrapper(context.Background(), &notesStub{})

		output, err := executeCommandContext(ctx, root, args...)

		assert.NoError(t, err)
		assert.Equal(t, "42", output)
	})
}
