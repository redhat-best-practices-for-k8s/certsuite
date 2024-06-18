package run

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/certsuite"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
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
	runCmd.PersistentFlags().String("tnf-debug-image", "debug-partner:5.1.2", "Name of the TNF debug image")

	return runCmd
}

func initFlags(arg interface{}) {
	cmd := arg.(*cobra.Command)

	outputDir, _ := cmd.Flags().GetString("output-dir")
	labelFilter, _ := cmd.Flags().GetString("label-filter")
	timeout, _ := cmd.Flags().GetString("timeout")
	list, _ := cmd.Flags().GetBool("list")
	serverMode, _ := cmd.Flags().GetBool("server-mode")
	configFile, _ := cmd.Flags().GetString("config-file")
	kubeconfigFile, _ := cmd.Flags().GetString("kubeconfig")
	omitZipFile, _ := cmd.Flags().GetBool("omit-artifacts-zip-file")
	logLevel, _ := cmd.Flags().GetString("log-level")
	offlineDB, _ := cmd.Flags().GetString("offline-db")
	pfltDockerconfig, _ := cmd.Flags().GetString("preflight-dockerconfig")
	nonIntrusive, _ := cmd.Flags().GetBool("non-intrusive")
	allowPfltInsecure, _ := cmd.Flags().GetBool("allow-preflight-insecure")
	includeWebFiles, _ := cmd.Flags().GetBool("include-web-files")
	dataCollection, _ := cmd.Flags().GetBool("enable-data-collection")
	createXMLJUnitFile, _ := cmd.Flags().GetBool("create-xml-junit-file")
	tnfImageRepo, _ := cmd.Flags().GetString("tnf-image-repository")
	tnfDebugImage, _ := cmd.Flags().GetString("tnf-debug-image")

	flags.OutputDir = &outputDir
	flags.LabelsFlag = &labelFilter
	flags.TimeoutFlag = &timeout
	flags.ListFlag = &list
	flags.ServerModeFlag = &serverMode
	flags.ConfigurationFile = configFile

	// Override env vars
	testParams := configuration.GetTestParameters()
	testParams.ConfigurationPath = configFile
	testParams.Kubeconfig = kubeconfigFile
	testParams.OmitArtifactsZipFile = omitZipFile
	testParams.LogLevel = logLevel
	testParams.OfflineDB = offlineDB
	testParams.PfltDockerconfig = pfltDockerconfig
	testParams.NonIntrusiveOnly = nonIntrusive
	testParams.AllowPreflightInsecure = allowPfltInsecure
	testParams.IncludeWebFilesInOutputFolder = includeWebFiles
	testParams.EnableDataCollection = dataCollection
	testParams.EnableXMLCreation = createXMLJUnitFile
	testParams.TnfPartnerRepo = tnfImageRepo
	testParams.SupportImage = tnfDebugImage
}
func runTestSuite(cmd *cobra.Command, _ []string) error {
	certsuite.Startup(initFlags, cmd)
	defer certsuite.Shutdown()

	err := certsuite.Run(*flags.LabelsFlag, *flags.OutputDir)
	if err != nil {
		log.Fatal("Failed to run CNF Certification Suite: %v", err) //nolint:gocritic // exitAfterDefer
	}

	return err
}
