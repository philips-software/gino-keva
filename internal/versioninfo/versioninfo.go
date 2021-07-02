package versioninfo

import "fmt"

//// Following variables will be statically linked at the time of compiling

// Version holds contents of ./VERSION file, if exists, or the value passed via the -version option
var Version = "<Version not set>"

// BuildDate holds RFC3339 formatted UTC date (build time)
var BuildDate = "<BuildDate not set>"

// GitCommit holds short commit hash of source tree
var GitCommit = "<GitCommit not set>"

// GitState shows whether there are uncommitted changes
var GitState = "<GitState not set>"

// Get returns the version information
func Get() string {
	return fmt.Sprintf(
		`version.Info{`+
			`Version: "%v", `+
			`BuildDate: "%v", `+
			`GitCommit: "%v", `+
			`GitState: "%v"`+
			`}`,
		Version, BuildDate, GitCommit, GitState)
}
