package main

import (
	"fmt"

	"github.com/philips-software/gino-keva/internal/git"
	"github.com/spf13/cobra"
)

func addGetCommandTo(root *cobra.Command) {
	var getCommand = &cobra.Command{
		Use:   "get [key]",
		Short: "Get the value of a specific key",
		Long:  `Get the value of a specific key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			notesAccess := git.GetNotesAccessFrom(cmd.Context())
			out, err := getValue(notesAccess, globalFlags.NotesRef, key, globalFlags.MaxDepth)
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), out)
			return nil
		},
		Args: cobra.ExactArgs(1),
	}

	root.AddCommand(getCommand)
}

func getValue(notesAccess git.Notes, notesRef string, key string, maxDepth int) (string, error) {
	values, err := getNoteValues(notesAccess, notesRef, maxDepth)
	if err != nil {
		return "", err
	}

	return values.GetJSON(key), nil
}
