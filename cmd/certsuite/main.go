package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/internal/log"

	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/check"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/claim"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/generate"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/run"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/version"
)

var (
	rootCmd = &cobra.Command{
		Use:   "certsuite",
		Short: "A CLI tool for the Red Hat Best Practices Test Suite for Kubernetes.",
	}
)

func main() {
	rootCmd.AddCommand(claim.NewCommand())
	rootCmd.AddCommand(generate.NewCommand())
	rootCmd.AddCommand(check.NewCommand())
	rootCmd.AddCommand(run.NewCommand())
	rootCmd.AddCommand(version.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}
}
