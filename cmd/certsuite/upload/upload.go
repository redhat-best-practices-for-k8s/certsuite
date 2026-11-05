package upload

import (
	resultsspreadsheet "github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/upload/results_spreadsheet"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/cli"
	"github.com/spf13/cobra"
)

var (
	upload = &cobra.Command{
		Use:   "upload",
		Short: "upload tool for various test suite assets",
	}
)

func NewCommand() *cobra.Command {
	cli.AddCommand(upload, resultsspreadsheet.NewCommand())

	return upload
}
