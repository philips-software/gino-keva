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

// Values represents a collection of values
type Values struct {
	values map[string]Value
}

// Add a key/value to the collection
func (v *Values) Add(key string, value Value) {
	v.values[key] = value
}

// Count returns number of items in collection
func (v *Values) Count() int {
	return len(v.values)
}

// GetJSON returns the value data in json format
func (v Values) GetJSON(key string) string {
	return v.values[key].Data
}

// Iterate the collection values
func (v *Values) Iterate() map[string]Value {
	return v.values
}

// Remove a key from the collection
func (v *Values) Remove(key string) {
	delete(v.values, key)
}

// Value represents the parsed value as stored in git notes
type Value struct {
	Data   string `json:"data"`
	Source string `json:"source"`
}

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

		ctx := git.ContextWithNotes(context.Background(), &git.GoGitCmdWrapper{})

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
