package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	claim "github.com/test-network-function/cnf-certification-test/cmd/tnf/addclaim"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/fetch"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/generate/catalog"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/grade"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tnf",
		Short: "A CLI for creating, validating , and test-network-function tests.",
	}

	generate = &cobra.Command{
		Use:   "generate",
		Short: "generator tool for various tnf artifacts.",
	}
)

func main() {
	rootCmd.AddCommand(claim.NewCommand())
	rootCmd.AddCommand(generate)
	rootCmd.AddCommand(fetch.NewCommand())
	generate.AddCommand(catalog.NewCommand())
	rootCmd.AddCommand(grade.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
