package testcases

import (
	"reflect"
	"testing"

	"github.com/test-network-function/cnf-certification-test/cmd/tnf/pkg/claim"
	"gotest.tools/v3/assert"
)

func TestGetTestCasesResultsMap(t *testing.T) {
	testCases := []struct {
		description                 string
		results                     claim.TestSuiteResults
		expectedTestCasesResultsMap map[string]string
	}{
		{
			description:                 "nil input map",
			results:                     nil,
			expectedTestCasesResultsMap: map[string]string{},
		},
		{
			description:                 "empty input map",
			results:                     claim.TestSuiteResults{},
			expectedTestCasesResultsMap: map[string]string{},
		},
		{
			description: "one test case in the access-control ts",
			results: claim.TestSuiteResults{
				"access-control": claim.TestCaseResult{
					TestID: claim.TestCaseID{
						ID: "access-control-ssh-daemons",
					},
					State: "skipped",
				},
			},
			expectedTestCasesResultsMap: map[string]string{
				"access-control-ssh-daemons": "skipped",
			},
		},
		{
			description: "two test suites with two test cases each",
			results: claim.TestSuiteResults{
				"access-control-ssh-daemons": claim.TestCaseResult{
					TestID: claim.TestCaseID{
						ID: "access-control-ssh-daemons",
					},
					State: "skipped",
				},
				"access-control-sys-admin-capability-check": claim.TestCaseResult{
					TestID: claim.TestCaseID{
						ID: "access-control-sys-admin-capability-check",
					},
					State: "passed",
				},
				"lifecycle-pod-scheduling": claim.TestCaseResult{
					TestID: claim.TestCaseID{
						ID: "lifecycle-pod-scheduling",
					},
					State: "skipped",
				},
				"lifecycle-pod-high-availability": claim.TestCaseResult{
					TestID: claim.TestCaseID{
						ID: "lifecycle-pod-high-availability",
					},
					State: "failed",
				},
			},
			expectedTestCasesResultsMap: map[string]string{
				"access-control-ssh-daemons":                "skipped",
				"access-control-sys-admin-capability-check": "passed",
				"lifecycle-pod-scheduling":                  "skipped",
				"lifecycle-pod-high-availability":           "failed",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			resultsMap := getTestCasesResultsMap(tc.results)
			assert.Equal(t, true, reflect.DeepEqual(tc.expectedTestCasesResultsMap, resultsMap))
		})
	}
}

func TestGetMergedTestCasesNames(t *testing.T) {
	testCases := []struct {
		description           string
		claim1Results         map[string]string
		claim2Results         map[string]string
		expectedMergedTcNames []string
	}{
		{
			description:           "nil maps 1",
			claim1Results:         nil,
			claim2Results:         nil,
			expectedMergedTcNames: []string{},
		},
		{
			description:           "nil maps 2",
			claim1Results:         nil,
			claim2Results:         map[string]string{},
			expectedMergedTcNames: []string{},
		},
		{
			description:           "nil maps 3",
			claim1Results:         map[string]string{},
			claim2Results:         nil,
			expectedMergedTcNames: []string{},
		},
		{
			description:           "first empty but second with two results",
			claim1Results:         nil,
			claim2Results:         map[string]string{"tc1": "passed", "tc2": "failed"},
			expectedMergedTcNames: []string{"tc1", "tc2"},
		},
		{
			description:           "second empty but first with two results",
			claim1Results:         map[string]string{"tc1": "passed", "tc2": "failed"},
			claim2Results:         nil,
			expectedMergedTcNames: []string{"tc1", "tc2"},
		},
		{
			description:           "merging two different results names lists",
			claim1Results:         map[string]string{"tc1": "passed", "tc3": "failed"},
			claim2Results:         map[string]string{"tc2": "passed", "tc4": "failed"},
			expectedMergedTcNames: []string{"tc1", "tc2", "tc3", "tc4"},
		},
		{
			description:           "merging two maps with one common tc name tc2",
			claim1Results:         map[string]string{"tc2": "passed", "tc3": "failed", "tc4": "skipped"},
			claim2Results:         map[string]string{"tc1": "passed", "tc2": "failed"},
			expectedMergedTcNames: []string{"tc1", "tc2", "tc3", "tc4"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			tcMergedNamesList := getMergedTestCasesNames(tc.claim1Results, tc.claim2Results)
			assert.Equal(t, true, reflect.DeepEqual(tc.expectedMergedTcNames, tcMergedNamesList))
		})
	}
}

func TestGetTestCasesResultsSummary(t *testing.T) {
	testCases := []struct {
		description     string
		results         map[string]string
		expectedSummary TcResultsSummary
	}{
		{
			description:     "nil map",
			results:         nil,
			expectedSummary: TcResultsSummary{},
		},
		{
			description:     "empty map",
			results:         map[string]string{},
			expectedSummary: TcResultsSummary{},
		},
		{
			description: "map with one passed tc",
			results:     map[string]string{"tc1": "passed"},
			expectedSummary: TcResultsSummary{
				Passed:  1,
				Skipped: 0,
				Failed:  0,
			},
		},
		{
			description: "map with one tc of each result type",
			results:     map[string]string{"tc1": "passed", "tc2": "skipped", "tc3": "failed"},
			expectedSummary: TcResultsSummary{
				Passed:  1,
				Skipped: 1,
				Failed:  1,
			},
		},
		{
			description: "map with one passing tcs, two skipped tcs and three failed tcs",
			results:     map[string]string{"tc1": "passed", "tc2": "skipped", "tc3": "skipped", "tc4": "failed", "tc5": "failed", "tc6": "failed"},
			expectedSummary: TcResultsSummary{
				Passed:  1,
				Skipped: 2,
				Failed:  3,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			summary := getTestCasesResultsSummary(tc.results)
			assert.Equal(t, tc.expectedSummary, summary)
		})
	}
}

func TestGetDiffReport(t *testing.T) {
	testCases := []struct {
		description        string
		results1           claim.TestSuiteResults
		results2           claim.TestSuiteResults
		expectedDiffReport DiffReport
	}{
		{
			description: "results1 empty, results2 with one tc result",
			results1:    map[string]claim.TestCaseResult{},
			results2: map[string]claim.TestCaseResult{
				"access-control": {TestID: claim.TestCaseID{ID: "access-control-ssh-daemons"}, State: "passed"},
			},
			expectedDiffReport: DiffReport{
				Claim1ResultsSummary: TcResultsSummary{},
				Claim2ResultsSummary: TcResultsSummary{Passed: 1},
				TestCases: []TcResultDifference{
					{
						Name:         "access-control-ssh-daemons",
						Claim1Result: "not found",
						Claim2Result: "passed",
					},
				},
				DifferentTestCasesResults: 1,
			},
		},
		{
			description: "results1 and results2 have the same passing tc",
			results1: map[string]claim.TestCaseResult{
				"access-control": {TestID: claim.TestCaseID{ID: "access-control-ssh-daemons"}, State: "passed"},
			},
			results2: map[string]claim.TestCaseResult{
				"access-control": {TestID: claim.TestCaseID{ID: "access-control-ssh-daemons"}, State: "passed"},
			},
			expectedDiffReport: DiffReport{
				Claim1ResultsSummary:      TcResultsSummary{Passed: 1},
				Claim2ResultsSummary:      TcResultsSummary{Passed: 1},
				TestCases:                 []TcResultDifference{},
				DifferentTestCasesResults: 0,
			},
		},
		{
			description: "results1 and results2 have same tc with different result",
			results1: map[string]claim.TestCaseResult{
				"access-control": {TestID: claim.TestCaseID{ID: "access-control-ssh-daemons"}, State: "passed"},
			},
			results2: map[string]claim.TestCaseResult{
				"access-control": {TestID: claim.TestCaseID{ID: "access-control-ssh-daemons"}, State: "failed"},
			},
			expectedDiffReport: DiffReport{
				Claim1ResultsSummary:      TcResultsSummary{Passed: 1},
				Claim2ResultsSummary:      TcResultsSummary{Failed: 1},
				TestCases:                 []TcResultDifference{{Name: "access-control-ssh-daemons", Claim1Result: "passed", Claim2Result: "failed"}},
				DifferentTestCasesResults: 1,
			},
		},
		{
			description: "results1 and results2 have the same two tcs from different test suites, both with different results",
			results1: map[string]claim.TestCaseResult{
				"access-control": {TestID: claim.TestCaseID{ID: "access-control-ssh-daemons"}, State: "passed"},
				"lifecycle":      {TestID: claim.TestCaseID{ID: "lifecycle-pod-scheduling"}, State: "failed"},
			},
			results2: map[string]claim.TestCaseResult{
				"access-control": {TestID: claim.TestCaseID{ID: "access-control-ssh-daemons"}, State: "failed"},
				"lifecycle":      {TestID: claim.TestCaseID{ID: "lifecycle-pod-scheduling"}, State: "passed"},
			},
			expectedDiffReport: DiffReport{
				Claim1ResultsSummary: TcResultsSummary{Passed: 1, Failed: 1},
				Claim2ResultsSummary: TcResultsSummary{Passed: 1, Failed: 1},
				TestCases: []TcResultDifference{
					{Name: "access-control-ssh-daemons", Claim1Result: "passed", Claim2Result: "failed"},
					{Name: "lifecycle-pod-scheduling", Claim1Result: "failed", Claim2Result: "passed"},
				},
				DifferentTestCasesResults: 2,
			},
		},
		{
			description: "one same test case result and another different",
			results1: map[string]claim.TestCaseResult{
				"access-control": {TestID: claim.TestCaseID{ID: "access-control-ssh-daemons"}, State: "passed"},
				"lifecycle":      {TestID: claim.TestCaseID{ID: "lifecycle-pod-scheduling"}, State: "failed"},
			},
			results2: map[string]claim.TestCaseResult{
				"access-control": {TestID: claim.TestCaseID{ID: "access-control-ssh-daemons"}, State: "passed"},
				"lifecycle":      {TestID: claim.TestCaseID{ID: "lifecycle-pod-scheduling"}, State: "skipped"},
			},
			expectedDiffReport: DiffReport{
				Claim1ResultsSummary: TcResultsSummary{Passed: 1, Failed: 1},
				Claim2ResultsSummary: TcResultsSummary{Passed: 1, Skipped: 1},
				TestCases: []TcResultDifference{
					{Name: "lifecycle-pod-scheduling", Claim1Result: "failed", Claim2Result: "skipped"},
				},
				DifferentTestCasesResults: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			diffReport := GetDiffReport(tc.results1, tc.results2)

			// Check summaries
			assert.Equal(t, tc.expectedDiffReport.Claim1ResultsSummary, diffReport.Claim1ResultsSummary)
			assert.Equal(t, tc.expectedDiffReport.Claim2ResultsSummary, diffReport.Claim2ResultsSummary)

			// Check test case results differences
			t.Logf("diffs: %+v", diffReport.TestCases)
			assert.Equal(t, true, reflect.DeepEqual(tc.expectedDiffReport.TestCases, diffReport.TestCases))

			// Check count
			assert.Equal(t, tc.expectedDiffReport.DifferentTestCasesResults, diffReport.DifferentTestCasesResults)
		})
	}
}
