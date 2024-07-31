package certsuite

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/test-network-function/cnf-certification-test/internal/cli"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/internal/results"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/claimhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/collector"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/versions"
	"github.com/test-network-function/cnf-certification-test/tests/accesscontrol"
	"github.com/test-network-function/cnf-certification-test/tests/certification"
	"github.com/test-network-function/cnf-certification-test/tests/lifecycle"
	"github.com/test-network-function/cnf-certification-test/tests/manageability"
	"github.com/test-network-function/cnf-certification-test/tests/networking"
	"github.com/test-network-function/cnf-certification-test/tests/observability"
	"github.com/test-network-function/cnf-certification-test/tests/operator"
	"github.com/test-network-function/cnf-certification-test/tests/performance"
	"github.com/test-network-function/cnf-certification-test/tests/platform"
	"github.com/test-network-function/cnf-certification-test/tests/preflight"
)

func LoadInternalChecksDB() {
	accesscontrol.LoadChecks()
	certification.LoadChecks()
	lifecycle.LoadChecks()
	manageability.LoadChecks()
	networking.LoadChecks()
	observability.LoadChecks()
	performance.LoadChecks()
	platform.LoadChecks()
	operator.LoadChecks()
}

func LoadChecksDB(labelsExpr string) {
	LoadInternalChecksDB()

	if preflight.ShouldRun(labelsExpr) {
		preflight.LoadChecks()
	}
}

const (
	junitXMLOutputFileName = "cnf-certification-tests_junit.xml"
	claimFileName          = "claim.json"
	collectorAppURL        = "http://claims-collector.cnf-certifications.sysdeseng.com"
	timeoutDefaultvalue    = 24 * time.Hour
	noLabelsFilterExpr     = "none"
)

func getK8sClientsConfigFileNames() []string {
	params := configuration.GetTestParameters()
	fileNames := []string{}
	if params.Kubeconfig != "" {
		// Add the kubeconfig path
		fileNames = append(fileNames, params.Kubeconfig)
	}
	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		kubeConfigFilePath := filepath.Join(homeDir, ".kube", "config")
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

func Startup() {
	testParams := configuration.GetTestParameters()

	// Create an evaluator to filter test cases with labels
	if err := checksdb.InitLabelsExprEvaluator(testParams.LabelsFilter); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize a test case label evaluator, err: %v", err)
		os.Exit(1)
	}

	if err := log.CreateGlobalLogFile(testParams.OutputDir, testParams.LogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create the log file, err: %v\n", err)
		os.Exit(1)
	}

	// Diagnostic functions will run when no labels are provided.
	if testParams.LabelsFilter == noLabelsFilterExpr {
		log.Warn("The Best Practices Test Suite will run in diagnostic mode so no test case will be launched")
	}

	// Set clientsholder singleton with the filenames from the env vars.
	_ = clientsholder.GetClientsHolder(getK8sClientsConfigFileNames()...)
	LoadChecksDB(testParams.LabelsFilter)

	log.Info("Certsuite Version: %v", versions.GitVersion())
	log.Info("Claim Format Version: %s", versions.ClaimFormatVersion)
	log.Info("Labels filter: %v", testParams.LabelsFilter)
	log.Info("Log level: %s", strings.ToUpper(testParams.LogLevel))

	log.Debug("Test parameters: %#v", *configuration.GetTestParameters())

	cli.PrintBanner()

	fmt.Printf("Certsuite version: %s\n", versions.GitVersion())
	fmt.Printf("Claim file version: %s\n", versions.ClaimFormatVersion)
	fmt.Printf("Checks filter: %s\n", testParams.LabelsFilter)
	fmt.Printf("Output folder: %s\n", testParams.OutputDir)
	fmt.Printf("Log file: %s (level=%s)\n", log.LogFileName, testParams.LogLevel)
	fmt.Printf("\n")
}

func Shutdown() {
	err := log.CloseGlobalLogFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not close the log file, err: %v\n", err)
		os.Exit(1)
	}
}

//nolint:funlen
func Run(labelsFilter, outputFolder string) error {
	testParams := configuration.GetTestParameters()

	fmt.Println("Running discovery of CNF target resources...")
	fmt.Print("\n")

	env := provider.GetTestEnvironment()

	claimBuilder, err := claimhelper.NewClaimBuilder()
	if err != nil {
		log.Fatal("Failed to get claim builder: %v", err)
	}

	claimOutputFile := filepath.Join(outputFolder, claimFileName)

	log.Info("Running checks matching labels expr %q with timeout %v", labelsFilter, testParams.Timeout)
	startTime := time.Now()
	failedCtr, err := checksdb.RunChecks(testParams.Timeout)
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

	if configuration.GetTestParameters().SanitizeClaim {
		claimOutputFile, err = claimhelper.SanitizeClaimFile(claimOutputFile, configuration.GetTestParameters().LabelsFilter)
		if err != nil {
			log.Error("Failed to sanitize claim file: %v", err)
		}
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
	webFilePaths, err := results.CreateResultsWebFiles(resultsOutputDir, claimFileName)
	if err != nil {
		log.Error("Failed to create results web files: %v", err)
	}

	allArtifactsFilePaths := []string{filepath.Join(outputFolder, claimFileName)}

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
