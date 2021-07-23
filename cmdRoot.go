package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NewRootCommand builds the cobra command that handles our command line tool.
func NewRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "gino-keva",
		Short: "A tool to store key value data as git notes",
		Long: `Git Notes Key Value (gino-keva) is a tool used to store and manage key values
in git notes. You can store any sort of data you want against each commit in your
repository`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			initializeConfig(cmd)
			return err
		},
	}

	addRootFlagsTo(rootCommand)
	addShowFlagCommandTo(rootCommand)
	addListCommandTo(rootCommand)
	addGetCommandTo(rootCommand)
	addSetCommandTo(rootCommand)
	addUnsetCommandTo(rootCommand)
	addVersionCommandTo(rootCommand)

	return rootCommand
}

func initializeConfig(cmd *cobra.Command) {
	v := viper.New()

	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv() // read in environment variables that match

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	verboseLogs, err := cmd.Flags().GetBool("verbose")
	checkIfError(err)

	if verboseLogs {
		log.SetLevel(log.DebugLevel)
	}
}

// Bind each cobra flag to its associated viper configuration
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func addRootFlagsTo(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&globalFlags.NotesRef, "ref", "gino_keva", "Name of notes reference")
	cmd.PersistentFlags().BoolVarP(&globalFlags.VerboseLog, "verbose", "v", false, "Turn on verbose logging")

	cmd.PersistentFlags().BoolVar(&globalFlags.Fetch, "fetch", true, "Fetch notes from upstream")
}
