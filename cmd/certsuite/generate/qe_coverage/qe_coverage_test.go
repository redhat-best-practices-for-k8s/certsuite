package qecoverage

import (
	"sort"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/stretchr/testify/assert"
)

type catalogEntry struct {
	testSuiteName string
	testCaseName  string
	qe            bool
}

func createCatalog(entries []catalogEntry) map[claim.Identifier]claim.TestCaseDescription {
	catalog := map[claim.Identifier]claim.TestCaseDescription{}

	for _, entry := range entries {
		tcDescription, aID := claim.BuildTestCaseDescription(entry.testCaseName, entry.testSuiteName, "", "", "", "", entry.qe, map[string]string{})
		catalog[aID] = tcDescription
	}

	return catalog
}

func TestQeCoverage(t *testing.T) {
	type testSuiteCoverage struct {
		coverage  float32
		tcs       int
		tcsWithQe int
	}

	tcs := []struct {
		catalogEntries         []catalogEntry
		expectedTotalCoverage  float32
		expectedTotalTcs       int
		expectedTotalTcswithQe int
		expectedCoverageByTS   map[string]testSuiteCoverage
	}{
		// Corner case: no test cases.
		{
			catalogEntries:         []catalogEntry{},
			expectedTotalCoverage:  0.0,
			expectedTotalTcs:       0,
			expectedTotalTcswithQe: 0,
			expectedCoverageByTS:   map[string]testSuiteCoverage{},
		},

		// 1 test suite with, 1 test case with QE flag set.
		{
			catalogEntries: []catalogEntry{
				{
					testSuiteName: "test-suite-1",
					testCaseName:  "test-suite-1-tc-1",
					qe:            true,
				},
			},
			expectedTotalCoverage:  100.0,
			expectedTotalTcs:       1,
			expectedTotalTcswithQe: 1,
			expectedCoverageByTS: map[string]testSuiteCoverage{
				"test-suite-1": {
					coverage:  100.0,
					tcs:       1,
					tcsWithQe: 1,
				},
			},
		},
		// 1 test suite with 2 test cases, but only 1 with QE flag set.
		{
			catalogEntries: []catalogEntry{
				{
					testSuiteName: "test-suite-1",
					testCaseName:  "test-suite-1-tc-1",
					qe:            false,
				},
				{
					testSuiteName: "test-suite-1",
					testCaseName:  "test-suite-1-tc-2",
					qe:            true,
				},
			},
			expectedTotalCoverage:  50.0,
			expectedTotalTcs:       2,
			expectedTotalTcswithQe: 1,
			expectedCoverageByTS: map[string]testSuiteCoverage{
				"test-suite-1": {
					coverage:  50.0,
					tcs:       1,
					tcsWithQe: 1,
				},
			},
		},
		// 3 test suites: 2 of them with 1 test case not in QE. The third has
		// 2 test cases, only 1 of them in QE.
		{
			catalogEntries: []catalogEntry{
				{
					testSuiteName: "test-suite-1",
					testCaseName:  "test-suite-1-tc-1",
					qe:            false,
				},
				{
					testSuiteName: "test-suite-2",
					testCaseName:  "test-suite-2-tc-1",
					qe:            false,
				},
				{
					testSuiteName: "test-suite-3",
					testCaseName:  "test-suite-3-tc-1",
					qe:            false,
				},
				{
					testSuiteName: "test-suite-3",
					testCaseName:  "test-suite-3-tc-2",
					qe:            true,
				},
			},
			expectedTotalCoverage:  25.0,
			expectedTotalTcs:       4,
			expectedTotalTcswithQe: 1,
			expectedCoverageByTS: map[string]testSuiteCoverage{
				"test-suite-1": {
					coverage:  0.0,
					tcs:       1,
					tcsWithQe: 0,
				},
				"test-suite-2": {
					coverage:  0.0,
					tcs:       1,
					tcsWithQe: 0,
				},
				"test-suite-3": {
					coverage:  50.0,
					tcs:       2,
					tcsWithQe: 1,
				},
			},
		},
	}

	for _, tc := range tcs {
		catalog := createCatalog(tc.catalogEntries)
		qeCoverage := GetQeCoverage(catalog)
		assert.Equal(t, tc.expectedTotalTcs, qeCoverage.TestCasesTotal)
		assert.Equal(t, tc.expectedTotalTcswithQe, qeCoverage.TestCasesWithQe)
		assert.Equal(t, tc.expectedTotalCoverage, qeCoverage.TotalCoveragePercentage)

		// Sort expectd/actual test suites by name before comparing test suite coverages
		expectedTestSuites := []string{}
		for testSuite := range tc.expectedCoverageByTS {
			expectedTestSuites = append(expectedTestSuites, testSuite)
		}
		sort.Strings(expectedTestSuites)

		actualTestSuites := []string{}
		for testSuite := range qeCoverage.CoverageByTestSuite {
			actualTestSuites = append(actualTestSuites, testSuite)
		}
		sort.Strings(actualTestSuites)

		assert.Equal(t, expectedTestSuites, actualTestSuites)
		for _, expectedTestSuite := range expectedTestSuites {
			assert.Equal(t, tc.expectedCoverageByTS[expectedTestSuite].coverage, qeCoverage.CoverageByTestSuite[expectedTestSuite].Coverage)
		}
	}
}
