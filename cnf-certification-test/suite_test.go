// Copyright (C) 2020-2023 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package suite

import (
	_ "embed"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/claimhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/collector"
	"github.com/test-network-function/cnf-certification-test/pkg/loghelper"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"

	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/certification"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/chaostesting"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/manageability"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/observability"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/operator"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/performance"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/preflight"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/version"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
)

const (
	claimPathFlagKey              = "claimloc"
	CnfCertificationTestSuiteName = "CNF Certification Test Suite"
	defaultClaimPath              = ".."
	defaultCliArgValue            = ""
	junitFlagKey                  = "junit"
	TNFReportKey                  = "cnf-certification-test"
	extraInfoKey                  = "testsExtraInfo"
)

var (
	claimPath *string
	junitPath *string
)

func init() {
	claimPath = flag.String(claimPathFlagKey, defaultClaimPath,
		"the path where the claimfile will be output")
	junitPath = flag.String(junitFlagKey, defaultCliArgValue,
		"the path for the junit format report")
}

// setLogLevel sets the log level for logrus based on the "TNF_LOG_LEVEL" environment variable
func setLogLevel() {
	params := configuration.GetTestParameters()

	var logLevel, err = log.ParseLevel(params.LogLevel)
	if err != nil {
		log.Error("TNF_LOG_LEVEL environment set with an invalid value, defaulting to DEBUG \n Valid values are:  trace, debug, info, warn, error, fatal, panic")
		logLevel = log.DebugLevel
	}

	log.Info("Log level set to: ", logLevel)
	log.SetLevel(logLevel)
}

func getK8sClientsConfigFileNames() []string {
	params := configuration.GetTestParameters()
	fileNames := []string{}
	if params.Kubeconfig != "" {
		fileNames = append(fileNames, params.Kubeconfig)
	}
	if params.Home != "" {
		kubeConfigFilePath := filepath.Join(params.Home, ".kube", "config")
		fileNames = append(fileNames, kubeConfigFilePath)
	}

	return fileNames
}

// TestTest invokes the CNF Certification Test Suite.
func TestTest(t *testing.T) {
	// When running unit tests, skip the suite
	if os.Getenv("UNIT_TEST") != "" {
		t.Skip("Skipping test suite when running unit tests")
	}

	err := configuration.LoadEnvironmentVariables()
	if err != nil {
		log.Fatalf("could not load the environment variables, error: %v", err)
	}

	// Set up logging params for logrus
	loghelper.SetLogFormat()
	setLogLevel()

	ginkgoConfig, _ := ginkgo.GinkgoConfiguration()
	log.Infof("TNF Version         : %v", version.GetGitVersion())
	log.Infof("Ginkgo Version      : %v", ginkgo.GINKGO_VERSION)
	log.Infof("Labels filter       : %v", ginkgoConfig.LabelFilter)

	// Diagnostic functions will run when no labels are provided.
	var diagnosticMode bool
	if ginkgoConfig.LabelFilter == "" {
		log.Infof("TNF will run in diagnostic mode so no test case will be launched.")
		diagnosticMode = true
	}

	// Set clientsholder singleton with the filenames from the env vars.
	_ = clientsholder.GetClientsHolder(getK8sClientsConfigFileNames()...)

	// Initialize the claim with the start time, tnf version, etc.
	claimRoot := claimhelper.CreateClaimRoot()
	claimData := claimRoot.Claim
	claimData.Configurations = make(map[string]interface{})
	claimData.Nodes = make(map[string]interface{})
	claimhelper.IncorporateVersions(claimData)

	configurations, err := claimhelper.MarshalConfigurations()
	if err != nil {
		log.Errorf("Configuration node missing because of: %s", err)
		t.FailNow()
	}

	claimData.Nodes = claimhelper.GenerateNodes()
	claimhelper.UnmarshalConfigurations(configurations, claimData.Configurations)

	// initialize abort flag
	testhelper.AbortTrigger = ""

	// Run tests specs only if not in diagnostic mode, otherwise all TSs would run.
	var env provider.TestEnvironment
	if !diagnosticMode {
		env.SetNeedsRefresh()
		env = provider.GetTestEnvironment()
		ginkgo.RunSpecs(t, CnfCertificationTestSuiteName)
	}

	endTime := time.Now()
	claimData.Metadata.EndTime = endTime.UTC().Format(claimhelper.DateTimeFormatDirective)

	// Process the test results from the suites, the cnf-features-deploy test suite,
	// and any extra informational messages.
	junitMap := make(map[string]interface{})
	cnfCertificationJUnitFilename := filepath.Join(*junitPath, results.JunitXMLFileName)

	if !diagnosticMode {
		claimhelper.LoadJUnitXMLIntoMap(junitMap, cnfCertificationJUnitFilename, TNFReportKey)
		claimhelper.AppendCNFFeatureValidationReportResults(junitPath, junitMap)
	}

	junitMap[extraInfoKey] = "" // tnf.TestsExtraInfo

	// Append results to claim file data.
	claimData.RawResults = junitMap
	claimData.Results = results.GetReconciledResults()

	// Marshal the claim and output to file
	payload := claimhelper.MarshalClaimOutput(claimRoot)
	claimOutputFile := filepath.Join(*claimPath, results.ClaimFileName)
	claimhelper.WriteClaimOutput(claimOutputFile, payload)

	// Send claim file to the collector if specified by env var
	if configuration.GetTestParameters().EnableDataCollection {
		err = collector.SendClaimFileToCollector(env.CollectorAppEndPoint, claimOutputFile, env.ExecutedBy, env.PartnerName, env.CollectorAppPassword)
		if err != nil {
			log.Errorf("Failed to send post request to the collector: %v", err)
		}
	}

	// Create HTML artifacts for the web results viewer/parser.
	resultsOutputDir := *claimPath
	webFilePaths, err := results.CreateResultsWebFiles(resultsOutputDir)
	if err != nil {
		log.Errorf("Failed to create results web files: %v", err)
	}

	allArtifactsFilePaths := []string{filepath.Join(*claimPath, results.ClaimFileName)}

	// Add the junit xml file only if we're not in diagnostic mode.
	if !diagnosticMode {
		allArtifactsFilePaths = append(allArtifactsFilePaths, filepath.Join(*junitPath, results.JunitXMLFileName))
	}

	// Add all the web artifacts file paths.
	allArtifactsFilePaths = append(allArtifactsFilePaths, webFilePaths...)

	// tar.gz file creation with results and html artifacts, unless omitted by env var.
	if !configuration.GetTestParameters().OmitArtifactsZipFile {
		err = results.CompressResultsArtifacts(resultsOutputDir, allArtifactsFilePaths)
		if err != nil {
			log.Fatalf("Failed to compress results artifacts: %v", err)
		}
	}

	// Remove web artifacts if user does not want them.
	if !configuration.GetTestParameters().IncludeWebFilesInOutputFolder {
		for _, file := range webFilePaths {
			err := os.Remove(file)
			if err != nil {
				log.Fatalf("failed to remove web file %s: %v", file, err)
			}
		}
	}
}
