package main

import (
	"context"
	"encoding/json"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/philips-software/gino-keva/internal/utils"
)

const (
	maxRetryAttempts = 3
)

type contextKey string

var (
	notesContextKey contextKey = "gitNotesKey"
)

// GitWrapper interface
type GitWrapper interface {
	FetchNotes(notesRef string, force bool) (string, error)
	LogCommits(maxCount uint) (string, error)
	NotesAdd(notesRef, msg string) (string, error)
	NotesList(notesRef string) (string, error)
	NotesPrune(notesRef string) (string, error)
	NotesShow(notesRef, hash string) (string, error)
	PushNotes(notesRef string) (string, error)
	RevParseHead() (string, error)
}

// ContextWithGitWrapper returns a new context with the git wrapper object added
func ContextWithGitWrapper(ctx context.Context, gitWrapper GitWrapper) context.Context {
	return context.WithValue(ctx, notesContextKey, gitWrapper)
}

// GetGitWrapperFrom returns the git wrapper object from the provided context
func GetGitWrapperFrom(ctx context.Context) GitWrapper {
	v := ctx.Value(notesContextKey)
	if v == nil {
		log.Fatal("No Notes interface found in context")
	}

	return v.(GitWrapper)
}

var globalFlags = struct {
	MaxDepth   uint
	NotesRef   string
	VerboseLog bool

	Fetch bool
}{}

func getNoteValues(gitWrapper GitWrapper, notesRef string, maxDepth uint) (values *Values, err error) {
	noteText, err := findNoteText(gitWrapper, notesRef, maxDepth)
	if err != nil {
		return nil, err
	}

	values, err = unmarshal(noteText)
	if err != nil {
		return nil, err
	}

	return values, err
}

func findNoteText(gitWrapper GitWrapper, notesRef string, maxDepth uint) (noteText string, err error) {
	notes, err := getNotesHashes(gitWrapper, notesRef)
	if err != nil {
		return "", err
	}
	log.WithFields(log.Fields{
		"first 10 notes":   utils.LimitStringSlice(notes, 10),
		"Total # of notes": len(notes),
	}).Debug()

	// Count is 1 higher than depth, since depth of 0 refers would still include current HEAD commit
	maxCount := maxDepth + 1

	// Try to get one more commit so we can detect if commits were exhausted in case no note was found
	commits, err := getCommitHashes(gitWrapper, maxCount+1)
	if err != nil {
		return "", err
	}
	log.WithFields(log.Fields{
		"first 10 commits":   utils.LimitStringSlice(commits, 10),
		"Total # of commits": len(commits),
	}).Debug()

	// Get all notes for commits up to maxDepth
	notesIntersect := utils.GetSlicesIntersect(notes, utils.LimitStringSlice(commits, maxCount))
	log.WithFields(log.Fields{
		"first 10 notesIntersect":   utils.LimitStringSlice(notesIntersect, 10),
		"Total # of notesIntersect": len(notesIntersect),
	}).Debug()

	if len(notesIntersect) == 0 {
		if len(commits) == int(maxCount+1) {
			log.WithField("ref", notesRef).Warning("No prior notes found within maximum depth!")
		} else {
			log.WithField("ref", notesRef).Warning("Reached root commit. No prior notes found")
		}
		noteText = ""
	} else {
		noteText, err = gitWrapper.NotesShow(notesRef, notesIntersect[0])
		if err != nil {
			return "", err
		}
	}

	return noteText, nil
}

func unmarshal(rawText string) (*Values, error) {
	v := make(map[string]Value)

	if rawText != "" {
		err := json.Unmarshal([]byte(rawText), &v)
		if err != nil {
			return nil, err
		}
	}

	return &Values{values: v}, nil
}

func fetchNotes(gitWrapper GitWrapper) (err error) {
	err = fetchNotesWithForce(gitWrapper, globalFlags.NotesRef, false)

	if _, ok := err.(*UpstreamChanged); ok {
		log.Warning("Unpushed local changes are now discarded")
		err = fetchNotesWithForce(gitWrapper, globalFlags.NotesRef, true)
	}

	if _, ok := err.(*NoRemoteRef); ok {
		log.WithField("notesRef", globalFlags.NotesRef).Debug("Couldn't find remote ref. Nothing fetched")
		err = nil
	}

	if err != nil {
		log.Error(err.Error())
	}

	return err
}

func fetchNotesWithForce(gitWrapper GitWrapper, notesRef string, force bool) error {
	log.WithField("force", force).Debug("Fetching notes...")
	defer log.Debug("Done.")

	out, errorCode := gitWrapper.FetchNotes(globalFlags.NotesRef, force)
	return convertGitOutputToError(out, errorCode)
}

func pruneNotes(gitWrapper GitWrapper, notesRef string) error {
	log.Debug("Pruning notes...")
	defer log.Debug("Done.")

	out, errorCode := gitWrapper.NotesPrune(globalFlags.NotesRef)
	return convertGitOutputToError(out, errorCode)
}

func pushNotes(gitWrapper GitWrapper, notesRef string) error {
	log.Debug("Pushing notes...")
	defer log.Debug("Done.")

	out, errorCode := gitWrapper.PushNotes(globalFlags.NotesRef)
	err := convertGitOutputToError(out, errorCode)

	if _, ok := err.(*UpstreamChanged); ok {
		return err
	}

	if err != nil {
		log.Error(out)
	}

	return err
}

var getCommitHashes = func(gitWrapper GitWrapper, maxCount uint) (hashList []string, err error) {
	output, err := gitWrapper.LogCommits(maxCount)
	if err != nil {
		return nil, err
	}

	if output == "" {
		hashList = []string{}
	} else {
		output := strings.TrimSuffix(output, "\n")
		hashList = strings.Split(output, "\n")
	}

	return hashList, nil
}

var getNotesHashes = func(gitWrapper GitWrapper, notesRef string) (hashList []string, err error) {
	output, err := gitWrapper.NotesList(notesRef)
	if err != nil {
		return nil, err
	}

	if output == "" {
		hashList = []string{}
	} else {
		output := strings.TrimSuffix(output, "\n")
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			hashes := strings.Split(line, " ")
			hashList = append(hashList, hashes[1])
		}
	}
	return hashList, nil
}
