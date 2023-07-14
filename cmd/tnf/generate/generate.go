package generate

import (
	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/generate/catalog"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/generate/feedback"
	qecoverage "github.com/test-network-function/cnf-certification-test/cmd/tnf/generate/qe_coverage"
)

var (
	generate = &cobra.Command{
		Use:   "generate",
		Short: "generator tool for various tnf artifacts.",
	}
)

func NewCommand() *cobra.Command {
	generate.AddCommand(catalog.NewCommand())
	generate.AddCommand(feedback.NewCommand())
	generate.AddCommand(qecoverage.NewCommand())

	return generate
}
