package operator

import (
	"github.com/Masterminds/semver/v3"
)

func isValidSemanticVersion(version string) bool {
	_, err := semver.NewVersion(version)
	return err == nil
}
