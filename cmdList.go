package main

import (
	"encoding/json"
	"fmt"

	"github.com/philips-software/gino-keva/internal/git"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// InvalidOutputFormat error indicates the specified output format is invalid
type InvalidOutputFormat struct {
}

func (InvalidOutputFormat) Error() string {
	return "Invalid output format specified"
}

func addListCommandTo(root *cobra.Command) {
	var (
		outputFormat string
	)

	var listCommand = &cobra.Command{
		Use:   "list",
		Short: "List",
		Long:  `List all of the keys and values currently stored`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			gitWrapper := git.GetGitWrapperFrom(cmd.Context())

			if globalFlags.Fetch {
				err = fetchNotes(gitWrapper)
				if err != nil {
					return err
				}
			}

			out, err := getListOutput(gitWrapper, globalFlags.NotesRef, globalFlags.MaxDepth, outputFormat)
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
		Args: cobra.NoArgs,
	}
	listCommand.Flags().StringVarP(&outputFormat, "output", "o", "plain", "Set output format (plain/json/raw)")

	root.AddCommand(listCommand)
}

func getListOutput(gitWrapper git.Wrapper, notesRef string, maxDepth uint, outputFormat string) (out string, err error) {
	values, err := getNoteValues(gitWrapper, notesRef, maxDepth)
	if err != nil {
		return "", err
	}

	return convertValuesToOutput(values, outputFormat)
}

func getNoteValues(gitWrapper git.Wrapper, notesRef string, maxDepth uint) (values *Values, err error) {
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

func findNoteText(gitWrapper git.Wrapper, notesRef string, maxDepth uint) (noteText string, err error) {
	notes, err := getNotesHashes(gitWrapper, notesRef)
	if err != nil {
		return "", err
	}
	log.WithFields(log.Fields{
		"first 10 notes":   limitStringSlice(notes, 10),
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
		"first 10 commits":   limitStringSlice(commits, 10),
		"Total # of commits": len(commits),
	}).Debug()

	// Get all notes for commits up to maxDepth
	notesIntersect := getNotesIntersect(notes, limitStringSlice(commits, maxCount))
	log.WithFields(log.Fields{
		"first 10 notesIntersect":   limitStringSlice(notesIntersect, 10),
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

func limitStringSlice(slice []string, limit uint) []string {
	if len(slice) <= int(limit) {
		return slice
	}

	return slice[:limit]
}

func convertValuesToOutput(values *Values, outputFlag string) (out string, err error) {
	switch outputFlag {

	case "plain":
		if values.Count() == 0 {
			out = ""
		} else {
			for k, v := range values.Iterate() {
				out += fmt.Sprintf("%s=%s\n", k, v)
			}
		}

	case "json":
		out, err = marshalJSON(values)

	case "raw":
		out, err = marshalRaw(values)

	default:
		err = &InvalidOutputFormat{}
	}

	return out, err
}

func marshalJSON(values *Values) (string, error) {
	result, err := json.MarshalIndent(values.Iterate(), "", "  ")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n", result), nil
}

func marshalRaw(values *Values) (string, error) {
	result, err := json.Marshal(values.IterateRaw())
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n", result), nil
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
