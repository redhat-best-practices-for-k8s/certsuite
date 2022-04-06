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
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/claimhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/loghelper"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"

	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/certification"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle"

	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/observability"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/operator"
	_ "github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform"
	"github.com/test-network-function/cnf-certification-test/pkg/diagnostics"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
)

const (
	claimFileName                 = "claim.json"
	claimPathFlagKey              = "claimloc"
	CnfCertificationTestSuiteName = "CNF Certification Test Suite"
	defaultClaimPath              = ".."
	defaultCliArgValue            = ""
	junitFlagKey                  = "junit"
	TNFJunitXMLFileName           = "cnf-certification-tests_junit.xml"
	TNFReportKey                  = "cnf-certification-test"
	extraInfoKey                  = "testsExtraInfo"
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

//nolint:funlen // TestTest invokes the CNF Certification Test Suite.
func TestTest(t *testing.T) {
	// When running unit tests, skip the suite
	if os.Getenv("UNIT_TEST") != "" {
		t.Skip("Skipping test suite when running unit tests")
	}

	ginkgoConfig, _ := ginkgo.GinkgoConfiguration()
	log.Infof("TNF Version         : %v", getGitVersion())
	log.Infof("Ginkgo Version      : %v", ginkgo.GINKGO_VERSION)
	log.Infof("Focused test suites : %v", ginkgoConfig.FocusStrings)
	log.Infof("TC skip patterns    : %v", ginkgoConfig.SkipStrings)
	log.Infof("Labels filter       : %v", ginkgoConfig.LabelFilter)

	// Diagnostic functions will run also when no focus test suites were provided.
	diagnosticMode := len(ginkgoConfig.FocusStrings) == 0

	gomega.RegisterFailHandler(ginkgo.Fail)

	loghelper.SetLogFormat()

	// Set clientsholder singleton with the filenames from the env vars.
	_ = clientsholder.GetClientsHolder(getK8sClientsConfigFileNames()...)

	// Initialize the claim with the start time, tnf version, etc.
	claimRoot := claimhelper.CreateClaimRoot()
	claimData := claimRoot.Claim
	claimData.Configurations = make(map[string]interface{})
	claimData.Nodes = make(map[string]interface{})
	incorporateVersions(claimData)

	configurations, err := claimhelper.MarshalConfigurations()
	if err != nil {
		log.Errorf("Configuration node missing because of: %s", err)
	}

	claimData.Nodes = claimhelper.GenerateNodes()
	claimhelper.UnmarshalConfigurations(configurations, claimData.Configurations)

	// Run tests specs only if not in diagnostic mode, otherwise all TSs would run.
	if !diagnosticMode {
		ginkgo.RunSpecs(t, CnfCertificationTestSuiteName)
	}

	endTime := time.Now()
	claimData.Metadata.EndTime = endTime.UTC().Format(claimhelper.DateTimeFormatDirective)

	// Process the test results from the suites, the cnf-features-deploy test suite,
	// and any extra informational messages.
	junitMap := make(map[string]interface{})
	cnfCertificationJUnitFilename := filepath.Join(*junitPath, TNFJunitXMLFileName)

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
	claimOutputFile := filepath.Join(*claimPath, claimFileName)
	claimhelper.WriteClaimOutput(claimOutputFile, payload)
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
