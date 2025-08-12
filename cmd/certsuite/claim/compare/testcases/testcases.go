package testcases

import (
	"fmt"
	"sort"

	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
)

type TcResultsSummary struct {
	Passed  int
	Skipped int
	Failed  int
}

type TcResultDifference struct {
	Name         string
	Claim1Result string
	Claim2Result string
}

// Holds the results summary and the list of test cases whose result
// is different.
type DiffReport struct {
	Claim1ResultsSummary TcResultsSummary `json:"claimFile1ResultsSummary"`
	Claim2ResultsSummary TcResultsSummary `json:"claimFile2ResultsSummary"`

	TestCases                 []TcResultDifference `json:"resultsDifferences"`
	DifferentTestCasesResults int                  `json:"differentTestCasesResults"`
}

// Helper function that iterates over resultsByTestSuite, which maps a test suite name to a list
// of test case results, to create a map with test case results.
func getTestCasesResultsMap(testSuiteResults claim.TestSuiteResults) map[string]string {
	testCaseResults := map[string]string{}

	for testCase := range testSuiteResults {
		testCaseResults[testSuiteResults[testCase].TestID.ID] = testSuiteResults[testCase].State
	}

	return testCaseResults
}

// Given two results helper maps whose keys are test case names, returns a slice of
// all the test cases (sorted) names found in both maps, without repetitions.
// If one test case appears in both map, it will only appear once in the
// output slice.
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

// Helper function to fill a TcResultsSummary struct from a results map (tc name -> result).
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

// Process the results from different claim files and return the DiffReport.
// In case one tc name does not exist in the other claim file, the result will
// be marked as "not found" in the table.
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

// Stringer method for the DiffReport. Will return a string with two tables:
// Test cases summary table:
// STATUS         # in CLAIM-1        # in CLAIM-2
// passed         22                  21
// skipped        62                  62
// failed         3                   4
//
// Test cases with different results table:
// TEST CASE NAME                                              CLAIM-1   CLAIM-2
// access-control-net-admin-capability-check                   failed    passed
// access-control-pod-automount-service-account-token          passed    failed
// access-control-pod-role-bindings                            passed    failed
// access-control-pod-service-account                          passed    failed
// ...
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
