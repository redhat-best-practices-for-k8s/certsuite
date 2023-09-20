package check

import (
	"github.com/spf13/cobra"
	imagecert "github.com/test-network-function/cnf-certification-test/cmd/tnf/check/image_cert_status"
)

var (
	checkCmd = &cobra.Command{
		Use:   "check",
		Short: "check the status of CNF resources.",
	}
)

func NewCommand() *cobra.Command {
	checkCmd.AddCommand(imagecert.NewCommand())

	return checkCmd
}
