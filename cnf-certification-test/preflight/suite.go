// Copyright (C) 2022 Red Hat, Inc.
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
	"github.com/onsi/ginkgo/v2"
	plibRuntime "github.com/redhat-openshift-ecosystem/openshift-preflight/certification"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
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
	ginkgo.ReportAfterEach(results.RecordResult)

	testPreflightContainers(&env)
	if provider.IsOCPCluster() {
		logrus.Debugf("OCP cluster detected, allowing operator tests to run")
		testPreflightOperators(&env)
	} else {
		logrus.Debugf("Skipping the preflight operators test because it requires an OCP cluster to run against")
	}
})

func testPreflightOperators(env *provider.TestEnvironment) {
	// Loop through all of the operators, run preflight, and set their results into their respective object
	for _, op := range env.Operators {
		err := op.SetPreflightResults(env.IsPreflightInsecureAllowed())
		if err != nil {
			logrus.Fatalf("failed running preflight on operator: %s error: %v", op.Name, err)
		}
	}

	operatorTestEntries := gatherTestNamesFromOperatorResults(env.Operators)
	// Handle Operator-based preflight tests
	for testName, testEntry := range operatorTestEntries {
		// Store the test names into the Catalog map for results to be dynamically printed
		aID := identifiers.AddCatalogEntry(testName, common.PreflightTestKey, testEntry.Metadata().Description, testEntry.Help().Suggestion, "", "", "", "", false, identifiers.TagCommon)
		testID, tags := identifiers.GetGinkgoTestIDAndLabels(aID)

		logrus.Infof("Testing ginkgo test: %s ID: %s", testName, testID)
		GeneratePreflightOperatorGinkgoTest(testName, testID, tags, env.Operators)
	}
}

func testPreflightContainers(env *provider.TestEnvironment) {
	// Using a cache to prevent unnecessary processing of images if we already have the results available
	preflightImageCache := make(map[string]plibRuntime.Results)

	// Loop through all of the containers, run preflight, and set their results into their respective object
	for _, cut := range env.Containers {
		logrus.Debugf("Running preflight container tests for: %s", cut.Name)
		err := cut.SetPreflightResults(preflightImageCache, env.IsPreflightInsecureAllowed())
		if err != nil {
			logrus.Fatalf("failed running preflight on image: %s error: %v", cut.Image, err)
		}
	}

	containerTestEntries := gatherTestNamesFromContainerResults(env.Containers)

	// Handle Container-based preflight tests
	for testName, testEntry := range containerTestEntries {
		// Store the test names into the Catalog map for results to be dynamically printed
		aID := identifiers.AddCatalogEntry(testName, common.PreflightTestKey, testEntry.Metadata().Description, testEntry.Help().Suggestion, "", "", "", "", false, identifiers.TagCommon)
		testID, tags := identifiers.GetGinkgoTestIDAndLabels(aID)

		logrus.Infof("Testing ginkgo test: %s ID: %s", testName, testID)
		GeneratePreflightContainerGinkgoTest(testName, testID, tags, env.Containers)
	}
}

func GeneratePreflightContainerGinkgoTest(testName, testID string, tags []string, containers []*provider.Container) {
	// Based on a single test "name", we will be passing/failing in Ginkgo.
	// Brute force-ish type of method.
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		// Collect all of the failed and errored containers
		var failedContainers []string
		var erroredContainers []string
		for _, cut := range containers {
			for _, i := range cut.PreflightResults.Passed {
				if i.Name() == testName {
					logrus.Infof("%s has passed preflight test: %s", cut.String(), testName)
				}
			}
			for _, i := range cut.PreflightResults.Failed {
				if i.Name() == testName {
					logrus.Infof("%s has failed preflight test: %s", cut.String(), testName)
					tnf.ClaimFilePrintf("%s has failed preflight test: %s", cut.String(), testName)
					failedContainers = append(failedContainers, cut.String())
				}
			}
			for _, i := range cut.PreflightResults.Errors {
				if i.Name() == testName {
					logrus.Infof("%s has errored preflight test: %s", cut.String(), testName)
					tnf.ClaimFilePrintf("%s has errored preflight test: %s", cut.String(), testName)
					erroredContainers = append(erroredContainers, cut.String())
				}
			}
		}
		testhelper.AddTestResultLog("Non-compliant", failedContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
		testhelper.AddTestResultLog("Error", erroredContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
	})
}

func GeneratePreflightOperatorGinkgoTest(testName, testID string, tags []string, operators []*provider.Operator) {
	// Based on a single test "name", we will be passing/failing in Ginkgo.
	// Brute force-ish type of method.
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		// Collect all of the failed and errored containers
		var failedOperators []string
		var erroredOperators []string
		for _, op := range operators {
			for _, i := range op.PreflightResults.Passed {
				if i.Name() == testName {
					logrus.Infof("%s has passed preflight test: %s", op.String(), testName)
				}
			}
			for _, i := range op.PreflightResults.Failed {
				if i.Name() == testName {
					logrus.Infof("%s has failed preflight test: %s", op.String(), testName)
					tnf.ClaimFilePrintf("%s has failed preflight test: %s", op.String(), testName)
					failedOperators = append(failedOperators, op.String())
				}
			}
			for _, i := range op.PreflightResults.Errors {
				if i.Name() == testName {
					logrus.Infof("%s has errored preflight test: %s", op.String(), testName)
					tnf.ClaimFilePrintf("%s has errored preflight test: %s", op.String(), testName)
					erroredOperators = append(erroredOperators, op.String())
				}
			}
		}
		testhelper.AddTestResultLog("Non-compliant", failedOperators, tnf.ClaimFilePrintf, ginkgo.Fail)
		testhelper.AddTestResultLog("Error", erroredOperators, tnf.ClaimFilePrintf, ginkgo.Fail)
	})
}

func gatherTestNamesFromContainerResults(containers []*provider.Container) map[string]plibRuntime.Result {
	testEntries := make(map[string]plibRuntime.Result)
	for _, cut := range containers {
		for _, i := range cut.PreflightResults.Passed {
			testEntries[i.Name()] = i
		}
		// Failed Results have more information than the rest
		for _, i := range cut.PreflightResults.Failed {
			testEntries[i.Name()] = i
		}
		for _, i := range cut.PreflightResults.Errors {
			testEntries[i.Name()] = i
		}
	}

	return testEntries
}

func gatherTestNamesFromOperatorResults(operators []*provider.Operator) map[string]plibRuntime.Result {
	testEntries := make(map[string]plibRuntime.Result)
	for _, op := range operators {
		for _, i := range op.PreflightResults.Passed {
			testEntries[i.Name()] = i
		}
		// Failed Results have more information than the rest
		for _, i := range op.PreflightResults.Failed {
			testEntries[i.Name()] = i
		}
		for _, i := range op.PreflightResults.Errors {
			testEntries[i.Name()] = i
		}
	}
	return testEntries
}
