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

// NewCommand creates the root command for the claim subcommand.
//
// It constructs a new cobra.Command instance, configures it with
// usage information and registers any child commands by calling AddCommand.
// The returned *cobra.Command is ready to be added to the main application
// command tree.
func NewCommand() *cobra.Command {
	claimCommand.AddCommand(compare.NewCommand())
	claimCommand.AddCommand(show.NewCommand())

	return claimCommand
}
