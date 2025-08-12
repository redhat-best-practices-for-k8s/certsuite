package main

import (
	"os"

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

func newRootCmd() *cobra.Command {
	rootCmd := cobra.Command{
		Use:   "certsuite",
		Short: "A CLI tool for the Red Hat Best Practices Test Suite for Kubernetes.",
	}

	rootCmd.AddCommand(claim.NewCommand())
	rootCmd.AddCommand(generate.NewCommand())
	rootCmd.AddCommand(check.NewCommand())
	rootCmd.AddCommand(run.NewCommand())
	rootCmd.AddCommand(info.NewCommand())
	rootCmd.AddCommand(version.NewCommand())
	rootCmd.AddCommand(upload.NewCommand())

	return &rootCmd
}

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}
}
