package main

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/philips-software/gino-keva/internal/git"
	"github.com/spf13/cobra"
)

// InvalidKey error indicates the key is not valid
type InvalidKey struct {
	msg string
}

func (i InvalidKey) Error() string {
	return fmt.Sprintf("Invalid key: %v", i.msg)
}

func addSetCommandTo(root *cobra.Command) {
	var (
		push bool
	)

	var setCommand = &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set the value of a key",
		Long:  `Set the value of a key`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
			notesAccess := git.GetNotesAccessFrom(cmd.Context())

			err := set(notesAccess, globalFlags.NotesRef, key, value, globalFlags.MaxDepth)
			if err != nil {
				return err
			}

			if push {
				err = pushNotes(notesAccess, globalFlags.NotesRef)
			}

			return err
		},
		Args: cobra.ExactArgs(2),
	}

	setCommand.Flags().BoolVar(&push, "push", false, "Push notes to upstream")
	root.AddCommand(setCommand)
}

func set(notesAccess git.Notes, notesRef string, key string, value string, maxDepth int) (err error) {
	err = validateKey(key)
	if err != nil {
		return err
	}

	values, err := getNoteValues(notesAccess, notesRef, maxDepth)
	if err != nil {
		return err
	}

	var commitHash string
	{
		out, err := notesAccess.RevParseHead()
		if err != nil {
			return err
		}
		commitHash = strings.TrimSuffix(out, "\n")
	}

	values.Add(key, Value{
		Data:   value,
		Source: truncateHash(commitHash, 8),
	})

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

func validateKey(key string) error {
	if key == "" {
		return &InvalidKey{msg: "key cannot be empty"}
	}

	{
		pattern := `[^A-Za-z0-9_-]`
		matched, err := regexp.Match(pattern, []byte(key))
		if err != nil {
			return err
		}

		if matched {
			return &InvalidKey{msg: "key contains invalid characters"}
		}
	}

	{
		pattern := `^[^A-Za-z]`
		matched, err := regexp.Match(pattern, []byte(key))
		if err != nil {
			return err
		}

		if matched {
			return &InvalidKey{msg: "first character is not a letter"}
		}
	}

	{
		pattern := `[^A-Za-z]$`
		matched, err := regexp.Match(pattern, []byte(key))
		if err != nil {
			return err
		}

		if matched {
			return &InvalidKey{msg: "last character is not a letter"}
		}
	}

	return nil
}

func truncateHash(hash string, chars int) string {
	if len(hash) > chars {
		return hash[:chars]
	}
	return hash
}
