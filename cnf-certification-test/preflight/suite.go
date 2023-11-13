// Copyright (C) 2022-2023 Red Hat, Inc.
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

package preflight

import (
	"strings"

	"github.com/onsi/ginkgo/v2"
	plibRuntime "github.com/redhat-openshift-ecosystem/openshift-preflight/certification"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

var _ = ginkgo.Describe(common.PreflightTestKey, func() {
	logrus.Debugf("Entering %s suite", common.PreflightTestKey)
	env := provider.GetTestEnvironment()
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})

	// Add safeguard against running the tests if the docker config does not exist.
	if env.GetDockerConfigFile() == "" || env.GetDockerConfigFile() == "NA" {
		logrus.Debug("Skipping the preflight suite because the Docker Config file is not provided.")
		return
	}

	// Safeguard against running the preflight tests if the label filter is set but does not include the preflight label
	ginkgoConfig, _ := ginkgo.GinkgoConfiguration()
	if !labelsAllowTestRun(ginkgoConfig.LabelFilter, []string{common.PreflightTestKey, identifiers.TagPreflight}) {
		logrus.Warn("LabelFilter is set but 'preflight' tests are not targeted. Skipping the preflight tests.")
		return
	}

	testPreflightContainers(&env)
	if provider.IsOCPCluster() {
		logrus.Debugf("OCP cluster detected, allowing operator tests to run")
		testPreflightOperators(&env)
	} else {
		logrus.Debugf("Skipping the preflight operators test because it requires an OCP cluster to run against")
	}
})

func labelsAllowTestRun(labelFilter string, allowedLabels []string) bool {
	for _, label := range allowedLabels {
		if strings.Contains(labelFilter, label) {
			return true
		}
	}
	return false
}

func testPreflightOperators(env *provider.TestEnvironment) {
	// Loop through all of the operators, run preflight, and set their results into their respective object
	for _, op := range env.Operators {
		// Note: We are not using a cache here for the operator bundle images because
		// in-general you are only going to have an operator installed once in a cluster.
		err := op.SetPreflightResults(env)
		if err != nil {
			logrus.Fatalf("failed running preflight on operator: %s error: %v", op.Name, err)
		}
	}

	logrus.Infof("Completed running preflight operator tests for %d operators", len(env.Operators))

	// Handle Operator-based preflight tests
	// Note: We only care about the `testEntry` variable below because we need its 'Description' and 'Suggestion' variables.
	for testName, testEntry := range getUniqueTestEntriesFromOperatorResults(env.Operators) {
		logrus.Infof("Testing operator ginkgo test: %s", testName)
		generatePreflightOperatorGinkgoTest(testName, testEntry.Metadata().Description, testEntry.Help().Suggestion, env.Operators)
	}
}

func testPreflightContainers(env *provider.TestEnvironment) {
	// Using a cache to prevent unnecessary processing of images if we already have the results available
	preflightImageCache := make(map[string]plibRuntime.Results)

	// Loop through all of the containers, run preflight, and set their results into their respective objects
	for _, cut := range env.Containers {
		logrus.Debugf("Running preflight container tests for: %s", cut.Name)
		err := cut.SetPreflightResults(preflightImageCache, env)
		if err != nil {
			logrus.Fatalf("failed running preflight on image: %s error: %v", cut.Image, err)
		}
	}

	logrus.Infof("Completed running preflight container tests for %d containers", len(env.Containers))

	// Handle Container-based preflight tests
	// Note: We only care about the `testEntry` variable below because we need its 'Description' and 'Suggestion' variables.
	for testName, testEntry := range getUniqueTestEntriesFromContainerResults(env.Containers) {
		logrus.Infof("Testing container ginkgo test: %s", testName)
		generatePreflightContainerGinkgoTest(testName, testEntry.Metadata().Description, testEntry.Help().Suggestion, env.Containers)
	}
}

// func generatePreflightContainerGinkgoTest(testName, testID string, tags []string, containers []*provider.Container) {
func generatePreflightContainerGinkgoTest(testName, description, suggestion string, containers []*provider.Container) {
	// Based on a single test "name", we will be passing/failing in Ginkgo.
	// Brute force-ish type of method.

	// Store the test names into the Catalog map for results to be dynamically printed
	aID := identifiers.AddCatalogEntry(testName, common.PreflightTestKey, description, suggestion, "", "", false, map[string]string{
		identifiers.FarEdge:  identifiers.Optional,
		identifiers.Telco:    identifiers.Optional,
		identifiers.NonTelco: identifiers.Optional,
		identifiers.Extended: identifiers.Optional,
	}, identifiers.TagPreflight)
	testID, tags := identifiers.GetGinkgoTestIDAndLabels(aID)

	// Start the ginkgo It block
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		// Collect all of the failed and errored containers
		var failedContainers []string
		var erroredContainers []string
		for _, cut := range containers {
			for _, r := range cut.PreflightResults.Passed {
				if r.Name() == testName {
					logrus.Infof("%s has passed preflight test: %s", cut.String(), testName)
				}
			}
			for _, r := range cut.PreflightResults.Failed {
				if r.Name() == testName {
					tnf.Logf(logrus.WarnLevel, "%s has failed preflight test: %s", cut, testName)
					failedContainers = append(failedContainers, cut.String())
				}
			}
			for _, r := range cut.PreflightResults.Errors {
				if r.Name() == testName {
					tnf.Logf(logrus.ErrorLevel, "%s has errored preflight test: %s", cut, testName)
					erroredContainers = append(erroredContainers, cut.String())
				}
			}
		}
		testhelper.AddTestResultLog("Non-compliant", failedContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
		testhelper.AddTestResultLog("Error", erroredContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
	})
}

func generatePreflightOperatorGinkgoTest(testName, description, suggestion string, operators []*provider.Operator) {
	// Based on a single test "name", we will be passing/failing in Ginkgo.
	// Brute force-ish type of method.

	// Store the test names into the Catalog map for results to be dynamically printed
	aID := identifiers.AddCatalogEntry(testName, common.PreflightTestKey, description, suggestion, "", "", false, map[string]string{
		identifiers.FarEdge:  identifiers.Optional,
		identifiers.Telco:    identifiers.Optional,
		identifiers.NonTelco: identifiers.Optional,
		identifiers.Extended: identifiers.Optional,
	}, identifiers.TagPreflight)
	testID, tags := identifiers.GetGinkgoTestIDAndLabels(aID)

	// Start the ginkgo It block
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		// Collect all of the failed and errored containers
		var failedOperators []string
		var erroredOperators []string
		for _, op := range operators {
			for _, r := range op.PreflightResults.Passed {
				if r.Name() == testName {
					logrus.Infof("%s has passed preflight test: %s", op.String(), testName)
				}
			}
			for _, r := range op.PreflightResults.Failed {
				if r.Name() == testName {
					tnf.Logf(logrus.WarnLevel, "%s has failed preflight test: %s", op, testName)
					failedOperators = append(failedOperators, op.String())
				}
			}
			for _, r := range op.PreflightResults.Errors {
				if r.Name() == testName {
					tnf.Logf(logrus.ErrorLevel, "%s has errored preflight test: %s", op, testName)
					erroredOperators = append(erroredOperators, op.String())
				}
			}
		}
		testhelper.AddTestResultLog("Non-compliant", failedOperators, tnf.ClaimFilePrintf, ginkgo.Fail)
		testhelper.AddTestResultLog("Error", erroredOperators, tnf.ClaimFilePrintf, ginkgo.Fail)
	})
}

func getUniqueTestEntriesFromContainerResults(containers []*provider.Container) map[string]plibRuntime.Result {
	// If containers are sharing the same image, they should "presumably" have the same results returned from preflight.
	testEntries := make(map[string]plibRuntime.Result)
	for _, cut := range containers {
		for _, r := range cut.PreflightResults.Passed {
			testEntries[r.Name()] = r
		}
		// Failed Results have more information than the rest
		for _, r := range cut.PreflightResults.Failed {
			testEntries[r.Name()] = r
		}
		for _, r := range cut.PreflightResults.Errors {
			testEntries[r.Name()] = r
		}
	}

	return testEntries
}

func getUniqueTestEntriesFromOperatorResults(operators []*provider.Operator) map[string]plibRuntime.Result {
	testEntries := make(map[string]plibRuntime.Result)
	for _, op := range operators {
		for _, r := range op.PreflightResults.Passed {
			testEntries[r.Name()] = r
		}
		// Failed Results have more information than the rest
		for _, r := range op.PreflightResults.Failed {
			testEntries[r.Name()] = r
		}
		for _, r := range op.PreflightResults.Errors {
			testEntries[r.Name()] = r
		}
	}
	return testEntries
}
