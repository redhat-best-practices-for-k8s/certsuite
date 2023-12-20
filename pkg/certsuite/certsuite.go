package certsuite

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/claimhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/collector"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/flags"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

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
	junitXMLOutputFile = "cnf-certification-tests_junit.xml"
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

func processFlags() time.Duration {
	_ = clientsholder.GetClientsHolder(getK8sClientsConfigFileNames()...)

	LoadChecksDB(*flags.LabelsFlag)

	// Diagnostic functions will run when no labels are provided.
	if *flags.LabelsFlag == flags.NoLabelsExpr {
		log.Warn("CNF Certification Suite will run in diagnostic mode so no test case will be launched.")
	}

	timeout, err := time.ParseDuration(*flags.TimeoutFlag)
	if err != nil {
		log.Error("Failed to parse timeout flag %v: %v, using default timeout value %v", *flags.TimeoutFlag, err, flags.TimeoutFlagDefaultvalue)
		timeout = flags.TimeoutFlagDefaultvalue
	}
	return timeout
}

//nolint:funlen
func Run(labelsFilter, outputFolder string) error {
	timeout := processFlags()
	var returnErr bool

	fmt.Println("Running discovery of CNF target resources...")
	fmt.Print("\n")
	var env provider.TestEnvironment
	env.SetNeedsRefresh()
	env = provider.GetTestEnvironment()

	claimBuilder, err := claimhelper.NewClaimBuilder()
	if err != nil {
		log.Error("Failed to get claim builder: %v", err)
		os.Exit(1)
	}

	claimOutputFile := filepath.Join(outputFolder, results.ClaimFileName)

	log.Info("Running checks matching labels expr %q with timeout %v", labelsFilter, timeout)
	startTime := time.Now()
	failedCtr, err := checksdb.RunChecks(labelsFilter, timeout)
	if err != nil {
		log.Error("%v", err)
	}
	endTime := time.Now()
	log.Info("Finished running checks in %v", endTime.Sub(startTime))

	if failedCtr > 0 {
		log.Error("Some checks failed. See %s for details", claimOutputFile)
		returnErr = true
	}

	// Marshal the claim and output to file
	claimBuilder.Build(claimOutputFile)

	if configuration.GetTestParameters().EnableXMLCreation {
		log.Info("XML file creation is enabled. Creating JUnit XML file: %s", junitXMLOutputFile)
		claimBuilder.ToJUnitXML(junitXMLOutputFile, startTime, endTime)
	}

	// Send claim file to the collector if specified by env var
	if configuration.GetTestParameters().EnableDataCollection {
		err = collector.SendClaimFileToCollector(env.CollectorAppEndPoint, claimOutputFile, env.ExecutedBy, env.PartnerName, env.CollectorAppPassword)
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
			log.Error("Failed to compress results artifacts: %v", err)
			os.Exit(1)
		}
	}

	// Remove web artifacts if user does not want them.
	if !configuration.GetTestParameters().IncludeWebFilesInOutputFolder {
		for _, file := range webFilePaths {
			err := os.Remove(file)
			if err != nil {
				log.Error("failed to remove web file %s: %v", file, err)
				os.Exit(1)
			}
		}
	}

	if returnErr {
		return errors.New("some checks failed")
	}

	return nil
}
