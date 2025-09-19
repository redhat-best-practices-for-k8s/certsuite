package testcases

import (
	"fmt"
	"sort"

	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
)

// TcResultsSummary provides a count of test case outcomes
//
// This structure holds three integer counters: how many tests passed, were
// skipped, and failed. It is populated by iterating over result strings and
// incrementing the corresponding field. The counts can be used to report
// overall test performance.
type TcResultsSummary struct {
	Passed  int
	Skipped int
	Failed  int
}

// TcResultDifference Represents a discrepancy between two claim results
//
// This structure holds the name of a test case along with the outcomes from two
// different claims. By comparing Claim1Result and Claim2Result, users can
// identify mismatches or confirm consistency across claim evaluations.
type TcResultDifference struct {
	Name         string
	Claim1Result string
	Claim2Result string
}

// DiffReport Summarizes test result differences between two claim files
//
// This structure holds a summary of passed, skipped, and failed tests for each
// claim file, along with a list of individual test cases whose outcomes differ.
// It tracks the total number of differing test cases and provides a string
// representation that lists both the overall status counts and the specific
// differences. The data is used to report and compare results between two sets
// of claim executions.
type DiffReport struct {
	Claim1ResultsSummary TcResultsSummary `json:"claimFile1ResultsSummary"`
	Claim2ResultsSummary TcResultsSummary `json:"claimFile2ResultsSummary"`

	TestCases                 []TcResultDifference `json:"resultsDifferences"`
	DifferentTestCasesResults int                  `json:"differentTestCasesResults"`
}

// getTestCasesResultsMap Creates a map from test case identifiers to their execution state
//
// This helper traverses the provided test suite results, extracting each test
// case's unique ID and its current . It builds a string-to-string mapping where
// keys are the IDs and values are the states. The resulting map is used by
// other functions to compare outcomes between different claim results.
func getTestCasesResultsMap(testSuiteResults claim.TestSuiteResults) map[string]string {
	testCaseResults := map[string]string{}

	for testCase := range testSuiteResults {
		testCaseResults[testSuiteResults[testCase].TestID.ID] = testSuiteResults[testCase].State
	}

	return testCaseResults
}

// getMergedTestCasesNames Collects all unique test case names from two result maps
//
// The function iterates over each input map, adding every key to a temporary
// set to eliminate duplicates. After gathering the keys, it converts the set
// into a slice and sorts the entries alphabetically. The sorted list of test
// case names is returned for further processing.
func getMergedTestCasesNames(results1, results2 map[string]string) []string {
	testCasesNamesMap := map[string]struct{}{}

	for name := range results1 {
		testCasesNamesMap[name] = struct{}{}
	}

	for name := range results2 {
		testCasesNamesMap[name] = struct{}{}
	}

	// get the full list of names and sort it
	names := []string{}
	for name := range testCasesNamesMap {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// getTestCasesResultsSummary Aggregates test case results into a summary count
//
// The function iterates over a mapping of test case names to result strings and
// tallies the number of passed, skipped, and failed cases. It increments
// counters in a TcResultsSummary structure based on predefined result
// constants. The populated summary is then returned for use elsewhere.
func getTestCasesResultsSummary(results map[string]string) TcResultsSummary {
	summary := TcResultsSummary{}

	for _, result := range results {
		switch result {
		case claim.TestCaseResultPassed:
			summary.Passed++
		case claim.TestCaseResultSkipped:
			summary.Skipped++
		case claim.TestCaseResultFailed:
			summary.Failed++
		}
	}

	return summary
}

// GetDiffReport Creates a report of differences between two sets of test results
//
// The function compares test case outcomes from two claim files, marking any
// missing cases as "not found". It builds a list of differing results, counts
// the number of discrepancies, and summarizes each claim’s passed, skipped,
// and failed totals. The returned DiffReport contains this information for
// further analysis.
func GetDiffReport(resultsClaim1, resultsClaim2 claim.TestSuiteResults) *DiffReport {
	const tcResultNotFound = "not found"

	report := DiffReport{}

	claim1Results := getTestCasesResultsMap(resultsClaim1)
	claim2Results := getTestCasesResultsMap(resultsClaim2)

	tcNames := getMergedTestCasesNames(claim1Results, claim2Results)

	report.TestCases = []TcResultDifference{}
	for _, name := range tcNames {
		claim1TcResult, found := claim1Results[name]
		if !found {
			claim1TcResult = tcResultNotFound
		}

		claim2TcResult, found := claim2Results[name]
		if !found {
			claim2TcResult = tcResultNotFound
		}

		if claim1TcResult == claim2TcResult && claim1TcResult != tcResultNotFound {
			continue
		}

		report.TestCases = append(report.TestCases, TcResultDifference{
			Name:         name,
			Claim1Result: claim1TcResult,
			Claim2Result: claim2TcResult,
		})

		report.DifferentTestCasesResults++
	}

	report.Claim1ResultsSummary = getTestCasesResultsSummary(claim1Results)
	report.Claim2ResultsSummary = getTestCasesResultsSummary(claim2Results)

	return &report
}

// DiffReport.String Formats a detailed report of test case comparisons
//
// The method builds a human‑readable string containing two tables: one
// summarizing the count of passed, skipped and failed cases for each claim, and
// another listing individual test cases that differ between the claims. It uses
// formatted printing to align columns and returns the combined text.
func (r *DiffReport) String() string {
	const (
		tcDiffRowFmt          = "%-60s%-10s%-s\n"
		tcStatusSummaryRowFmt = "%-15s%-20s%-s\n"
	)

	str := "RESULTS SUMMARY\n"
	str += "---------------\n"
	str += fmt.Sprintf(tcStatusSummaryRowFmt, "STATUS", "# in CLAIM-1", "# in CLAIM-2")
	str += fmt.Sprintf(tcStatusSummaryRowFmt, "passed", fmt.Sprintf("%d", r.Claim1ResultsSummary.Passed), fmt.Sprintf("%d", r.Claim2ResultsSummary.Passed))
	str += fmt.Sprintf(tcStatusSummaryRowFmt, "skipped", fmt.Sprintf("%d", r.Claim1ResultsSummary.Skipped), fmt.Sprintf("%d", r.Claim2ResultsSummary.Skipped))
	str += fmt.Sprintf(tcStatusSummaryRowFmt, "failed", fmt.Sprintf("%d", r.Claim1ResultsSummary.Failed), fmt.Sprintf("%d", r.Claim2ResultsSummary.Failed))
	str += "\n"

	str += "RESULTS DIFFERENCES\n"
	str += "-------------------\n"
	if len(r.TestCases) == 0 {
		str += "<none>\n"
		return str
	}

	str += fmt.Sprintf(tcDiffRowFmt, "TEST CASE NAME", "CLAIM-1", "CLAIM-2")
	for _, diff := range r.TestCases {
		str += fmt.Sprintf(tcDiffRowFmt, diff.Name, diff.Claim1Result, diff.Claim2Result)
	}

	return str
}
