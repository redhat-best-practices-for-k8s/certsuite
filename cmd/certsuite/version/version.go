package version

import (
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/versions"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the Red Hat Best Practices Test Suite for Kubernetes version",
		RunE:  showVersion,
	}
)

// showVersion Displays the current application and claim file versions
//
// This function prints out two pieces of information: the version string for
// the Certsuite binary, which includes release and commit details, and the
// version number used for claim files. It formats both strings with newline
// separators and returns nil to indicate successful execution.
func showVersion(cmd *cobra.Command, _ []string) error {
	fmt.Printf("Certsuite version: %s\n", versions.GitVersion())
	fmt.Printf("Claim file version: %s\n", versions.ClaimFormatVersion)

	return nil
}

// NewCommand Provides the CLI command for displaying application version
//
// This function creates and returns a cobra command configured to show the
// current version of the tool when invoked. The command is set up elsewhere in
// the package, so this function simply exposes that preconfigured command
// instance for use by the main application.
func NewCommand() *cobra.Command {
	return versionCmd
}
