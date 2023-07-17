package enrich

import (
	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/enrich/preflight"
)

var (
	enrichCommand = &cobra.Command{
		Use:   "enrich",
		Short: "Enriches existing files with additional information.",
	}
)

func NewCommand() *cobra.Command {
	enrichCommand.AddCommand(preflight.NewCommand())

	return enrichCommand
}
