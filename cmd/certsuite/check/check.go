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

// NewCommand creates the root check command.
//
// It constructs a new cobra.Command instance configured for the certsuite
// checking functionality. The returned command is ready to have subcommands
// added to it and can be executed as part of the certsuite CLI. No arguments
// are required; the function returns a pointer to the created *cobra.Command.
func NewCommand() *cobra.Command {
	checkCmd.AddCommand(imagecert.NewCommand())
	checkCmd.AddCommand(results.NewCommand())

	return checkCmd
}
