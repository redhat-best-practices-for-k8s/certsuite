package generate

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/catalog"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/config"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/feedback"
	qecoverage "github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate/qe_coverage"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/cli"
	"github.com/spf13/cobra"
)

var (
	generate = &cobra.Command{
		Use:   "generate",
		Short: "generator tool for various test suite assets",
	}
)

func NewCommand() *cobra.Command {
	cli.AddCommand(generate, catalog.NewCommand())
	cli.AddCommand(generate, feedback.NewCommand())
	cli.AddCommand(generate, config.NewCommand())
	cli.AddCommand(generate, qecoverage.NewCommand())

	return generate
}
