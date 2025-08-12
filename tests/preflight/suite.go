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

// labelsAllowTestRun determines whether a test should be executed based on provided labels.
//
// It accepts a test identifier and a list of allowed labels, returning true if the
// identifier matches any label in the list. The function uses Contains to perform
// the membership check and returns false when no match is found.
func labelsAllowTestRun(labelFilter string, allowedLabels []string) bool {
	for _, label := range allowedLabels {
		if strings.Contains(labelFilter, label) {
			return true
		}
	}
	return false
}

// ShouldRun determines whether preflight checks should be executed.
//
// It returns true if the test environment contains any of the predefined
// preflight tags/labels and the required dockerconfig file exists.
// These conditions prevent unnecessary loading of all preflight checks,
// which can be time consuming. The function relies on the current test
// environment and parameters to make this decision.
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

// LoadChecks registers all preflight checks and returns a cleanup function.
//
// It initializes the test environment, logs diagnostic information,
// and creates check groups for containers and operators.
// The returned function should be called to perform any necessary
// teardown after the tests finish.
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

// testPreflightOperators verifies preflight operator results and records them.
//
// It receives a ChecksGroup and a TestEnvironment, evaluates the
// operator results for each check, logs information about missing
// entries, generates CNF certificate tests for any unique operators,
// and stores the aggregated results in the checks group.  
// If an unexpected condition occurs during processing it aborts the test suite.
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

// testPreflightContainers runs preflight container checks for a given ChecksGroup.
//
// It receives a pointer to a ChecksGroup and a TestEnvironment, executes the
// configured container tests, collects results, and stores them back into
// the group via SetPreflightResults. If any check fails, it logs the failure
// with Fatal; otherwise it logs progress using Info.
// The function does not return a value.
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

// generatePreflightContainerCnfCertTest creates a preflight test that verifies certificate configuration for a set of containers.
//
// It registers the test with the checks database, attaches metadata such as name, ID and tags, and supplies functions to execute the check and to skip it when no containers are present. The test function iterates over each container, logs relevant information, runs the certification logic, and records success or failure in a report object for that container.
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

// generatePreflightOperatorCnfCertTest creates a preflight test for validating operator CNF certificates.
//
// It constructs a ChecksGroup entry that verifies each provided Operator's
// certificate configuration against expected values such as CA, key usage,
// and SANs. The function accepts the checks group to populate, identifiers
// for the test ID and labels, and a slice of Operators to test.
// No return value; it registers the test via AddCatalogEntry on the ChecksGroup.
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

// getUniqueTestEntriesFromContainerResults extracts unique preflight test entries from a slice of container results.
//
// It accepts a slice of Container pointers and returns a map keyed by test identifiers,
// each mapping to the corresponding PreflightTest object. Duplicate tests are merged
// into a single entry in the returned map. The function constructs the map using make
// and populates it by iterating over the provided containers.
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

// getUniqueTestEntriesFromOperatorResults extracts unique preflight test entries from a slice of operator results.
//
// It takes a slice of Operator pointers and returns a map keyed by test identifiers,
// where each value is the corresponding PreflightTest.
// The function ensures that only distinct test entries are included in the resulting map.
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
