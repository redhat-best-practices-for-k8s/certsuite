package run

import (
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/certsuite"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/redhat-best-practices-for-k8s/certsuite/webserver"
	"github.com/spf13/cobra"
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
	runCmd.PersistentFlags().StringP("output-dir", "o", "results", "The directory where the output artifacts will be placed")
	runCmd.PersistentFlags().StringP("label-filter", "l", "none", "Label expression to filter test cases  (e.g. --label-filter 'access-control && !access-control-sys-admin-capability')")
	runCmd.PersistentFlags().String("timeout", timeoutFlagDefaultvalue.String(), "Time allowed for the test suite execution to complete (e.g. --timeout 30m  or -timeout 1h30m)")
	runCmd.PersistentFlags().StringP("config-file", "c", "config/certsuite_config.yml", "The certsuite configuration file")
	runCmd.PersistentFlags().StringP("kubeconfig", "k", "", "The target cluster's Kubeconfig file")
	runCmd.PersistentFlags().Bool("server-mode", false, "Run the certsuite in web server mode")
	runCmd.PersistentFlags().Bool("omit-artifacts-zip-file", false, "Prevents the creation of a zip file with the result artifacts")
	runCmd.PersistentFlags().String("log-level", "debug", "Sets the log level")
	runCmd.PersistentFlags().String("offline-db", "", "Set the location of an offline DB to check the certification status of for container images, operators and helm charts")
	runCmd.PersistentFlags().String("preflight-dockerconfig", "", "Set the dockerconfig file to be used by the Preflight test suite")
	runCmd.PersistentFlags().Bool("intrusive", true, "Run intrusive tests that may disrupt the test environment")
	runCmd.PersistentFlags().Bool("allow-preflight-insecure", false, "Allow insecure connections in the Preflight test suite")
	runCmd.PersistentFlags().Bool("include-web-files", false, "Save web files in the configured output folder")
	runCmd.PersistentFlags().Bool("enable-data-collection", false, "Allow sending test results to an external data collector")
	runCmd.PersistentFlags().Bool("create-xml-junit-file", false, "Create a JUnit file with the test results")
	runCmd.PersistentFlags().String("certsuite-probe-image", "quay.io/redhat-best-practices-for-k8s/certsuite-probe:v0.0.20", "Certsuite probe image")
	runCmd.PersistentFlags().String("daemonset-cpu-req", "100m", "CPU request for the probe daemonset container")
	runCmd.PersistentFlags().String("daemonset-cpu-lim", "100m", "CPU limit for the probe daemonset container")
	runCmd.PersistentFlags().String("daemonset-mem-req", "100M", "Memory request for the probe daemonset container")
	runCmd.PersistentFlags().String("daemonset-mem-lim", "100M", "Memory limit for the probe daemonset container")
	runCmd.PersistentFlags().Bool("sanitize-claim", false, "Sanitize the claim.json file before sending it to the collector")
	runCmd.PersistentFlags().String("connect-api-key", "", "API Key for Red Hat Connect portal")
	runCmd.PersistentFlags().String("connect-project-id", "", "Project ID for Red Hat Connect portal")
	runCmd.PersistentFlags().String("connect-api-base-url", "", "Base URL for Red Hat Connect API")
	runCmd.PersistentFlags().String("connect-api-proxy-url", "", "Proxy URL for Red Hat Connect API")
	runCmd.PersistentFlags().String("connect-api-proxy-port", "", "Proxy port for Red Hat Connect API")

	return runCmd
}

func initTestParamsFromFlags(cmd *cobra.Command) error {
	testParams := configuration.GetTestParameters()

	// Fetch test params from flags
	testParams.OutputDir, _ = cmd.Flags().GetString("output-dir")
	testParams.LabelsFilter, _ = cmd.Flags().GetString("label-filter")
	testParams.ServerMode, _ = cmd.Flags().GetBool("server-mode")
	testParams.ConfigFile, _ = cmd.Flags().GetString("config-file")
	testParams.Kubeconfig, _ = cmd.Flags().GetString("kubeconfig")
	testParams.OmitArtifactsZipFile, _ = cmd.Flags().GetBool("omit-artifacts-zip-file")
	testParams.LogLevel, _ = cmd.Flags().GetString("log-level")
	testParams.OfflineDB, _ = cmd.Flags().GetString("offline-db")
	testParams.PfltDockerconfig, _ = cmd.Flags().GetString("preflight-dockerconfig")
	testParams.Intrusive, _ = cmd.Flags().GetBool("intrusive")
	testParams.AllowPreflightInsecure, _ = cmd.Flags().GetBool("allow-preflight-insecure")
	testParams.IncludeWebFilesInOutputFolder, _ = cmd.Flags().GetBool("include-web-files")
	testParams.EnableDataCollection, _ = cmd.Flags().GetBool("enable-data-collection")
	testParams.EnableXMLCreation, _ = cmd.Flags().GetBool("create-xml-junit-file")
	testParams.CertSuiteProbeImage, _ = cmd.Flags().GetString("certsuite-probe-image")
	testParams.DaemonsetCPUReq, _ = cmd.Flags().GetString("daemonset-cpu-req")
	testParams.DaemonsetCPULim, _ = cmd.Flags().GetString("daemonset-cpu-lim")
	testParams.DaemonsetMemReq, _ = cmd.Flags().GetString("daemonset-mem-req")
	testParams.DaemonsetMemLim, _ = cmd.Flags().GetString("daemonset-mem-lim")
	testParams.SanitizeClaim, _ = cmd.Flags().GetBool("sanitize-claim")
	testParams.ConnectAPIKey, _ = cmd.Flags().GetString("connect-api-key")
	testParams.ConnectProjectID, _ = cmd.Flags().GetString("connect-project-id")
	testParams.ConnectAPIBaseURL, _ = cmd.Flags().GetString("connect-api-base-url")
	testParams.ConnectAPIProxyURL, _ = cmd.Flags().GetString("connect-api-proxy-url")
	testParams.ConnectAPIProxyPort, _ = cmd.Flags().GetString("connect-api-proxy-port")
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

	testParams := configuration.GetTestParameters()
	if testParams.ServerMode {
		log.Info("Running Certification Suite in web server mode")
		webserver.StartServer(testParams.OutputDir)
	} else {
		certsuite.Startup()
		defer certsuite.Shutdown()
		log.Info("Running Certification Suite in stand-alone mode")
		err := certsuite.Run(testParams.LabelsFilter, testParams.OutputDir)
		if err != nil {
			log.Fatal("Failed to run Certification Suite: %v", err) //nolint:gocritic // exitAfterDefer
		}
	}

	return nil
}
