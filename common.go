package main

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/philips-software/gino-keva/internal/event"
	"github.com/philips-software/gino-keva/internal/util"
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
	LogCommits() (string, error)
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
	NotesRef   string
	VerboseLog bool

	Fetch bool
}{}

func getEvents(gitWrapper GitWrapper, notesRef string) (*[]event.Event, error) {
	var commitHash string
	{
		out, err := gitWrapper.RevParseHead()
		if err != nil {
			return nil, convertGitOutputToError(out, err)
		}
		commitHash = strings.TrimSuffix(out, "\n")
	}

	log.WithField("hash", commitHash).Debug("Retrieving events from git note...")
	events, err := getEventsFromNote(gitWrapper, notesRef, commitHash)

	if _, ok := err.(*NoNotePresent); ok {
		log.WithField("notesRef", globalFlags.NotesRef).Debug("No git note present yet")
		err = nil
	}

	return &events, err
}

func persistEvents(gitWrapper GitWrapper, notesRef string, events *[]event.Event) error {
	noteText, err := event.Marshal(events)
	if err != nil {
		log.Fatal(err)
	}
	log.WithField("noteText", noteText).Debug("Persisting new note text...")

	{
		out, err := gitWrapper.NotesAdd(notesRef, noteText)
		if err != nil {
			return convertGitOutputToError(out, err)
		}
	}

	return nil
}

func calculateKeyValues(gitWrapper GitWrapper, notesRef string) (values *Values, err error) {
	notes, err := getRelevantNotes(gitWrapper, notesRef)
	if err != nil {
		return nil, err
	}

	if len(notes) == 0 {
		log.WithField("ref", notesRef).Warning("No prior notes found")
	}

	events, err := getEventsFromNotes(gitWrapper, notesRef, notes)
	if err != nil {
		return nil, err
	}

	return calculateKeyValuesFromEvents(events)
}

func getRelevantNotes(gitWrapper GitWrapper, notesRef string) (notes []string, err error) {
	allNotes, err := getNotesHashes(gitWrapper, notesRef)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"first 10 notes":   util.LimitStringSlice(allNotes, 10),
		"Total # of notes": len(allNotes),
	}).Debug()

	// Try to get one more commit so we can detect if commits were exhausted in case no note was found
	commits, err := getCommitHashes(gitWrapper)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"first 10 commits":   util.LimitStringSlice(commits, 10),
		"Total # of commits": len(commits),
	}).Debug()

	// Get all notes for commits
	notes = util.GetSlicesIntersect(commits, allNotes)
	log.WithFields(log.Fields{
		"first 10 notes found":   util.LimitStringSlice(notes, 10),
		"Total # of notes found": len(notes),
	}).Debug()

	return notes, nil
}

func getEventsFromNotes(gitWrapper GitWrapper, notesRef string, notes []string) (events []event.Event, err error) {
	for i := len(notes) - 1; i >= 0; i-- { // Iterate from old to new (newest note in front)
		n := notes[i]
		e, err := getEventsFromNote(gitWrapper, notesRef, n)
		if err != nil {
			return nil, err
		}
		events = append(events, e...)
	}
	return events, nil
}

func getEventsFromNote(gitWrapper GitWrapper, notesRef string, note string) (events []event.Event, err error) {
	var noteText string
	{
		out, err := gitWrapper.NotesShow(notesRef, note)
		if err != nil {
			return nil, convertGitOutputToError(out, err)
		}
		noteText = out
	}

	if noteText != "" {
		log.WithField("rawText", noteText).Debug("Unmarshalling...")
		err = event.Unmarshal(noteText, &events)

		if err != nil {
			return nil, err
		}
	}

	return events, nil
}

func calculateKeyValuesFromEvents(events []event.Event) (values *Values, err error) {
	v := NewValues()

	for _, e := range events { // Iterate from old to new (oldest event in front)
		switch e.EventType {
		case event.Set:
			v.Add(e.Key, Value(*e.Value))
		case event.Unset:
			v.Remove(e.Key)
		default:
			log.Fatal("Fatal: Unknown event type encountered")
		}
	}

	return v, nil
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

var getCommitHashes = func(gitWrapper GitWrapper) (hashList []string, err error) {
	out, err := gitWrapper.LogCommits()
	if err != nil {
		return nil, convertGitOutputToError(out, err)
	}

	if out == "" {
		hashList = []string{}
	} else {
		out := strings.TrimSuffix(out, "\n")
		hashList = strings.Split(out, "\n")
	}

	return hashList, nil
}

var getNotesHashes = func(gitWrapper GitWrapper, notesRef string) (hashList []string, err error) {
	out, err := gitWrapper.NotesList(notesRef)
	if err != nil {
		return nil, convertGitOutputToError(out, err)
	}

	if out == "" {
		hashList = []string{}
	} else {
		out := strings.TrimSuffix(out, "\n")
		lines := strings.Split(out, "\n")
		for _, line := range lines {
			hashes := strings.Split(line, " ")
			hashList = append(hashList, hashes[1])
		}
	}
	return hashList, nil
}
