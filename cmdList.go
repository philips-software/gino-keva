package main

import (
	"encoding/json"
	"fmt"

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
			gitWrapper := GetGitWrapperFrom(cmd.Context())

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

func getListOutput(gitWrapper GitWrapper, notesRef string, maxDepth uint, outputFormat string) (out string, err error) {
	values, err := getNoteValues(gitWrapper, notesRef, maxDepth)
	if err != nil {
		return "", err
	}

	return convertValuesToOutput(values, outputFormat)
}

func convertValuesToOutput(values *Values, outputFlag string) (out string, err error) {
	switch outputFlag {

	case "plain":
		if values.Count() == 0 {
			out = ""
		} else {
			for k, v := range values.IterateSnapshot() {
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
