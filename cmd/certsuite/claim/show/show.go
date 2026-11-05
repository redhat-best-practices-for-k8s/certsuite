package show

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/show/csv"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/show/failures"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/cli"
	"github.com/spf13/cobra"
)

var (
	showCommand = &cobra.Command{
		Use:   "show",
		Short: "Shows information from a claim file.",
	}
)

func NewCommand() *cobra.Command {
	cli.AddCommand(showCommand, failures.NewCommand())
	cli.AddCommand(showCommand, csv.NewCommand())
	return showCommand
}
