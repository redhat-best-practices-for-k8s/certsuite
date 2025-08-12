package upload

import (
	resultsspreadsheet "github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/upload/results_spreadsheet"
	"github.com/spf13/cobra"
)

var (
	upload = &cobra.Command{
		Use:   "upload",
		Short: "upload tool for various test suite assets",
	}
)

// NewCommand creates the root upload command.
//
// It constructs a new cobra.Command configured to handle certificate
// uploads. The returned command can be added to other commands via
// AddCommand. No parameters are required; it returns a pointer to
// the configured cobra.Command instance.
func NewCommand() *cobra.Command {
	upload.AddCommand(resultsspreadsheet.NewCommand())

	return upload
}
