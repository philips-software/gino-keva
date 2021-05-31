package git

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/ldez/go-git-cmd-wrapper/v2/fetch"
	gitCmdWrapper "github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/notes"
	"github.com/ldez/go-git-cmd-wrapper/v2/push"
	"github.com/ldez/go-git-cmd-wrapper/v2/revparse"
)

// Notes interface
type Notes interface {
	Add(notesRef string, msg string) (string, error)
	Fetch(notesRef string, force bool) (string, error)
	Push(notesRef string) (string, error)
	RevParseHead() (string, error)
	Show(notesRef string, hash string) (string, error)
}

// GoGitCmdWrapper implements the Notes interface using go-git-cmd-wrapper
type GoGitCmdWrapper struct {
}

// Add sets/overwrites a note
func (g GoGitCmdWrapper) Add(notesRef string, msg string) (string, error) {
	return gitCmdWrapper.Notes(notes.Ref(notesRef), notes.Add("", notes.Message(msg), notes.Force))
}

// Fetch notes
func (g GoGitCmdWrapper) Fetch(notesRef string, force bool) (string, error) {
	refSpec := fmt.Sprintf("refs/notes/%v:refs/notes/%v", notesRef, notesRef)
	if force {
		// Add + to force fetch
		refSpec = fmt.Sprintf("+%v", refSpec)
	}
	return gitCmdWrapper.Fetch(fetch.NoTags, fetch.Remote("origin"), fetch.RefSpec(refSpec))
}

// Push notes
func (g GoGitCmdWrapper) Push(notesRef string) (string, error) {
	refSpec := fmt.Sprintf("refs/notes/%v:refs/notes/%v", notesRef, notesRef)
	return gitCmdWrapper.Push(push.Remote("origin"), push.RefSpec(refSpec))
}

// RevParseHead returns the HEAD commit hash
func (g GoGitCmdWrapper) RevParseHead() (string, error) {
	return gitCmdWrapper.RevParse(revparse.Args("HEAD"))
}

// Show returns the note for provided hash, or error if there is none
func (g GoGitCmdWrapper) Show(notesRef string, hash string) (string, error) {
	return gitCmdWrapper.Notes(notes.Ref(notesRef), notes.Show(hash))
}

type contextKey string

var (
	notesContextKey contextKey = "gitNotesKey"
)

// ContextWithNotes returns a new context with the notes object added
func ContextWithNotes(ctx context.Context, notes Notes) context.Context {
	return context.WithValue(ctx, notesContextKey, notes)
}

// GetNotesAccessFrom returns the notes object from the provided context
func GetNotesAccessFrom(ctx context.Context) Notes {
	v := ctx.Value(notesContextKey)
	if v == nil {
		log.Fatal("No Notes interface found in context")
	}

	return v.(Notes)
}
