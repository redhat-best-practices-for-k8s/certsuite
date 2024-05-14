package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/internal/log"

	"github.com/test-network-function/cnf-certification-test/cmd/tnf/check"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/generate"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/run"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tnf",
		Short: "A CLI program with tools related to the CNF Certification Suite.",
	}
)

func main() {
	rootCmd.AddCommand(claim.NewCommand())
	rootCmd.AddCommand(generate.NewCommand())
	rootCmd.AddCommand(check.NewCommand())
	rootCmd.AddCommand(run.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}
}
