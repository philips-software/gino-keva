package main

import (
	"context"
	"testing"

	"github.com/philips-software/gino-keva/internal/git"
	"github.com/stretchr/testify/assert"
)

var dummyVersionInfo = "version.Info{Version: \"<Version not set>\", BuildDate: \"<BuildDate not set>\", GitCommit: \"<GitCommit not set>\", GitState: \"<GitState not set>\"}\n"

func TestVersion(t *testing.T) {
	t.Run("Version command returns an empty version string", func(t *testing.T) {
		root := NewRootCommand()
		args := []string{"version"}
		ctx := git.ContextWithGitWrapper(context.Background(), &notesStub{})

		out, err := executeCommandContext(ctx, root, args...)

		assert.NoError(t, err)
		assert.Equal(t, dummyVersionInfo, out)
	})
}
