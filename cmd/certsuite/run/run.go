// Copyright (C) 2024-2026 Red Hat, Inc.
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
	runCmd.PersistentFlags().String("certsuite-probe-image", "quay.io/redhat-best-practices-for-k8s/certsuite-probe:v0.0.39", "Certsuite probe image")
	runCmd.PersistentFlags().String("daemonset-cpu-req", "100m", "CPU request for the probe daemonset container")
	runCmd.PersistentFlags().String("daemonset-cpu-lim", "100m", "CPU limit for the probe daemonset container")
	runCmd.PersistentFlags().String("daemonset-mem-req", "100M", "Memory request for the probe daemonset container")
	runCmd.PersistentFlags().String("daemonset-mem-lim", "100M", "Memory limit for the probe daemonset container")
	runCmd.PersistentFlags().Bool("sanitize-claim", false, "Sanitize the claim.json file before sending it to the collector")
	// Include non-Running pods during autodiscovery when enabled (default false)
	runCmd.PersistentFlags().Bool("allow-non-running", false, "Include non-Running pods during autodiscovery phase")
	runCmd.PersistentFlags().String("connect-api-key", "", "API Key for Red Hat Connect portal")
	runCmd.PersistentFlags().String("connect-project-id", "", "Project ID for Red Hat Connect portal")
	runCmd.PersistentFlags().String("connect-api-base-url", "", "Base URL for Red Hat Connect API")
	runCmd.PersistentFlags().String("connect-api-proxy-url", "", "Proxy URL for Red Hat Connect API")
	runCmd.PersistentFlags().String("connect-api-proxy-port", "", "Proxy port for Red Hat Connect API")
	runCmd.PersistentFlags().Bool("cleanup-probe", true, "Delete the probe daemonset at the end of the test run")
	runCmd.PersistentFlags().Bool("require-probe", false, "Abort the test run if the probe daemonset fails to deploy")

	return runCmd
}

type flagReader struct {
	cmd *cobra.Command
	err error
}

func (f *flagReader) getString(dest *string, name string) {
	if f.err != nil {
		return
	}
	*dest, f.err = f.cmd.Flags().GetString(name)
	if f.err != nil {
		f.err = fmt.Errorf("flag %q: %w", name, f.err)
	}
}

func (f *flagReader) getBool(dest *bool, name string) {
	if f.err != nil {
		return
	}
	*dest, f.err = f.cmd.Flags().GetBool(name)
	if f.err != nil {
		f.err = fmt.Errorf("flag %q: %w", name, f.err)
	}
}

func initTestParamsFromFlags(cmd *cobra.Command) error {
	testParams := configuration.GetTestParameters()
	f := &flagReader{cmd: cmd}

	f.getString(&testParams.OutputDir, "output-dir")
	f.getString(&testParams.LabelsFilter, "label-filter")
	f.getBool(&testParams.ServerMode, "server-mode")
	f.getString(&testParams.ConfigFile, "config-file")
	f.getString(&testParams.Kubeconfig, "kubeconfig")
	f.getBool(&testParams.OmitArtifactsZipFile, "omit-artifacts-zip-file")
	f.getString(&testParams.LogLevel, "log-level")
	f.getString(&testParams.OfflineDB, "offline-db")
	f.getString(&testParams.PfltDockerconfig, "preflight-dockerconfig")
	f.getBool(&testParams.Intrusive, "intrusive")
	f.getBool(&testParams.AllowPreflightInsecure, "allow-preflight-insecure")
	f.getBool(&testParams.IncludeWebFilesInOutputFolder, "include-web-files")
	f.getBool(&testParams.EnableDataCollection, "enable-data-collection")
	f.getBool(&testParams.EnableXMLCreation, "create-xml-junit-file")
	f.getString(&testParams.CertSuiteProbeImage, "certsuite-probe-image")
	f.getString(&testParams.DaemonsetCPUReq, "daemonset-cpu-req")
	f.getString(&testParams.DaemonsetCPULim, "daemonset-cpu-lim")
	f.getString(&testParams.DaemonsetMemReq, "daemonset-mem-req")
	f.getString(&testParams.DaemonsetMemLim, "daemonset-mem-lim")
	f.getBool(&testParams.SanitizeClaim, "sanitize-claim")
	f.getBool(&testParams.AllowNonRunning, "allow-non-running")
	f.getString(&testParams.ConnectAPIKey, "connect-api-key")
	f.getString(&testParams.ConnectProjectID, "connect-project-id")
	f.getString(&testParams.ConnectAPIBaseURL, "connect-api-base-url")
	f.getString(&testParams.ConnectAPIProxyURL, "connect-api-proxy-url")
	f.getString(&testParams.ConnectAPIProxyPort, "connect-api-proxy-port")
	f.getBool(&testParams.CleanupProbe, "cleanup-probe")
	f.getBool(&testParams.RequireProbe, "require-probe")

	var timeoutStr string
	f.getString(&timeoutStr, "timeout")
	if f.err != nil {
		return f.err
	}

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
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse timeout flag %q, err: %v. Using default timeout value %v", timeoutStr, err, timeoutFlagDefaultvalue)
		testParams.Timeout = timeoutFlagDefaultvalue
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
		if err := webserver.StartServer(testParams.OutputDir); err != nil {
			log.Fatal("Failed to start web server: %v", err)
		}
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
