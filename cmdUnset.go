package main

import (
	"log"

	"github.com/philips-software/gino-keva/internal/git"
	"github.com/spf13/cobra"
)

func addUnsetCommandTo(root *cobra.Command) {
	var (
		push bool
	)

	var unsetCommand = &cobra.Command{
		Use:   "unset [key]",
		Short: "Unset a key",
		Long:  `Unset a key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			gitWrapper := git.GetGitWrapperFrom(cmd.Context())

			err := unset(gitWrapper, globalFlags.NotesRef, key, globalFlags.MaxDepth)
			if err != nil {
				return err
			}

			if push {
				err = pushNotes(gitWrapper, globalFlags.NotesRef)
			}

			return err
		},
		Args: cobra.ExactArgs(1),
	}

	unsetCommand.Flags().BoolVar(&push, "push", false, "Push notes to upstream")
	root.AddCommand(unsetCommand)
}

func unset(gitWrapper git.Wrapper, notesRef string, key string, maxDepth uint) error {
	key = sanitizeKey(key)
	err := validateKey(key)
	if err != nil {
		return err
	}

	values, err := getNoteValues(gitWrapper, notesRef, maxDepth)
	if err != nil {
		return err
	}

	values.Remove(key)

	noteText, err := convertValuesToOutput(values, "json")
	if err != nil {
		return err
	}

	{
		out, err := gitWrapper.NotesAdd(notesRef, noteText)
		if err != nil {
			log.Fatal(out)
		}
	}

	return err
}
