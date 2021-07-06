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

			gitWrapper := git.GetGitWrapperFrom(cmd.Context())
			out, err := getValue(gitWrapper, globalFlags.NotesRef, key, globalFlags.MaxDepth)
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

func getValue(gitWrapper git.Wrapper, notesRef string, key string, maxDepth uint) (string, error) {
	key = sanitizeKey(key)
	err := validateKey(key)
	if err != nil {
		return "", err
	}

	values, err := getNoteValues(gitWrapper, notesRef, maxDepth)
	if err != nil {
		return "", err
	}

	return values.GetJSON(key), nil
}
