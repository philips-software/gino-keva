package main

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/philips-software/gino-keva/internal/git"
)

const (
	envPrefix = "GINO_KEVA"
)

func checkIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.SetOutput(os.Stderr)

	attemptsLeft := maxRetryAttempts

	var err error
	for {
		root := NewRootCommand()
		root.SilenceUsage = true
		root.SilenceErrors = true

		ctx := ContextWithGitWrapper(context.Background(), &git.GoGitCmdWrapper{})

		err = root.ExecuteContext(ctx)
		attemptsLeft--
		uc, upstreamChanged := err.(*UpstreamChanged)

		if attemptsLeft > 0 && upstreamChanged && uc.fetchEnabled {
			log.WithField("attemptsLeft", attemptsLeft).Info("Upstream has changed in the meanwhile. Starting again from fetch")
		} else {
			break
		}
	}

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
