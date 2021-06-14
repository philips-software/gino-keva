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
			notesAccess := git.GetNotesAccessFrom(cmd.Context())

			err := unset(notesAccess, globalFlags.NotesRef, key, globalFlags.MaxDepth)
			if err != nil {
				return err
			}

			if push {
				err = pushNotes(notesAccess, globalFlags.NotesRef)
			}

			return err
		},
		Args: cobra.ExactArgs(1),
	}

	unsetCommand.Flags().BoolVar(&push, "push", false, "Push notes to upstream")
	root.AddCommand(unsetCommand)
}

func unset(notesAccess git.Notes, notesRef string, key string, maxDepth int) error {
	key = sanitizeKey(key)
	err := validateKey(key)
	if err != nil {
		return err
	}

	values, err := getNoteValues(notesAccess, notesRef, maxDepth)
	if err != nil {
		return err
	}

	values.Remove(key)

	noteText, err := convertValuesToOutput(values, "json")
	if err != nil {
		return err
	}

	{
		out, err := notesAccess.Add(notesRef, noteText)
		if err != nil {
			log.Fatal(out)
		}
	}

	return err
}
