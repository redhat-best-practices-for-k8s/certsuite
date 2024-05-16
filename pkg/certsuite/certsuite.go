package certsuite

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/certification"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/manageability"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/observability"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/operator"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/performance"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/preflight"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/cli"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/claimhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/collector"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/flags"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/versions"
)

var timeout time.Duration

func LoadChecksDB(labelsExpr string) {
	accesscontrol.LoadChecks()
	certification.LoadChecks()
	lifecycle.LoadChecks()
	manageability.LoadChecks()
	networking.LoadChecks()
	observability.LoadChecks()
	performance.LoadChecks()
	platform.LoadChecks()
	operator.LoadChecks()

	if preflight.ShouldRun(labelsExpr) {
		preflight.LoadChecks()
	}
}

const (
	junitXMLOutputFileName = "cnf-certification-tests_junit.xml"
	collectorAppURL        = "http://claims-collector.cnf-certifications.sysdeseng.com"
)

func getK8sClientsConfigFileNames() []string {
	params := configuration.GetTestParameters()
	fileNames := []string{}
	if params.Kubeconfig != "" {
		// Add the kubeconfig path
		fileNames = append(fileNames, params.Kubeconfig)
	}
	if params.Home != "" {
		kubeConfigFilePath := filepath.Join(params.Home, ".kube", "config")
		// Check if the kubeconfig path exists
		if _, err := os.Stat(kubeConfigFilePath); err == nil {
			log.Info("kubeconfig path %s is present", kubeConfigFilePath)
			// Only add the kubeconfig to the list of paths if it exists, since it is not added by the user
			fileNames = append(fileNames, kubeConfigFilePath)
		} else {
			log.Info("kubeconfig path %s is not present", kubeConfigFilePath)
		}
	}

	return fileNames
}

func processFlags() {
	// Diagnostic functions will run when no labels are provided.
	if *flags.LabelsFlag == flags.NoLabelsExpr {
		log.Warn("CNF Certification Suite will run in diagnostic mode so no test case will be launched")
	}

	// If the list flag is passed, print the checks filtered with --labels and leave
	if *flags.ListFlag {
		checksIDs, err := checksdb.FilterCheckIDs()
		if err != nil {
			log.Fatal("Could not list test cases, err: %v", err)
		} else {
			cli.PrintChecksList(checksIDs)
			os.Exit(0)
		}
	}

	t, err := time.ParseDuration(*flags.TimeoutFlag)
	if err != nil {
		log.Error("Failed to parse timeout flag %q, err: %v, using default timeout value %v", *flags.TimeoutFlag, err, flags.TimeoutFlagDefaultvalue)
		timeout = flags.TimeoutFlagDefaultvalue
	} else {
		timeout = t
	}
}

func Startup(initFlags bool) {
	if initFlags {
		flags.InitFlags()
	}

	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load the environment variables, err: %v", err)
		os.Exit(1)
	}

	logLevel := configuration.GetTestParameters().LogLevel
	err = log.CreateGlobalLogFile(*flags.OutputDir, logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create the log file, err: %v", err)
		os.Exit(1)
	}

	log.Info("TNF Version: %v", versions.GitVersion())
	log.Info("Claim Format Version: %s", versions.ClaimFormatVersion)
	log.Info("Labels filter: %v", *flags.LabelsFlag)
	log.Info("Log level: %s", strings.ToUpper(logLevel))

	log.Debug("Test environment variables: %#v", *configuration.GetTestParameters())

	cli.PrintBanner()

	fmt.Printf("CNFCERT version: %s\n", versions.GitVersion())
	fmt.Printf("Claim file version: %s\n", versions.ClaimFormatVersion)
	fmt.Printf("Checks filter: %s\n", *flags.LabelsFlag)
	fmt.Printf("Output folder: %s\n", *flags.OutputDir)
	fmt.Printf("Log file: %s (level=%s)\n", log.LogFileName, logLevel)
	fmt.Printf("\n")
}

func Shutdown() {
	err := log.CloseGlobalLogFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not close the log file, err: %v", err)
		os.Exit(1)
	}
}

//nolint:funlen
func Run(labelsFilter, outputFolder string) error {
	// Set clientsholder singleton with the filenames from the env vars.
	_ = clientsholder.GetClientsHolder(getK8sClientsConfigFileNames()...)
	LoadChecksDB(*flags.LabelsFlag)

	processFlags()

	// Create an evaluator to filter test cases with labels
	if err := checksdb.InitLabelsExprEvaluator(labelsFilter); err != nil {
		return fmt.Errorf("failed to initialize a test case label evaluator, err: %v", err)
	}

	fmt.Println("Running discovery of CNF target resources...")
	fmt.Print("\n")

	env := provider.GetTestEnvironment()

	claimBuilder, err := claimhelper.NewClaimBuilder()
	if err != nil {
		log.Fatal("Failed to get claim builder: %v", err)
	}

	claimOutputFile := filepath.Join(outputFolder, results.ClaimFileName)

	log.Info("Running checks matching labels expr %q with timeout %v", labelsFilter, timeout)
	startTime := time.Now()
	failedCtr, err := checksdb.RunChecks(timeout)
	if err != nil {
		log.Error("%v", err)
	}
	endTime := time.Now()
	log.Info("Finished running checks in %v", endTime.Sub(startTime))

	if failedCtr > 0 {
		log.Warn("Some checks failed. See %s for details", claimOutputFile)
	}

	// Marshal the claim and output to file
	claimBuilder.Build(claimOutputFile)

	// Create JUnit file if required
	if configuration.GetTestParameters().EnableXMLCreation {
		junitOutputFileName := filepath.Join(outputFolder, junitXMLOutputFileName)
		log.Info("JUnit XML file creation is enabled. Creating JUnit XML file: %s", junitOutputFileName)
		claimBuilder.ToJUnitXML(junitOutputFileName, startTime, endTime)
	}

	// Send claim file to the collector if specified by env var
	if configuration.GetTestParameters().EnableDataCollection {
		if env.CollectorAppEndpoint == "" {
			env.CollectorAppEndpoint = collectorAppURL
		}
		err = collector.SendClaimFileToCollector(env.CollectorAppEndpoint, claimOutputFile, env.ExecutedBy, env.PartnerName, env.CollectorAppPassword)
		if err != nil {
			log.Error("Failed to send post request to the collector: %v", err)
		}
	}

	// Create HTML artifacts for the web results viewer/parser.
	resultsOutputDir := outputFolder
	webFilePaths, err := results.CreateResultsWebFiles(resultsOutputDir)
	if err != nil {
		log.Error("Failed to create results web files: %v", err)
	}

	allArtifactsFilePaths := []string{filepath.Join(outputFolder, results.ClaimFileName)}

	// Add all the web artifacts file paths.
	allArtifactsFilePaths = append(allArtifactsFilePaths, webFilePaths...)

	// Add the log file path
	allArtifactsFilePaths = append(allArtifactsFilePaths, filepath.Join(outputFolder, log.LogFileName))

	// tar.gz file creation with results and html artifacts, unless omitted by env var.
	if !configuration.GetTestParameters().OmitArtifactsZipFile {
		err = results.CompressResultsArtifacts(resultsOutputDir, allArtifactsFilePaths)
		if err != nil {
			log.Fatal("Failed to compress results artifacts: %v", err)
		}
	}

	// Remove web artifacts if user does not want them.
	if !configuration.GetTestParameters().IncludeWebFilesInOutputFolder {
		for _, file := range webFilePaths {
			err := os.Remove(file)
			if err != nil {
				log.Fatal("Failed to remove web file %s: %v", file, err)
			}
		}
	}

	return nil
}
