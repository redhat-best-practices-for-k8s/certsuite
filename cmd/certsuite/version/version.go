package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/pkg/versions"
)

var (
	runCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the Red Hat Best Practices Test Suite for Kubernetes version",
		RunE:  showVersion,
	}
)

func showVersion(cmd *cobra.Command, _ []string) error {
	fmt.Printf("Certsuite version: %s\n", versions.GitVersion())
	fmt.Printf("Claim file version: %s\n", versions.ClaimFormatVersion)

	return nil
}

func NewCommand() *cobra.Command {
	return runCmd
}
