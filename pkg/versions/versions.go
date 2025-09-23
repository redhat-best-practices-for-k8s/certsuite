package versions

import (
	"regexp"

	"github.com/Masterminds/semver/v3"
)

var (
	// GitCommit is the latest commit in the current git branch
	GitCommit string
	// GitRelease is the list of tags (if any) applied to the latest commit
	// in the current branch
	GitRelease string
	// GitPreviousRelease is the last release at the date of the latest commit
	// in the current branch
	GitPreviousRelease string
	// GitDisplayRelease is a string used to hold the text to display
	// the version on screen and in the claim file
	GitDisplayRelease string
	// ClaimFormat is the current version for the claim file format to be produced by the TNF test suite.
	// A client decoding this claim file must support decoding its specific version.
	ClaimFormatVersion string
)

// GitVersion provides the current build’s git display version
//
// The function checks if a release tag is defined; if not it falls back to an
// unreleased build label combined with the previous release information. It
// then appends the short commit hash in parentheses and returns the resulting
// string, which is used throughout the application to report the running
// version.
func GitVersion() string {
	if GitRelease == "" {
		GitDisplayRelease = "Unreleased build post " + GitPreviousRelease
	} else {
		GitDisplayRelease = GitRelease
	}

	return GitDisplayRelease + " (" + GitCommit + ")"
}

// IsValidSemanticVersion Validates that a string is a proper semantic version
//
// The function attempts to parse the input using a semantic version parser. If
// parsing succeeds without error, it returns true, indicating a valid semantic
// version; otherwise, it returns false.
func IsValidSemanticVersion(version string) bool {
	_, err := semver.NewVersion(version)
	return err == nil
}

// IsValidK8sVersion Checks if a string matches Kubernetes version naming conventions
//
// The function compiles a regular expression that enforces the pattern for
// Kubernetes versions, allowing optional pre-release identifiers such as alpha
// or beta with numeric suffixes. It returns true when the input string conforms
// to this format and false otherwise.
func IsValidK8sVersion(version string) bool {
	r := regexp.MustCompile(`^(v)([1-9]\d*)+((alpha|beta)([1-9]\d*)+){0,2}$`)
	return r.MatchString(version)
}
