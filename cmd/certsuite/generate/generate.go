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

// NewCommand Builds the generate CLI command with its subcommands
//
// This function initializes a cobra.Command for the generate group and
// registers several child commands—catalog, feedback, config, and QE coverage
// reporting—by calling their NewCommand functions. It then returns the fully
// configured parent command ready to be added to the main application root. The
// returned value is a pointer to the cobra.Command instance.
func NewCommand() *cobra.Command {
	generate.AddCommand(catalog.NewCommand())
	generate.AddCommand(feedback.NewCommand())
	generate.AddCommand(config.NewCommand())
	generate.AddCommand(qecoverage.NewCommand())

	return generate
}
