package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/philips-software/gino-keva/internal/git"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	maxRetryAttempts = 3
)

var globalFlags = struct {
	MaxDepth   uint
	NotesRef   string
	VerboseLog bool

	Fetch bool
}{}

// UpstreamChanged error indicates there's been a change in the upstream preventing a fetch/push without force
type UpstreamChanged struct {
	fetchEnabled bool
}

func (UpstreamChanged) Error() string {
	return "Upstream has changed in the meanwhile"
}

func checkIfErrorStringIsUpstreamChanged(s string) bool {
	return strings.Contains(s, "! [rejected]") || strings.Contains(s, "! [remote rejected]")
}

// NoRemoteRef error indicates that the remote reference isn't there
type NoRemoteRef struct {
}

func (NoRemoteRef) Error() string {
	return "No remote reference found"
}

func checkIfErrorStringIsNoRemoteRef(s string) bool {
	return strings.HasPrefix(strings.ToLower(s), "fatal: couldn't find remote ref refs/notes/")
}

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

			gitWrapper := git.GetGitWrapperFrom(cmd.Context())

			if globalFlags.Fetch {
				err = fetchNotes(gitWrapper, globalFlags.NotesRef, false)

				if _, ok := err.(*UpstreamChanged); ok {
					log.Warning("Unpushed local changes are now discarded")
					err = fetchNotes(gitWrapper, globalFlags.NotesRef, true)
				}

				if _, ok := err.(*NoRemoteRef); ok {
					log.WithField("notesRef", globalFlags.NotesRef).Debug("Couldn't find remote ref. Nothing fetched")
					err = nil
				}

				if err != nil {
					log.Error(err.Error())
				}
			}

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

func fetchNotes(gitWrapper git.Wrapper, notesRef string, force bool) error {
	log.WithField("force", force).Debug("Fetching notes...")
	defer log.Debug("Done.")

	out, errorCode := gitWrapper.FetchNotes(globalFlags.NotesRef, force)
	return convertGitOutputToError(out, errorCode)
}

func pruneNotes(gitWrapper git.Wrapper, notesRef string) error {
	log.Debug("Pruning notes...")
	defer log.Debug("Done.")

	out, errorCode := gitWrapper.NotesPrune(globalFlags.NotesRef)
	return convertGitOutputToError(out, errorCode)
}

func pushNotes(gitWrapper git.Wrapper, notesRef string) error {
	log.Debug("Pushing notes...")
	defer log.Debug("Done.")

	out, errorCode := gitWrapper.PushNotes(globalFlags.NotesRef)
	err := convertGitOutputToError(out, errorCode)

	if _, ok := err.(*UpstreamChanged); ok {
		return err
	}

	if err != nil {
		log.Error(out)
	}

	return err
}

func convertGitOutputToError(out string, errorCode error) (err error) {
	if errorCode == nil {
		return nil
	}

	if checkIfErrorStringIsNoRemoteRef(out) {
		err = &NoRemoteRef{}
	} else if checkIfErrorStringIsUpstreamChanged(out) {
		err = &UpstreamChanged{fetchEnabled: globalFlags.Fetch}
	} else {
		err = errors.New(out)
	}

	return err
}

func getCommitHashes(gitWrapper git.Wrapper, maxCount uint) (hashList []string, err error) {
	output, err := gitWrapper.LogCommits(maxCount)
	if err != nil {
		return nil, err
	}

	if output == "" {
		hashList = []string{}
	} else {
		output := strings.TrimSuffix(output, "\n")
		hashList = strings.Split(output, "\n")
	}

	return hashList, nil
}

func getNotesHashes(gitWrapper git.Wrapper, notesRef string) (hashList []string, err error) {
	output, err := gitWrapper.NotesList(notesRef)
	if err != nil {
		return nil, err
	}

	if output == "" {
		hashList = []string{}
	} else {
		output := strings.TrimSuffix(output, "\n")
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			hashes := strings.Split(line, " ")
			hashList = append(hashList, hashes[1])
		}
	}
	return hashList, nil
}

func getNotesIntersect(notes, commits []string) (notesIntersect []string) {
	for i := 0; i < len(notes); i++ {
		if contains(commits, notes[i]) {
			notesIntersect = append(notesIntersect, notes[i])
		}
	}
	return notesIntersect
}

func contains(slice []string, s string) bool {
	for _, a := range slice {
		if a == s {
			return true
		}
	}
	return false
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
	cmd.PersistentFlags().UintVar(&globalFlags.MaxDepth, "max-depth", 500, "Set maximum search depth for a note")
	cmd.PersistentFlags().StringVar(&globalFlags.NotesRef, "ref", "gino_keva", "Name of notes reference")
	cmd.PersistentFlags().BoolVarP(&globalFlags.VerboseLog, "verbose", "v", false, "Turn on verbose logging")

	cmd.PersistentFlags().BoolVar(&globalFlags.Fetch, "fetch", true, "Fetch notes from upstream")
}
