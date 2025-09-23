// Copyright (C) 2022-2024 Red Hat, Inc.
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
	"fmt"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}
)

// labelsAllowTestRun checks whether a test run is permitted based on labels
//
// The function receives a string of labels and a list of allowed label
// identifiers. It scans each allowed identifier to see if it appears within the
// provided string, returning true upon the first match. If none of the allowed
// labels are found, it returns false.
func labelsAllowTestRun(labelFilter string, allowedLabels []string) bool {
	for _, label := range allowedLabels {
		if strings.Contains(labelFilter, label) {
			return true
		}
	}
	return false
}

// ShouldRun Determines whether preflight checks should be executed
//
// The function evaluates the provided label expression to see if it includes
// any preflight-specific tags, then verifies that a Docker configuration file
// is available. If either condition fails, it returns false or logs a warning
// and marks the environment to skip preflight tests. When both conditions are
// satisfied, it signals that the preflight suite may run.
func ShouldRun(labelsExpr string) bool {
	env = provider.GetTestEnvironment()
	preflightAllowedLabels := []string{common.PreflightTestKey, identifiers.TagPreflight}

	if !labelsAllowTestRun(labelsExpr, preflightAllowedLabels) {
		return false
	}

	// Add safeguard against running the preflight tests if the docker config does not exist.
	preflightDockerConfigFile := configuration.GetTestParameters().PfltDockerconfig
	if preflightDockerConfigFile == "" || preflightDockerConfigFile == "NA" {
		log.Warn("Skipping the preflight suite because the Docker Config file is not provided.")
		env.SkipPreflight = true
	}

	return true
}

// LoadChecks Initializes the test environment and runs Preflight checks for containers and operators
//
// The function sets up logging, retrieves the current test environment, and
// creates a checks group for Preflight tests. It executes container preflight
// tests and conditionally runs operator tests if the cluster is OpenShift.
// Results are recorded in the checks group for later reporting.
func LoadChecks() {
	log.Debug("Running %s suite checks", common.PreflightTestKey)

	// As the preflight lib's checks need to run here, we need to get the test environment now.
	env = provider.GetTestEnvironment()

	checksGroup := checksdb.NewChecksGroup(common.PreflightTestKey).
		WithBeforeEachFn(beforeEachFn)

	testPreflightContainers(checksGroup, &env)
	if provider.IsOCPCluster() {
		log.Info("OCP cluster detected, allowing Preflight operator tests to run")
		testPreflightOperators(checksGroup, &env)
	} else {
		log.Info("Skipping the Preflight operators test because it requires an OCP cluster to run against")
	}
}

// testPreflightOperators Runs preflight checks on all operators and records their outcomes
//
// This function iterates over each operator in the test environment, executing
// its preflight tests and capturing any errors. After collecting results, it
// logs completion of operator testing. Finally, it creates catalog entries for
// every unique preflight test found across operators, adding these checks to
// the provided group so they can be reported.
func testPreflightOperators(checksGroup *checksdb.ChecksGroup, env *provider.TestEnvironment) {
	// Loop through all of the operators, run preflight, and set their results into their respective object
	for _, op := range env.Operators {
		// Note: We are not using a cache here for the operator bundle images because
		// in-general you are only going to have an operator installed once in a cluster.
		err := op.SetPreflightResults(env)
		if err != nil {
			log.Fatal("Failed running Preflight on operator %q,  err: %v", op.Name, err)
		}
	}

	log.Info("Completed running Preflight operator tests for %d operators", len(env.Operators))

	// Handle Operator-based preflight tests
	// Note: We only care about the `testEntry` variable below because we need its 'Description' and 'Suggestion' variables.
	for testName, testEntry := range getUniqueTestEntriesFromOperatorResults(env.Operators) {
		log.Info("Setting Preflight operator test results for %q", testName)
		generatePreflightOperatorCnfCertTest(checksGroup, testName, testEntry.Description, testEntry.Remediation, env.Operators)
	}
}

// testPreflightContainers runs Preflight checks on all containers in the test environment
//
// The function iterates over each container, executing Preflight diagnostics
// while caching results per image to avoid duplicate work. It logs any errors
// encountered during execution and records completion of tests for the entire
// set. After processing, it aggregates unique test entries from container
// results and generates corresponding checks in the provided group.
func testPreflightContainers(checksGroup *checksdb.ChecksGroup, env *provider.TestEnvironment) {
	// Using a cache to prevent unnecessary processing of images if we already have the results available
	preflightImageCache := make(map[string]provider.PreflightResultsDB)

	// Loop through all of the containers, run preflight, and set their results into their respective objects
	for _, cut := range env.Containers {
		err := cut.SetPreflightResults(preflightImageCache, env)
		if err != nil {
			log.Fatal("Failed running Preflight on image %q, err: %v", cut.Image, err)
		}
	}

	log.Info("Completed running Preflight container tests for %d containers", len(env.Containers))

	// Handle Container-based preflight tests
	// Note: We only care about the `testEntry` variable below because we need its 'Description' and 'Suggestion' variables.
	for testName, testEntry := range getUniqueTestEntriesFromContainerResults(env.Containers) {
		log.Info("Setting Preflight container test results for %q", testName)
		generatePreflightContainerCnfCertTest(checksGroup, testName, testEntry.Description, testEntry.Remediation, env.Containers)
	}
}

// generatePreflightContainerCnfCertTest Creates a test entry for each Preflight container check
//
// The function registers a catalog entry using the supplied name, description,
// and remediation, then adds a corresponding check to the checks group. For
// every container passed in, it examines preflight results and records which
// containers passed, failed, or errored on that specific test. The outcome is
// stored as compliant or non‑compliant objects within the check's result.
func generatePreflightContainerCnfCertTest(checksGroup *checksdb.ChecksGroup, testName, description, remediation string, containers []*provider.Container) {
	// Based on a single test "name", we will be passing/failing in our test framework.
	// Brute force-ish type of method.

	// Store the test names into the Catalog map for results to be dynamically printed
	aID := identifiers.AddCatalogEntry(testName, common.PreflightTestKey, description, remediation, "", "", false, map[string]string{
		identifiers.FarEdge:  identifiers.Optional,
		identifiers.Telco:    identifiers.Optional,
		identifiers.NonTelco: identifiers.Optional,
		identifiers.Extended: identifiers.Optional,
	}, identifiers.TagPreflight)

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(aID)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(check *checksdb.Check) error {
			var compliantObjects []*testhelper.ReportObject
			var nonCompliantObjects []*testhelper.ReportObject
			for _, cut := range containers {
				for _, r := range cut.PreflightResults.Passed {
					if r.Name == testName {
						check.LogInfo("Container %q has passed Preflight test %q", cut, testName)
						compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has passed preflight test "+testName, true))
					}
				}
				for _, r := range cut.PreflightResults.Failed {
					if r.Name == testName {
						check.LogError("Container %q has failed Preflight test %q", cut, testName)
						nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has failed preflight test "+testName, false))
					}
				}
				for _, r := range cut.PreflightResults.Errors {
					if r.Name == testName {
						check.LogError("Container %q has errored Preflight test %q, err: %v", cut, testName, r.Error)
						nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, fmt.Sprintf("Container has errored preflight test %s, err: %v", testName, r.Error), false))
					}
				}
			}

			check.SetResult(compliantObjects, nonCompliantObjects)
			return nil
		}))
}

// generatePreflightOperatorCnfCertTest Creates a test case that aggregates preflight results across operators
//
// The function registers a new test in the catalog, then builds a check that
// iterates over all operators to collect passed, failed, or errored preflight
// outcomes for a given test name. It constructs report objects for each
// operator and sets the overall result accordingly. The check is skipped if no
// operators are present.
func generatePreflightOperatorCnfCertTest(checksGroup *checksdb.ChecksGroup, testName, description, remediation string, operators []*provider.Operator) {
	// Based on a single test "name", we will be passing/failing in our test framework.
	// Brute force-ish type of method.

	// Store the test names into the Catalog map for results to be dynamically printed
	aID := identifiers.AddCatalogEntry(testName, common.PreflightTestKey, description, remediation, "", "", false, map[string]string{
		identifiers.FarEdge:  identifiers.Optional,
		identifiers.Telco:    identifiers.Optional,
		identifiers.NonTelco: identifiers.Optional,
		identifiers.Extended: identifiers.Optional,
	}, identifiers.TagPreflight)

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(aID)).
		WithSkipCheckFn(testhelper.GetNoOperatorsSkipFn(&env)).
		WithCheckFn(func(check *checksdb.Check) error {
			var compliantObjects []*testhelper.ReportObject
			var nonCompliantObjects []*testhelper.ReportObject

			for _, op := range operators {
				for _, r := range op.PreflightResults.Passed {
					if r.Name == testName {
						check.LogInfo("Operator %q has passed Preflight test %q", op, testName)
						compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(op.Namespace, op.Name, "Operator passed preflight test "+testName, true))
					}
				}
				for _, r := range op.PreflightResults.Failed {
					if r.Name == testName {
						check.LogError("Operator %q has failed Preflight test %q", op, testName)
						nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(op.Namespace, op.Name, "Operator failed preflight test "+testName, false))
					}
				}
				for _, r := range op.PreflightResults.Errors {
					if r.Name == testName {
						check.LogError("Operator %q has errored Preflight test %q, err: %v", op, testName, r.Error)
						nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(op.Namespace, op.Name, fmt.Sprintf("Operator has errored preflight test %s, err: %v", testName, r.Error), false))
					}
				}
			}

			check.SetResult(compliantObjects, nonCompliantObjects)
			return nil
		}))
}

// getUniqueTestEntriesFromContainerResults Collects unique preflight test results from multiple containers
//
// This function iterates over a slice of container objects, extracting all
// passed, failed, and error preflight tests. It aggregates them into a map
// keyed by test name, ensuring that duplicate entries are overridden with the
// most recent result. The resulting map contains one entry per unique test
// across all containers.
func getUniqueTestEntriesFromContainerResults(containers []*provider.Container) map[string]provider.PreflightTest {
	// If containers are sharing the same image, they should "presumably" have the same results returned from Preflight.
	testEntries := make(map[string]provider.PreflightTest)
	for _, cut := range containers {
		for _, r := range cut.PreflightResults.Passed {
			testEntries[r.Name] = r
		}
		// Failed Results have more information than the rest
		for _, r := range cut.PreflightResults.Failed {
			testEntries[r.Name] = r
		}
		for _, r := range cut.PreflightResults.Errors {
			testEntries[r.Name] = r
		}
	}

	return testEntries
}

// getUniqueTestEntriesFromOperatorResults collects unique preflight test results from all operators
//
// The function iterates over a slice of operator objects, extracting each
// passed, failed, or errored test result. For every test name it stores the
// corresponding test entry in a map, ensuring that only one instance per test
// name is kept even if multiple operators report the same test. The resulting
// map associates test names with their detailed preflight test information for
// later use.
func getUniqueTestEntriesFromOperatorResults(operators []*provider.Operator) map[string]provider.PreflightTest {
	testEntries := make(map[string]provider.PreflightTest)
	for _, op := range operators {
		for _, r := range op.PreflightResults.Passed {
			testEntries[r.Name] = r
		}
		// Failed Results have more information than the rest
		for _, r := range op.PreflightResults.Failed {
			testEntries[r.Name] = r
		}
		for _, r := range op.PreflightResults.Errors {
			testEntries[r.Name] = r
		}
	}
	return testEntries
}
