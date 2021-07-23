package git

import (
	"fmt"

	"github.com/ldez/go-git-cmd-wrapper/v2/fetch"
	gitCmdWrapper "github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/notes"
	"github.com/ldez/go-git-cmd-wrapper/v2/push"
	"github.com/ldez/go-git-cmd-wrapper/v2/revparse"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
)

// GoGitCmdWrapper implements the Wrapper interface using go-git-cmd-wrapper
type GoGitCmdWrapper struct {
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

// LogCommits returns log output with commit hashes
func (GoGitCmdWrapper) LogCommits() (string, error) {
	return gitCmdWrapper.Raw("log", func(g *types.Cmd) {
		g.AddOptions("--pretty=format:%H")
	})
}

// NotesAdd sets/overwrites a note
func (GoGitCmdWrapper) NotesAdd(notesRef, msg string) (string, error) {
	return gitCmdWrapper.Notes(notes.Ref(notesRef), notes.Add("", notes.Message(msg), notes.Force))
}

// NotesList returns all the notes
func (GoGitCmdWrapper) NotesList(notesRef string) (string, error) {
	return gitCmdWrapper.Notes(notes.Ref(notesRef), notes.List(""))
}

// NotesPrune prunes unreachable notes
func (GoGitCmdWrapper) NotesPrune(notesRef string) (string, error) {
	return gitCmdWrapper.Notes(notes.Ref(notesRef), notes.Prune())
}

// NotesShow returns the note for provided hash, or error if there is none
func (GoGitCmdWrapper) NotesShow(notesRef, hash string) (string, error) {
	return gitCmdWrapper.Notes(notes.Ref(notesRef), notes.Show(hash))
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
