package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/philips-internal/gino-keva/internal/git"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func addListCommandTo(root *cobra.Command) {
	var (
		outputFormat string
	)

	var listCommand = &cobra.Command{
		Use:   "list",
		Short: "List",
		Long:  `List all of the keys and values currently stored`,
		RunE: func(cmd *cobra.Command, args []string) error {
			notesAccess := git.GetNotesAccessFrom(cmd.Context())

			out, err := getListOutput(notesAccess, globalFlags.NotesRef, globalFlags.MaxDepth, outputFormat)
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
		Args: cobra.NoArgs,
	}
	listCommand.Flags().StringVarP(&outputFormat, "output", "o", "plain", "Set output format (plain/json)")

	root.AddCommand(listCommand)
}

type findTextOptions struct {
	AttemptsLeft int
	CommitRef    string
	NotesRef     string
}

func getListOutput(notesAccess git.Notes, notesRef string, maxDepth int, outputFormat string) (out string, err error) {
	values, err := getNoteValues(notesAccess, notesRef, maxDepth)
	if err != nil {
		return "", err
	}

	return convertValuesToOutput(values, outputFormat)
}

func getNoteValues(notesAccess git.Notes, notesRef string, maxDepth int) (values *Values, err error) {
	noteText, err := findNoteText(notesAccess, &findTextOptions{
		AttemptsLeft: maxDepth,
		CommitRef:    "HEAD",
		NotesRef:     notesRef,
	})

	if err != nil {
		return nil, err
	}

	values, err = unmarshal(noteText)
	if err != nil {
		return nil, err
	}

	return values, err
}

func findNoteText(notesAccess git.Notes, options *findTextOptions) (string, error) {
	log.WithField("hash", options.CommitRef).Debug("Looking for git notes...")

	out, err := notesAccess.Show(options.NotesRef, options.CommitRef)
	if err == nil {
		log.WithField("note", out).Debug("Found note")
		return out, nil
	} else if strings.HasPrefix(out, "fatal: failed to resolve '") {
		log.WithField("ref", options.NotesRef).Warning("Reached root commit. No prior notes found")
		return "", nil
	} else if strings.HasPrefix(out, "error: no note found for object ") {
		log.Debug("No note here")
	} else {
		log.Error(out)
		return "", err
	}

	if options.AttemptsLeft == 0 {
		log.WithField("ref", options.NotesRef).Warning("No prior notes found within maximum depth!")

		return "", nil
	}

	options.CommitRef = fmt.Sprintf("%v^", options.CommitRef)
	options.AttemptsLeft--

	return findNoteText(notesAccess, options)
}

func convertValuesToOutput(values *Values, outputFlag string) (out string, err error) {
	switch outputFlag {
	case "plain":
		if values.Count() == 0 {
			out = "\n"
		} else {
			for k, v := range values.Iterate() {
				out += fmt.Sprintf("%s=%s\n", k, v.Data)
			}
		}
	case "json":
		out, err = marshal(values)
	}

	return out, err
}

func marshal(values *Values) (string, error) {
	result, err := json.Marshal(values.Iterate())
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
