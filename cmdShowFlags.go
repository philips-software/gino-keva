package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func addShowFlagCommandTo(root *cobra.Command) {
	var showFlagCommand = &cobra.Command{
		Use:    "show-flag [name]",
		Short:  "Show flag value",
		Long:   `Show the value of a flag`,
		Hidden: true, // For testing purposes only
		RunE: func(cmd *cobra.Command, args []string) error {
			flagName := args[0]
			flagValue, err := cmd.Flags().GetString(flagName)

			if err != nil && strings.HasPrefix(err.Error(), "flag accessed but not defined:") {
				return fmt.Errorf("unknown flag [%v]", flagName)
			}

			fmt.Fprint(cmd.OutOrStdout(), flagValue)
			return nil
		},
		Args: cobra.ExactArgs(1),
	}
	root.AddCommand(showFlagCommand)
}
