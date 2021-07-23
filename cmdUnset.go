package main

import (
	"github.com/philips-software/gino-keva/internal/event"
	log "github.com/sirupsen/logrus"
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
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			key := args[0]
			gitWrapper := GetGitWrapperFrom(cmd.Context())

			if globalFlags.Fetch {
				err = fetchNotes(gitWrapper)
				if err != nil {
					return err
				}
			}

			err = unset(gitWrapper, globalFlags.NotesRef, key)
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

func unset(gitWrapper GitWrapper, notesRef string, key string) error {
	unsetEvent, err := event.NewUnsetEvent(key)
	if err != nil {
		return err
	}

	err = persistNewEvent(gitWrapper, notesRef, unsetEvent)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"key": key,
	}).Debug("Unset event added successfully")

	return nil
}
