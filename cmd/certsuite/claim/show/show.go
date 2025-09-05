package show

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/show/csv"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/show/failures"
	"github.com/spf13/cobra"
)

var (
	showCommand = &cobra.Command{
		Use:   "show",
		Short: "Shows information from a claim file.",
	}
)

// NewCommand Creates the show command with its subcommands
//
// This function constructs a Cobra command responsible for displaying claim
// information. It registers two child commands—one that shows failures and
// another that outputs CSV dumps—by adding them to the parent command before
// returning it. The returned command can then be integrated into the larger CLI
// hierarchy.
func NewCommand() *cobra.Command {
	showCommand.AddCommand(failures.NewCommand())
	showCommand.AddCommand(csv.NewCommand())
	return showCommand
}
