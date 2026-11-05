package claim

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/show"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/cli"
	"github.com/spf13/cobra"
)

var (
	claimCommand = &cobra.Command{
		Use:   "claim",
		Short: "Help tools for working with claim files.",
	}
)

func NewCommand() *cobra.Command {
	cli.AddCommand(claimCommand, compare.NewCommand())
	cli.AddCommand(claimCommand, show.NewCommand())

	return claimCommand
}
