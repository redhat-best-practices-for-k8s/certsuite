// Copyright (C) 2020-2021 Red Hat, Inc.
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
	j "encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/junit"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"

	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/certification"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle"

	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/observability"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/operator"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform"
	"github.com/test-network-function/cnf-certification-test/pkg/diagnostics"
)

const (
	claimFileName                        = "claim.json"
	claimFilePermissions                 = 0o644
	claimPathFlagKey                     = "claimloc"
	CnfCertificationTestSuiteName        = "CNF Certification Test Suite"
	defaultClaimPath                     = ".."
	defaultCliArgValue                   = ""
	junitFlagKey                         = "junit"
	TNFJunitXMLFileName                  = "cnf-certification-tests_junit.xml"
	TNFReportKey                         = "cnf-certification-test"
	CNFFeatureValidationJunitXMLFileName = "validation_junit.xml"
	CNFFeatureValidationReportKey        = "cnf-feature-validation"
	// dateTimeFormatDirective is the directive used to format date/time according to ISO 8601.
	dateTimeFormatDirective = "2006-01-02T15:04:05+00:00"
	extraInfoKey            = "testsExtraInfo"
)

var (
	claimPath *string
	junitPath *string
	// GitCommit is the latest commit in the current git branch
	GitCommit string
	// GitRelease is the list of tags (if any) applied to the latest commit
	// in the current branch
	GitRelease string
	// GitPreviousRelease is the last release at the date of the latest commit
	// in the current branch
	GitPreviousRelease string
	// gitDisplayRelease is a string used to hold the text to display
	// the version on screen and in the claim file
	gitDisplayRelease string
)

func init() {
	claimPath = flag.String(claimPathFlagKey, defaultClaimPath,
		"the path where the claimfile will be output")
	junitPath = flag.String(junitFlagKey, defaultCliArgValue,
		"the path for the junit format report")
}

// createClaimRoot creates the claim based on the model created in
// https://github.com/test-network-function/cnf-certification-test-claim.
func createClaimRoot() *claim.Root {
	// Initialize the claim with the start time.
	startTime := time.Now()
	c := &claim.Claim{
		Metadata: &claim.Metadata{
			StartTime: startTime.UTC().Format(dateTimeFormatDirective),
		},
	}
	return &claim.Root{
		Claim: c,
	}
}

// loadJUnitXMLIntoMap converts junitFilename's XML-formatted JUnit test results into a Go map, and adds the result to
// the result Map.
func loadJUnitXMLIntoMap(result map[string]interface{}, junitFilename, key string) {
	var err error
	if key == "" {
		var extension = filepath.Ext(junitFilename)
		key = junitFilename[0 : len(junitFilename)-len(extension)]
	}
	result[key], err = junit.ExportJUnitAsMap(junitFilename)
	if err != nil {
		log.Fatalf("error reading JUnit XML file into JSON: %v", err)
	}
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

// setLogFormat sets the log format for logrus
func setLogFormat() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = time.StampMilli
	customFormatter.PadLevelText = true
	customFormatter.FullTimestamp = true
	customFormatter.ForceColors = true
	log.SetReportCaller(true)
	customFormatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
		_, filename := path.Split(f.File)
		return strconv.Itoa(f.Line) + "]", fmt.Sprintf("[%s:", filename)
	}
	log.SetFormatter(customFormatter)
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

// getGitVersion returns the git display version: the latest previously released
// build in case this build is not released. Otherwise display the build version
func getGitVersion() string {
	if GitRelease == "" {
		gitDisplayRelease = "Unreleased build post " + GitPreviousRelease
	} else {
		gitDisplayRelease = GitRelease
	}

	return gitDisplayRelease + " ( " + GitCommit + " )"
}

//nolint:funlen // TestTest invokes the CNF Certification Test Suite.
func TestTest(t *testing.T) {
	// When running unit tests, skip the suite
	if os.Getenv("UNIT_TEST") != "" {
		t.Skip("Skipping test suite when running unit tests")
	}

	// Set up logging params for logrus
	setLogFormat()
	setLogLevel()

	ginkgoConfig, _ := ginkgo.GinkgoConfiguration()
	log.Infof("TNF Version         : %v", getGitVersion())
	log.Infof("Ginkgo Version      : %v", ginkgo.GINKGO_VERSION)
	log.Infof("Focused test suites : %v", ginkgoConfig.FocusStrings)
	log.Infof("TC skip patterns    : %v", ginkgoConfig.SkipStrings)
	log.Infof("Labels filter       : %v", ginkgoConfig.LabelFilter)

	// Diagnostic functions will run also when no focus test suites were provided.
	diagnosticMode := len(ginkgoConfig.FocusStrings) == 0

	gomega.RegisterFailHandler(ginkgo.Fail)

	// Set clientsholder singleton with the filenames from the env vars.
	_ = clientsholder.GetClientsHolder(getK8sClientsConfigFileNames()...)

	// Initialize the claim with the start time, tnf version, etc.
	claimRoot := createClaimRoot()
	claimData := claimRoot.Claim
	claimData.Configurations = make(map[string]interface{})
	claimData.Nodes = make(map[string]interface{})
	incorporateVersions(claimData)

	configurations := marshalConfigurations()
	claimData.Nodes = generateNodes()
	unmarshalConfigurations(configurations, claimData.Configurations)

	// Run tests specs only if not in diagnostic mode, otherwise all TSs would run.
	if !diagnosticMode {
		ginkgo.RunSpecs(t, CnfCertificationTestSuiteName)
	}

	endTime := time.Now()
	claimData.Metadata.EndTime = endTime.UTC().Format(dateTimeFormatDirective)

	// Process the test results from the suites, the cnf-features-deploy test suite,
	// and any extra informational messages.
	junitMap := make(map[string]interface{})
	cnfCertificationJUnitFilename := filepath.Join(*junitPath, TNFJunitXMLFileName)

	if !diagnosticMode {
		loadJUnitXMLIntoMap(junitMap, cnfCertificationJUnitFilename, TNFReportKey)
		appendCNFFeatureValidationReportResults(junitPath, junitMap)
	}

	junitMap[extraInfoKey] = "" // tnf.TestsExtraInfo

	// Append results to claim file data.
	claimData.RawResults = junitMap
	claimData.Results = results.GetReconciledResults()

	// Marshal the claim and output to file
	payload := marshalClaimOutput(claimRoot)
	claimOutputFile := filepath.Join(*claimPath, claimFileName)
	writeClaimOutput(claimOutputFile, payload)
}

// incorporateTNFVersion adds the TNF version to the claim.
func incorporateVersions(claimData *claim.Claim) {
	claimData.Versions = &claim.Versions{
		Tnf:          gitDisplayRelease,
		TnfGitCommit: GitCommit,
		OcClient:     diagnostics.GetVersionOcClient(),
		Ocp:          diagnostics.GetVersionOcp(),
		K8s:          diagnostics.GetVersionK8s(),
	}
}

// appendCNFFeatureValidationReportResults is a helper method to add the results of running the cnf-features-deploy
// test suite to the claim file.
func appendCNFFeatureValidationReportResults(junitPath *string, junitMap map[string]interface{}) {
	cnfFeaturesDeployJUnitFile := filepath.Join(*junitPath, CNFFeatureValidationJunitXMLFileName)
	if _, err := os.Stat(cnfFeaturesDeployJUnitFile); err == nil {
		loadJUnitXMLIntoMap(junitMap, cnfFeaturesDeployJUnitFile, CNFFeatureValidationReportKey)
	}
}

// marshalConfigurations creates a byte stream representation of the test configurations.  In the event of an error,
// this method fatally fails.
func marshalConfigurations() []byte {
	config := provider.GetTestEnvironment()
	configurations, err := j.Marshal(config)
	if err != nil {
		log.Fatalf("error converting configurations to JSON: %v", err)
	}
	return configurations
}

// unmarshalConfigurations creates a map from configurations byte stream.  In the event of an error, this method fatally
// fails.
func unmarshalConfigurations(configurations []byte, claimConfigurations map[string]interface{}) {
	err := j.Unmarshal(configurations, &claimConfigurations)
	if err != nil {
		log.Fatalf("error unmarshalling configurations: %v", err)
	}
}

// marshalClaimOutput is a helper function to serialize a claim as JSON for output.  In the event of an error, this
// method fatally fails.
func marshalClaimOutput(claimRoot *claim.Root) []byte {
	payload, err := j.MarshalIndent(claimRoot, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate the claim: %v", err)
	}
	return payload
}

// writeClaimOutput writes the output payload to the claim file.  In the event of an error, this method fatally fails.
func writeClaimOutput(claimOutputFile string, payload []byte) {
	err := os.WriteFile(claimOutputFile, payload, claimFilePermissions)
	if err != nil {
		log.Fatalf("Error writing claim data:\n%s", string(payload))
	}
}

//no-lint:commentedOutCode
func generateNodes() map[string]interface{} {
	const (
		nodeSummaryField = "nodeSummary"
		cniPluginsField  = "cniPlugins"
		nodesHwInfo      = "nodesHwInfo"
		csiDriverInfo    = "csiDriver"
	)
	nodes := map[string]interface{}{}
	nodes[nodeSummaryField] = diagnostics.GetNodeJSON()  // add node summary
	nodes[cniPluginsField] = diagnostics.GetCniPlugins() // add cni plugins
	nodes[nodesHwInfo] = diagnostics.GetHwInfoAllNodes() // add nodes hardware information
	nodes[csiDriverInfo] = diagnostics.GetCsiDriver()    // add csi drivers info
	return nodes
}
