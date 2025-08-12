package testcases

import (
	"fmt"
	"sort"

	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
)

// TcResultsSummary represents a summary of test case results.
//
// It contains three integer counters:
// - Failed counts the number of failed test cases.
// - Passed counts the number of passed test cases.
// - Skipped counts the number of skipped test cases.
type TcResultsSummary struct {
	Passed  int
	Skipped int
	Failed  int
}

// TcResultDifference represents a difference between two claim results.
//
// It holds the names of the claims compared, as well as their individual result strings.
// The struct is used to report discrepancies during test case comparisons.
type TcResultDifference struct {
	Name         string
	Claim1Result string
	Claim2Result string
}

// DiffReport holds the results summary and the list of test cases whose result is different.
//
// It aggregates summary statistics from two claim files and identifies
// which individual test cases differ between them. The struct contains
// summaries for each claim, a count of differing test cases, and a slice
// detailing each difference. This information is used by the String method
// to produce a human‑readable comparison report.
type DiffReport struct {
	Claim1ResultsSummary TcResultsSummary `json:"claimFile1ResultsSummary"`
	Claim2ResultsSummary TcResultsSummary `json:"claimFile2ResultsSummary"`

	TestCases                 []TcResultDifference `json:"resultsDifferences"`
	DifferentTestCasesResults int                  `json:"differentTestCasesResults"`
}

// getTestCasesResultsMap creates a map from test case names to their results.
//
// It iterates over the TestSuiteResults argument, which contains a mapping of
// test suite names to slices of individual test case results, and builds a new
// map where each key is a test case name and the value is its corresponding
// result string. The returned map can be used for quick lookup of test case
// outcomes by name.
func getTestCasesResultsMap(testSuiteResults claim.TestSuiteResults) map[string]string {
	testCaseResults := map[string]string{}

	for testCase := range testSuiteResults {
		testCaseResults[testSuiteResults[testCase].TestID.ID] = testSuiteResults[testCase].State
	}

	return testCaseResults
}

// getMergedTestCasesNames returns a sorted slice of all test case names found in two maps, without duplicates.
//
// It accepts two maps whose keys are test case names and produces a slice containing each unique name from both maps.
// The resulting slice is sorted alphabetically. Duplicate names that appear in both input maps are included only once.
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

// getTestCasesResultsSummary generates a summary of test case results.
//
// It accepts a map where keys are test case names and values are the corresponding
// results. The function creates and returns a TcResultsSummary struct populated
// with this data, providing a convenient way to convert raw result maps into
// structured summaries for further processing.
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

// GetDiffReport compares the results from two claim files and returns a DiffReport.
//
// It takes two TestSuiteResults values, one from each claim file, and produces a report
// summarizing differences between them. For any test case that exists in only one of the
// inputs, the report marks its status as "not found". The function builds maps of results,
// merges test case names, aggregates summaries, and assembles the final DiffReport structure.
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

// String returns a formatted summary of the diff report.
//
// It builds two tables: one summarizing test case counts per status
// across CLAIM-1 and CLAIM-2, and another listing individual test cases
// that differ between the two claims along with their statuses.
// The returned string is suitable for printing or logging.
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
