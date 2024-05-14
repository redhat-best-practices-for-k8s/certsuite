package run

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/certsuite"
	"github.com/test-network-function/cnf-certification-test/pkg/flags"
)

const timeoutFlagDefaultvalue = 24 * time.Hour

var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the Red Hat Best Practices Test Suite for Kubernetes",
		RunE:  runTestSuite,
	}
)

func NewCommand() *cobra.Command {
	runCmd.PersistentFlags().String("output-dir", "cnf-certification-test", "The directory where the output artifacts will be placed")
	runCmd.PersistentFlags().String("label-filter", "none", "Label expression to filter test cases  (e.g. --label-filter 'access-control && !access-control-sys-admin-capability')")
	runCmd.PersistentFlags().String("timeout", timeoutFlagDefaultvalue.String(), "Time allowed for the test suite execution to complete (e.g. --timeout 30m  or -timeout 1h30m)")
	runCmd.PersistentFlags().String("config-file", "cnf-certification-test/tnf_config.yml", "Name of the workload configuration file")
	runCmd.PersistentFlags().Bool("list", false, "Shows all the available checks/tests. Can be filtered with --label-filter.")
	runCmd.PersistentFlags().Bool("server-mode", false, "Run the certsuite in web server mode.")

	return runCmd
}

func initFlags(cmd *cobra.Command) {
	outputDir, _ := cmd.Flags().GetString("output-dir")
	labelFilter, _ := cmd.Flags().GetString("label-filter")
	timeout, _ := cmd.Flags().GetString("timeout")
	list, _ := cmd.Flags().GetBool("list")
	serverMode, _ := cmd.Flags().GetBool("server-mode")
	configFile, _ := cmd.Flags().GetString("config-file")

	flags.OutputDir = &outputDir
	flags.LabelsFlag = &labelFilter
	flags.TimeoutFlag = &timeout
	flags.ListFlag = &list
	flags.ServerModeFlag = &serverMode
	flags.ConfigurationFile = configFile
}
func runTestSuite(cmd *cobra.Command, _ []string) error {
	initFlags(cmd)

	certsuite.Startup()
	defer certsuite.Shutdown()

	err := certsuite.Run(*flags.LabelsFlag, *flags.OutputDir)
	if err != nil {
		log.Fatal("Failed to run CNF Certification Suite: %v", err) //nolint:gocritic // exitAfterDefer
	}

	return err
}
