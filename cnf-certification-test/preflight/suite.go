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
	plibRuntime "github.com/sebrandon1/openshift-preflight/certification/runtime"
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
	var env provider.TestEnvironment
	env = provider.GetTestEnvironment()
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	logrus.Infof("Number of containers to gather results from: %d", len(env.Containers))
	logrus.Infof("Number of operators to gather results from: %d", len(env.Operators))
	containerTestEntries := gatherTestNamesFromContainerResults(env.Containers)
	operatorTestEntries := gatherTestNamesFromOperatorResults(env.Operators)

	// Handle Container-based preflight tests
	for testName, testEntry := range containerTestEntries {
		// Store the test names into the Catalog map for results to be dynamically printed
		aID := identifiers.AddCatalogEntry(testName, common.PreflightTestKey, testEntry.Metadata().Description, testEntry.Help().Suggestion, "", "", "", "", "common")
		testID, tags := identifiers.GetGinkgoTestIDAndLabels(aID)

		logrus.Infof("Testing ginkgo test: %s ID: %s", testName, testID)
		GeneratePreflightContainerGinkgoTest(testName, testID, tags, env.Containers)
	}

	// Handle Operator-based preflight tests
	for testName, testEntry := range operatorTestEntries {
		// Store the test names into the Catalog map for results to be dynamically printed
		aID := identifiers.AddCatalogEntry(testName, common.PreflightTestKey, testEntry.Metadata().Description, testEntry.Help().Suggestion, "", "", "", "", "common")
		testID, tags := identifiers.GetGinkgoTestIDAndLabels(aID)

		logrus.Infof("Testing ginkgo test: %s ID: %s", testName, testID)
		GeneratePreflightOperatorGinkgoTest(testName, testID, tags, env.Operators)
	}
})

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
		var failedContainers []string
		var erroredContainers []string
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
					failedContainers = append(failedContainers, op.String())
				}
			}
			for _, i := range op.PreflightResults.Errors {
				if i.Name() == testName {
					logrus.Infof("%s has errored preflight test: %s", op.String(), testName)
					tnf.ClaimFilePrintf("%s has errored preflight test: %s", op.String(), testName)
					erroredContainers = append(erroredContainers, op.String())
				}
			}
		}
		testhelper.AddTestResultLog("Non-compliant", failedContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
		testhelper.AddTestResultLog("Error", erroredContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
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
