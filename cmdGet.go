package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func addGetCommandTo(root *cobra.Command) {
	var getCommand = &cobra.Command{
		Use:   "get [key]",
		Short: "Get the value of a specific key",
		Long:  `Get the value of a specific key`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			key := args[0]

			gitWrapper := GetGitWrapperFrom(cmd.Context())

			if globalFlags.Fetch {
				err = fetchNotes(gitWrapper)
				if err != nil {
					return err
				}
			}

			out, err := getValue(gitWrapper, globalFlags.NotesRef, key)
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

func getValue(gitWrapper GitWrapper, notesRef string, key string) (string, error) {
	values, err := calculateKeyValues(gitWrapper, notesRef)
	if err != nil {
		return "", err
	}

	return string(values.Get(key)), nil
}
