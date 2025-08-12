package check

import (
	imagecert "github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/check/image_cert_status"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/check/results"
	"github.com/spf13/cobra"
)

var (
	checkCmd = &cobra.Command{
		Use:   "check",
		Short: "check the status of certsuite resources or artifacts.",
	}
)

func NewCommand() *cobra.Command {
	checkCmd.AddCommand(imagecert.NewCommand())
	checkCmd.AddCommand(results.NewCommand())

	return checkCmd
}
