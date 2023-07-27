package claim

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/add"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/compare"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/show"
)

const (
	supportedClaimFormatVersion = "v0.1.0"
)

var (
	claimCommand = &cobra.Command{
		Use:   "claim",
		Short: "Help tools for working with claim files.",
	}
)

func NewCommand() *cobra.Command {
	claimCommand.AddCommand(add.NewCommand())
	claimCommand.AddCommand(compare.NewCommand())
	claimCommand.AddCommand(show.NewCommand())

	return claimCommand
}

func CheckClaimVersion(version string) error {
	claimSemVersion, err := semver.NewVersion(version)
	if err != nil {
		return fmt.Errorf("claim file version %q is not valid: %v", version, err)
	}

	supportedSemVersion, err := semver.NewVersion(supportedClaimFormatVersion)
	if err != nil {
		return fmt.Errorf("supported claim file version v%v is not valid: v%v", supportedClaimFormatVersion, err)
	}

	if claimSemVersion.Compare(supportedSemVersion) != 0 {
		return fmt.Errorf("claim format version v%v is not supported. Supported version is v%v",
			claimSemVersion, supportedSemVersion)
	}

	return nil
}
