package check

import (
	"github.com/spf13/cobra"
	imagecert "github.com/test-network-function/cnf-certification-test/cmd/tnf/check/image_cert_status"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/check/results"
)

var (
	checkCmd = &cobra.Command{
		Use:   "check",
		Short: "check the status of CNF resources or artifacts.",
	}
)

func NewCommand() *cobra.Command {
	checkCmd.AddCommand(imagecert.NewCommand())
	checkCmd.AddCommand(results.NewCommand())

	return checkCmd
}
