package generate

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/catalog"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/config"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/feedback"
	qecoverage "github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/qe_coverage"
	"github.com/spf13/cobra"
)

var (
	generate = &cobra.Command{
		Use:   "generate",
		Short: "generator tool for various test suite assets",
	}
)

// NewCommand creates the root generate command.
//
// It constructs a new cobra.Command that serves as the entry point
// for the generate subcommand of certsuite. The function sets up
// the command's usage, description, and registers any child commands
// by calling AddCommand on the returned command.
// The resulting *cobra.Command is returned to be added to the main
// application command tree.
func NewCommand() *cobra.Command {
	generate.AddCommand(catalog.NewCommand())
	generate.AddCommand(feedback.NewCommand())
	generate.AddCommand(config.NewCommand())
	generate.AddCommand(qecoverage.NewCommand())

	return generate
}
