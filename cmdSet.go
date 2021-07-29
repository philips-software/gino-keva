package main

import (
	"github.com/philips-software/gino-keva/internal/event"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func addSetCommandTo(root *cobra.Command) {
	var (
		push bool
	)

	var setCommand = &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set the value of a key",
		Long:  `Set the value of a key`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			key := args[0]
			value := args[1]
			gitWrapper := GetGitWrapperFrom(cmd.Context())

			if globalFlags.Fetch {
				err = fetchNotes(gitWrapper)
				if err != nil {
					return err
				}
			}

			err = set(gitWrapper, globalFlags.NotesRef, key, value)
			if err != nil {
				return err
			}

			err = pruneNotes(gitWrapper, globalFlags.NotesRef)
			if err != nil {
				return err
			}

			if push {
				err = pushNotes(gitWrapper, globalFlags.NotesRef)
			}

			return err
		},
		Args: cobra.ExactArgs(2),
	}

	setCommand.Flags().BoolVar(&push, "push", false, "Push notes to upstream")
	root.AddCommand(setCommand)
}

func set(gitWrapper GitWrapper, notesRef string, key string, value string) error {
	setEvent, err := event.NewSetEvent(key, value)
	if err != nil {
		return err
	}

	events, err := getEvents(gitWrapper, notesRef)
	if err != nil {
		return err
	}

	*events = event.AddNewEvent(events, setEvent)

	err = persistEvents(gitWrapper, notesRef, events)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"key":   key,
		"value": value,
	}).Debug("Set event added successfully")

	return nil
}
