package main

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

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

			err = set(gitWrapper, globalFlags.NotesRef, key, value, globalFlags.MaxDepth)
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

func set(gitWrapper GitWrapper, notesRef string, key string, value string, maxDepth uint) (err error) {
	key = sanitizeKey(key)
	err = validateKey(key)
	if err != nil {
		return err
	}

	values, err := getNoteValues(gitWrapper, notesRef, maxDepth)
	if err != nil {
		return err
	}

	var commitHash string
	{
		out, err := gitWrapper.RevParseHead()
		if err != nil {
			return err
		}
		commitHash = truncateHash(strings.TrimSuffix(out, "\n"), 8)
	}

	values.Add(key, Value{
		Data:   value,
		Source: commitHash,
	})

	noteText, err := convertValuesToOutput(values, "raw")
	if err != nil {
		return err
	}

	{
		out, err := gitWrapper.NotesAdd(notesRef, noteText)
		if err != nil {
			log.Fatal(out)
		}
	}

	log.WithFields(log.Fields{
		"key":        key,
		"value":      value,
		"commitHash": commitHash,
	}).Debug("Key/value added successfully")

	return err
}

func sanitizeKey(key string) string {
	return strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
}

func validateKey(key string) error {
	if key == "" {
		return &InvalidKey{msg: "key cannot be empty"}
	}

	{
		pattern := `[^A-Za-z0-9_]`
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
		pattern := `[^A-Za-z0-9]$`
		matched, err := regexp.Match(pattern, []byte(key))
		if err != nil {
			return err
		}

		if matched {
			return &InvalidKey{msg: "last character is not a letter or number"}
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
