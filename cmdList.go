package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/philips-software/gino-keva/internal/git"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			gitWrapper := git.GetGitWrapperFrom(cmd.Context())

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

type findTextOptions struct {
	AttemptsLeft int
	CommitRef    string
	NotesRef     string
}

func getListOutput(gitWrapper git.Wrapper, notesRef string, maxDepth int, outputFormat string) (out string, err error) {
	values, err := getNoteValues(gitWrapper, notesRef, maxDepth)
	if err != nil {
		return "", err
	}

	return convertValuesToOutput(values, outputFormat)
}

func getNoteValues(gitWrapper git.Wrapper, notesRef string, maxDepth int) (values *Values, err error) {
	noteText, err := findNoteText(gitWrapper, &findTextOptions{
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

func findNoteText(gitWrapper git.Wrapper, options *findTextOptions) (string, error) {
	log.WithField("hash", options.CommitRef).Debug("Looking for git notes...")

	out, err := gitWrapper.ShowNote(options.NotesRef, options.CommitRef)
	if err == nil {
		log.WithField("note", out).Debug("Found note")
		return out, nil
	}

	if strings.HasPrefix(out, "fatal: failed to resolve '") {
		log.WithField("ref", options.NotesRef).Warning("Reached root commit. No prior notes found")
		return "", nil
	}

	r, _ := regexp.Compile("error: no note found for object ([a-f0-9]+).\n")
	matches := r.FindStringSubmatch(out)

	if len(matches) == 0 {
		log.Error(out)
		return "", err
	}

	options.CommitRef = matches[1]
	log.WithField("resolvedRef", options.CommitRef).Debug("No notes here.")

	if options.AttemptsLeft == 0 {
		log.WithField("ref", options.NotesRef).Warning("No prior notes found within maximum depth!")

		return "", nil
	}

	options.CommitRef = fmt.Sprintf("%v^", options.CommitRef)
	options.AttemptsLeft--

	return findNoteText(gitWrapper, options)
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
