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

// GitVersion returns the git display version string for this build.
//
// It selects the most appropriate version to display: if the current
// build is a released one, it uses that release number; otherwise it falls
// back to the previous release or the commit hash, ensuring a meaningful
// version identifier is always available.
func GitVersion() string {
	if GitRelease == "" {
		GitDisplayRelease = "Unreleased build post " + GitPreviousRelease
	} else {
		GitDisplayRelease = GitRelease
	}

	return GitDisplayRelease + " (" + GitCommit + ")"
}

// IsValidSemanticVersion reports whether a string is a valid semantic version.
//
// It attempts to parse the provided string using the semantic versioning rules
// defined in the package. If parsing succeeds, it returns true; otherwise,
// false. This function does not return an error, so callers should check the
// boolean result before using the parsed value.
func IsValidSemanticVersion(version string) bool {
	_, err := semver.NewVersion(version)
	return err == nil
}

// IsValidK8sVersion determines whether a given Kubernetes version string is valid.
//
// It accepts a single string argument representing the version to check.
// The function uses a regular expression to verify that the format
// conforms to expected Kubernetes semantic versioning (e.g., "1.22.3").
// It returns true if the input matches the pattern, otherwise false.
func IsValidK8sVersion(version string) bool {
	r := regexp.MustCompile(`^(v)([1-9]\d*)+((alpha|beta)([1-9]\d*)+){0,2}$`)
	return r.MatchString(version)
}
