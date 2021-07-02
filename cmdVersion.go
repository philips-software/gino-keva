package main

import (
	"fmt"

	"github.com/philips-software/gino-keva/internal/versioninfo"
	"github.com/spf13/cobra"
)

func addVersionCommandTo(root *cobra.Command) {
	var versionCommand = &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Long:  `Show version info`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprint(cmd.OutOrStdout(), versioninfo.Get())
			return nil
		},
		Args: cobra.NoArgs,
	}

	root.AddCommand(versionCommand)
}
