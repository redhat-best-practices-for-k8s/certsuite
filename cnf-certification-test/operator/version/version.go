package version

import (
	"regexp"

	"github.com/Masterminds/semver"
)

func IsValidSemanticVersion(version string) bool {
	_, err := semver.NewVersion(version)
	return err == nil
}

func IsValidK8sVersion(version string) bool {
	r := regexp.MustCompile(`^(v)([1-9]\d*)+((alpha|beta)([1-9]\d*)+){0,2}$`)
	return r.MatchString(version)
}
