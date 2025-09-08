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

// NewCommand Creates the upload command group for the CLI
//
// This function constructs a cobra.Command that represents the upload feature
// of the tool. It registers subcommands, such as those handling result
// spreadsheets, by adding them to the main upload command. The resulting
// command is returned for integration into the root command hierarchy.
func NewCommand() *cobra.Command {
	upload.AddCommand(resultsspreadsheet.NewCommand())

	return upload
}
