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

// NewCommand Creates a check command that aggregates image certification and result verification actions
//
// This function builds a new Cobra command for the tool’s check
// functionality. It registers two child commands: one to verify image
// certificates and another to validate test results against expected outputs or
// logs. The resulting command is returned for inclusion in the main CLI
// hierarchy.
func NewCommand() *cobra.Command {
	checkCmd.AddCommand(imagecert.NewCommand())
	checkCmd.AddCommand(results.NewCommand())

	return checkCmd
}
