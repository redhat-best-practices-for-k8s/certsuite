package run

import (
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/certsuite"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/webserver"
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
	runCmd.PersistentFlags().StringP("output-dir", "o", "cnf-certification-test", "The directory where the output artifacts will be placed")
	runCmd.PersistentFlags().StringP("label-filter", "l", "none", "Label expression to filter test cases  (e.g. --label-filter 'access-control && !access-control-sys-admin-capability')")
	runCmd.PersistentFlags().String("timeout", timeoutFlagDefaultvalue.String(), "Time allowed for the test suite execution to complete (e.g. --timeout 30m  or -timeout 1h30m)")
	runCmd.PersistentFlags().StringP("config-file", "c", "cnf-certification-test/tnf_config.yml", "Name of the workload configuration file")
	runCmd.PersistentFlags().StringP("kubeconfig", "k", "", "The target cluster's Kubeconfig file")
	runCmd.PersistentFlags().Bool("list", false, "Shows all the available checks/tests (can be filtered with --label-filter)")
	runCmd.PersistentFlags().Bool("server-mode", false, "Run the certsuite in web server mode")
	runCmd.PersistentFlags().Bool("omit-artifacts-zip-file", false, "Prevents the creation of a zip file with the result artifacts")
	runCmd.PersistentFlags().String("log-level", "debug", "Sets the log level")
	runCmd.PersistentFlags().String("offline-db", "", "Set the location of an offline DB to check the certification status of for container images, operators and helm charts")
	runCmd.PersistentFlags().String("preflight-dockerconfig", "", "Set the dockerconfig file to be used by the Preflight test suite")
	runCmd.PersistentFlags().Bool("non-intrusive", false, "Run only the test that do not disrupt the test environment")
	runCmd.PersistentFlags().Bool("allow-preflight-insecure", false, "Allow insecure connections in the Preflight test suite")
	runCmd.PersistentFlags().Bool("include-web-files", false, "Save web files in the configured output folder")
	runCmd.PersistentFlags().Bool("enable-data-collection", false, "Allow sending test results to an external data collector")
	runCmd.PersistentFlags().Bool("create-xml-junit-file", false, "Create a JUnit file with the test results")
	runCmd.PersistentFlags().String("tnf-image-repository", "quay.io/testnetworkfunction", "The repository where TNF images are stored")
	runCmd.PersistentFlags().String("tnf-debug-image", "debug-partner:5.2.0", "Name of the TNF debug image")
	runCmd.PersistentFlags().String("daemonset-cpu-req", "100m", "CPU request for the debug DaemonSet container")
	runCmd.PersistentFlags().String("daemonset-cpu-lim", "100m", "CPU limit for the debug DaemonSet container")
	runCmd.PersistentFlags().String("daemonset-mem-req", "100M", "Memory request for the debug DaemonSet container")
	runCmd.PersistentFlags().String("daemonset-mem-lim", "100M", "Memory limit for the debug DaemonSet container")

	return runCmd
}

func initTestParamsFromFlags(cmd *cobra.Command) error {
	testParams := configuration.GetTestParameters()

	// Fetch test params from flags
	testParams.OutputDir, _ = cmd.Flags().GetString("output-dir")
	testParams.LabelsFilter, _ = cmd.Flags().GetString("label-filter")
	testParams.ListOnly, _ = cmd.Flags().GetBool("list")
	testParams.ServerMode, _ = cmd.Flags().GetBool("server-mode")
	testParams.ConfigFile, _ = cmd.Flags().GetString("config-file")
	testParams.Kubeconfig, _ = cmd.Flags().GetString("kubeconfig")
	testParams.OmitArtifactsZipFile, _ = cmd.Flags().GetBool("omit-artifacts-zip-file")
	testParams.LogLevel, _ = cmd.Flags().GetString("log-level")
	testParams.OfflineDB, _ = cmd.Flags().GetString("offline-db")
	testParams.PfltDockerconfig, _ = cmd.Flags().GetString("preflight-dockerconfig")
	testParams.NonIntrusiveOnly, _ = cmd.Flags().GetBool("non-intrusive")
	testParams.AllowPreflightInsecure, _ = cmd.Flags().GetBool("allow-preflight-insecure")
	testParams.IncludeWebFilesInOutputFolder, _ = cmd.Flags().GetBool("include-web-files")
	testParams.EnableDataCollection, _ = cmd.Flags().GetBool("enable-data-collection")
	testParams.EnableXMLCreation, _ = cmd.Flags().GetBool("create-xml-junit-file")
	testParams.TnfImageRepo, _ = cmd.Flags().GetString("tnf-image-repository")
	testParams.TnfDebugImage, _ = cmd.Flags().GetString("tnf-debug-image")
	testParams.DaemonsetCPUReq, _ = cmd.Flags().GetString("daemonset-cpu-req")
	testParams.DaemonsetCPULim, _ = cmd.Flags().GetString("daemonset-cpu-lim")
	testParams.DaemonsetMemReq, _ = cmd.Flags().GetString("daemonset-mem-req")
	testParams.DaemonsetMemLim, _ = cmd.Flags().GetString("daemonset-mem-lim")
	timeoutStr, _ := cmd.Flags().GetString("timeout")

	// Check if the output directory exists and, if not, create it
	if _, err := os.Stat(testParams.OutputDir); os.IsNotExist(err) {
		var dirPerm fs.FileMode = 0o755 // default permissions for a directory
		err := os.MkdirAll(testParams.OutputDir, dirPerm)
		if err != nil {
			return fmt.Errorf("could not create directory %q, err: %v", testParams.OutputDir, err)
		}
	} else if err != nil {
		return fmt.Errorf("could not check directory %q, err: %v", testParams.OutputDir, err)
	}

	// Process the timeout flag
	const timeoutDefaultvalue = 24 * time.Hour
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse timeout flag %q, err: %v. Using default timeout value %v", timeoutStr, err, timeoutDefaultvalue)
		testParams.Timeout = timeoutDefaultvalue
	} else {
		testParams.Timeout = timeout
	}

	return nil
}
func runTestSuite(cmd *cobra.Command, _ []string) error {
	err := initTestParamsFromFlags(cmd)
	if err != nil {
		log.Fatal("Failed to initialize the test parameters, err: %v", err)
	}

	certsuite.Startup()
	defer certsuite.Shutdown()

	testParams := configuration.GetTestParameters()
	if testParams.ServerMode {
		log.Info("Running CNF Certification Suite in web server mode")
		webserver.StartServer(testParams.OutputDir)
	} else {
		log.Info("Running CNF Certification Suite in stand-alone mode")
		err := certsuite.Run(testParams.LabelsFilter, testParams.OutputDir)
		if err != nil {
			log.Fatal("Failed to run CNF Certification Suite: %v", err) //nolint:gocritic // exitAfterDefer
		}
	}

	return nil
}
