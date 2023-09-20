package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/test-network-function/cnf-certification-test/cmd/tnf/check"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/generate"
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

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
