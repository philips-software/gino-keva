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

// Wrapper interface
type Wrapper interface {
	AddNote(notesRef, msg string) (string, error)
	FetchNotes(notesRef string, force bool) (string, error)
	PushNotes(notesRef string) (string, error)
	RevParseHead() (string, error)
	ShowNote(notesRef, hash string) (string, error)
}

// GoGitCmdWrapper implements the Wrapper interface using go-git-cmd-wrapper
type GoGitCmdWrapper struct {
}

// AddNote sets/overwrites a note
func (GoGitCmdWrapper) AddNote(notesRef, msg string) (string, error) {
	return gitCmdWrapper.Notes(notes.Ref(notesRef), notes.Add("", notes.Message(msg), notes.Force))
}

// FetchNotes notes
func (GoGitCmdWrapper) FetchNotes(notesRef string, force bool) (string, error) {
	refSpec := fmt.Sprintf("refs/notes/%v:refs/notes/%v", notesRef, notesRef)
	if force {
		// Add + to force fetch
		refSpec = fmt.Sprintf("+%v", refSpec)
	}
	return gitCmdWrapper.Fetch(fetch.NoTags, fetch.Remote("origin"), fetch.RefSpec(refSpec))
}

// PushNotes notes
func (GoGitCmdWrapper) PushNotes(notesRef string) (string, error) {
	refSpec := fmt.Sprintf("refs/notes/%v:refs/notes/%v", notesRef, notesRef)
	return gitCmdWrapper.Push(push.Remote("origin"), push.RefSpec(refSpec))
}

// RevParseHead returns the HEAD commit hash
func (g GoGitCmdWrapper) RevParseHead() (string, error) {
	return gitCmdWrapper.RevParse(revparse.Args("HEAD"))
}

// ShowNote returns the note for provided hash, or error if there is none
func (GoGitCmdWrapper) ShowNote(notesRef, hash string) (string, error) {
	return gitCmdWrapper.Notes(notes.Ref(notesRef), notes.Show(hash))
}

type contextKey string

var (
	notesContextKey contextKey = "gitNotesKey"
)

// ContextWithGitWrapper returns a new context with the git wrapper object added
func ContextWithGitWrapper(ctx context.Context, gitWrapper Wrapper) context.Context {
	return context.WithValue(ctx, notesContextKey, gitWrapper)
}

// GetGitWrapperFrom returns the git wrapper object from the provided context
func GetGitWrapperFrom(ctx context.Context) Wrapper {
	v := ctx.Value(notesContextKey)
	if v == nil {
		log.Fatal("No Notes interface found in context")
	}

	return v.(Wrapper)
}
