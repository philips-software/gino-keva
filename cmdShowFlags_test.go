package main

import (
	"context"
	"testing"

	"github.com/philips-internal/gino-keva/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestShowFlags(t *testing.T) {
	t.Run("Error when trying to show unknown flag", func(t *testing.T) {
		root := NewRootCommand()
		args := []string{"show-flag", "Unknown_Flag"}
		ctx := git.ContextWithNotes(context.Background(), &notesDummy{})

		_, err := executeCommandContext(ctx, root, args...)

		assert.Error(t, err)
	})
}
