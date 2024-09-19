package main

import (
	"io"
	"os"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/info"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/version"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/versions"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCertsuiteVersionCmd(t *testing.T) {
	// Prepare context
	versions.GitCommit = "aaabbbccc"
	versions.GitRelease = "v0.0.0"
	versions.ClaimFormatVersion = "v0.0.0"

	// Run the command
	savedStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := createCommand(version.NewCommand())
	cmd.SetArgs([]string{
		"version",
	})
	err := cmd.Execute()
	assert.Nil(t, err)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = savedStdout

	// Check the result
	const expectedOutput = "Certsuite version: v0.0.0 (aaabbbccc)\nClaim file version: v0.0.0\n"
	assert.Equal(t, expectedOutput, string(out))
}

func TestCertsuiteInfoCmd(t *testing.T) {
	// Run the command
	savedStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := createCommand(info.NewCommand())
	cmd.SetArgs([]string{
		"info",
		"--test-label=observability",
		"--list",
	})
	err := cmd.Execute()
	assert.Nil(t, err)

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = savedStdout

	// Check the result
	const expectedOutput = `------------------------------------------------------------
|                   TEST CASE SELECTION                    |
------------------------------------------------------------
| observability-container-logging                          |
| observability-crd-status                                 |
| observability-termination-policy                         |
| observability-pod-disruption-budget                      |
| observability-compatibility-with-next-ocp-release        |
------------------------------------------------------------
`
	assert.Equal(t, expectedOutput, string(out))
}

func createCommand(cmd *cobra.Command) *cobra.Command {
	rootCmd := cobra.Command{
		Use:   "certsuite",
		Short: "A CLI tool for the Red Hat Best Practices Test Suite for Kubernetes.",
	}
	rootCmd.AddCommand(cmd)

	return &rootCmd
}
