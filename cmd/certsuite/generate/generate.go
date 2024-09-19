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

func NewCommand() *cobra.Command {
	generate.AddCommand(catalog.NewCommand())
	generate.AddCommand(feedback.NewCommand())
	generate.AddCommand(config.NewCommand())
	generate.AddCommand(qecoverage.NewCommand())

	return generate
}
