package claim

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/show"
	"github.com/spf13/cobra"
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
