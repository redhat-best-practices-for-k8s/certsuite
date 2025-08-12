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

func showVersion(cmd *cobra.Command, _ []string) error {
	fmt.Printf("Certsuite version: %s\n", versions.GitVersion())
	fmt.Printf("Claim file version: %s\n", versions.ClaimFormatVersion)

	return nil
}

func NewCommand() *cobra.Command {
	return versionCmd
}
