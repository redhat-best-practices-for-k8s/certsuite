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

// NewCommand returns the root claim show command.
//
// It constructs a new cobra.Command that represents the "show" subcommand
// of the certsuite CLI, configures its usage string, short description,
// and registers any child commands needed for displaying claim information.
// The returned command can be added to a parent command with AddCommand.
func NewCommand() *cobra.Command {
	showCommand.AddCommand(failures.NewCommand())
	showCommand.AddCommand(csv.NewCommand())
	return showCommand
}
