package main

import (
	"errors"
	"strings"
)

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

// NoNotePresent error indicates there's no note present on HEAD commit
type NoNotePresent struct {
}

func (NoNotePresent) Error() string {
	return "No note present on HEAD commit"
}

func checkIfErrorStringIsNoNotePresent(s string) bool {
	return strings.HasPrefix(strings.ToLower(s), "error: no note found for object ")
}

func convertGitOutputToError(out string, errorCode error) (err error) {
	if errorCode == nil {
		return nil
	}

	if checkIfErrorStringIsNoRemoteRef(out) {
		err = &NoRemoteRef{}
	} else if checkIfErrorStringIsNoNotePresent(out) {
		err = &NoNotePresent{}
	} else if checkIfErrorStringIsUpstreamChanged(out) {
		err = &UpstreamChanged{fetchEnabled: globalFlags.Fetch}
	} else {
		err = errors.New(out)
	}

	return err
}
