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

// showVersion prints the current build and Git commit information.
//
// It writes two lines to standard output: the first contains a friendly
// message with the binary name, the second shows the Git commit hash.
// The function returns an error if any printing operation fails.
func showVersion(cmd *cobra.Command, _ []string) error {
	fmt.Printf("Certsuite version: %s\n", versions.GitVersion())
	fmt.Printf("Claim file version: %s\n", versions.ClaimFormatVersion)

	return nil
}

// NewCommand creates the version subcommand.
//
// It constructs and returns a cobra.Command that displays the
// application version information when invoked. The returned
// command can be added to the root command of the certsuite CLI.
func NewCommand() *cobra.Command {
	return versionCmd
}
