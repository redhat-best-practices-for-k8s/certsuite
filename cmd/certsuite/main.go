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

// newRootCmd creates the top-level command for the certsuite application.
//
// It constructs a new cobra.Command, sets its usage and description,
// and registers all of the sub‑commands that make up the certsuite CLI.
// The function returns a pointer to this root command so it can be
// passed to Execute or used in tests.
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

// main initializes and executes the CertSuite CLI.
//
// It creates the root command, invokes its Execute method to run the
// application based on user input, handles any errors by printing them,
// and exits with a non‑zero status if execution fails.
func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}
}
