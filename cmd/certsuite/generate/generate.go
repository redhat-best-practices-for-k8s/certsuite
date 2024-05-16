package generate

import (
	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/generate/catalog"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/generate/config"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/generate/feedback"
	qecoverage "github.com/test-network-function/cnf-certification-test/cmd/certsuite/generate/qe_coverage"
)

var (
	generate = &cobra.Command{
		Use:   "generate",
		Short: "generator tool for various test suite assets",
	}
)

func NewCommand() *cobra.Command {
	generate.AddCommand(catalog.NewCommand())
	generate.AddCommand(feedback.NewCommand())
	generate.AddCommand(config.NewCommand())
	generate.AddCommand(qecoverage.NewCommand())

	return generate
}
