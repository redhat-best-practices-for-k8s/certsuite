package show

import (
	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/show/csvtelco"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/show/failures"
)

var (
	showCommand = &cobra.Command{
		Use:   "show",
		Short: "Shows information from a claim file.",
	}
)

func NewCommand() *cobra.Command {
	showCommand.AddCommand(failures.NewCommand())
	showCommand.AddCommand(csvtelco.NewCommand())
	return showCommand
}
