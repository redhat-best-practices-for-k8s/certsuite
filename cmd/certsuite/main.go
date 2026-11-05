package main

import (
	"os"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/cli"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/spf13/cobra"

	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/check"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/generate"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/info"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/run"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/upload"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/version"
)

const (
	rootCmdUse   = "certsuite"
	rootCmdShort = "A CLI tool for the Red Hat Best Practices Test Suite for Kubernetes."
)

func newRootCmd() *cobra.Command {
	rootCmd := cobra.Command{
		Use:   rootCmdUse,
		Short: rootCmdShort,
	}

	cli.AddCommand(&rootCmd, claim.NewCommand())
	cli.AddCommand(&rootCmd, generate.NewCommand())
	cli.AddCommand(&rootCmd, check.NewCommand())
	cli.AddCommand(&rootCmd, run.NewCommand())
	cli.AddCommand(&rootCmd, info.NewCommand())
	cli.AddCommand(&rootCmd, version.NewCommand())
	cli.AddCommand(&rootCmd, upload.NewCommand())

	return &rootCmd
}

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}
}
