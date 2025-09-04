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

// NewCommand Creates a subcommand for claim operations
//
// It initializes the claim command by attaching its compare and show
// subcommands, each of which provides functionality for comparing claim files
// or displaying claim information. The function returns the configured
// cobra.Command ready to be added to the main application root command.
func NewCommand() *cobra.Command {
	claimCommand.AddCommand(compare.NewCommand())
	claimCommand.AddCommand(show.NewCommand())

	return claimCommand
}
