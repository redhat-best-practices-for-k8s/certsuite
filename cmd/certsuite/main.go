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

// newRootCmd Creates the top-level command for the certsuite CLI
//
// This function initializes a new root command with usage information and
// attaches subcommands such as claim, generate, check, run, info, version, and
// upload. Each subcommand is constructed by calling its own NewCommand
// function. The resulting command object is returned to be executed in the main
// entry point.
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

// main Runs the certsuite command-line interface
//
// It creates a root command with subcommands, executes it, and exits with an
// error code if execution fails. Errors are logged before terminating the
// program.
func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}
}
