package certsuite

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/certification"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/manageability"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/observability"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/performance"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/preflight"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/claimhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/collector"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
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

	if preflight.ShouldRun(labelsExpr) {
		preflight.LoadChecks()
	}
}

func Run(labelsFilter, outputFolder string, timeout time.Duration) {
	var env provider.TestEnvironment
	env.SetNeedsRefresh()
	env = provider.GetTestEnvironment()

	claimBuilder, err := claimhelper.NewClaimBuilder()
	if err != nil {
		logrus.Fatalf("Failed to get claim builder: %v", err)
	}

	claimOutputFile := filepath.Join(outputFolder, results.ClaimFileName)

	logrus.Infof("Running checks matching labels expr %q with timeout %v", labelsFilter, timeout)
	err = checksdb.RunChecks(labelsFilter, timeout)
	if err != nil {
		logrus.Error(err)
	}

	// Marshal the claim and output to file
	claimBuilder.Build(claimOutputFile)

	// Send claim file to the collector if specified by env var
	if configuration.GetTestParameters().EnableDataCollection {
		err = collector.SendClaimFileToCollector(env.CollectorAppEndPoint, claimOutputFile, env.ExecutedBy, env.PartnerName, env.CollectorAppPassword)
		if err != nil {
			logrus.Errorf("Failed to send post request to the collector: %v", err)
		}
	}

	// Create HTML artifacts for the web results viewer/parser.
	resultsOutputDir := outputFolder
	webFilePaths, err := results.CreateResultsWebFiles(resultsOutputDir)
	if err != nil {
		logrus.Errorf("Failed to create results web files: %v", err)
	}

	allArtifactsFilePaths := []string{filepath.Join(outputFolder, results.ClaimFileName)}

	// Add all the web artifacts file paths.
	allArtifactsFilePaths = append(allArtifactsFilePaths, webFilePaths...)

	// tar.gz file creation with results and html artifacts, unless omitted by env var.
	if !configuration.GetTestParameters().OmitArtifactsZipFile {
		err = results.CompressResultsArtifacts(resultsOutputDir, allArtifactsFilePaths)
		if err != nil {
			logrus.Fatalf("Failed to compress results artifacts: %v", err)
		}
	}

	// Remove web artifacts if user does not want them.
	if !configuration.GetTestParameters().IncludeWebFilesInOutputFolder {
		for _, file := range webFilePaths {
			err := os.Remove(file)
			if err != nil {
				logrus.Fatalf("failed to remove web file %s: %v", file, err)
			}
		}
	}
}
