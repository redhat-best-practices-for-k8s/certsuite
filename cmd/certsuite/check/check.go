package check

import (
	imagecert "github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/check/image_cert_status"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/check/results"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/cli"
	"github.com/spf13/cobra"
)

var (
	checkCmd = &cobra.Command{
		Use:   "check",
		Short: "check the status of certsuite resources or artifacts.",
	}
)

func NewCommand() *cobra.Command {
	cli.AddCommand(checkCmd, imagecert.NewCommand())
	cli.AddCommand(checkCmd, results.NewCommand())

	return checkCmd
}
