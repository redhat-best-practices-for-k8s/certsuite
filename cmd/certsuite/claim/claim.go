package claim

import (
	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/claim/compare"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/claim/show"
)

var (
	claimCommand = &cobra.Command{
		Use:   "claim",
		Short: "Help tools for working with claim files.",
	}
)

func NewCommand() *cobra.Command {
	claimCommand.AddCommand(compare.NewCommand())
	claimCommand.AddCommand(show.NewCommand())

	return claimCommand
}
