package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowFlags(t *testing.T) {
	t.Run("Error when trying to show unknown flag", func(t *testing.T) {
		root := NewRootCommand()
		args := []string{"show-flag", "Unknown_Flag"}

		ctx := ContextWithGitWrapper(context.Background(), &notesStub{})

		_, err := executeCommandContext(ctx, root, args...)

		assert.Error(t, err)
	})
}
