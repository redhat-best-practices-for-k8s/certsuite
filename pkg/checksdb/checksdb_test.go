package checksdb

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func saveAndResetDBState(t *testing.T) {
	t.Helper()
	origResultsDB := resultsDB
	origDbByGroup := dbByGroup
	origEvaluator := labelsExprEvaluator
	t.Cleanup(func() {
		resultsDB = origResultsDB
		dbByGroup = origDbByGroup
		labelsExprEvaluator = origEvaluator
	})
	resultsDB = map[string]claim.Result{}
	dbByGroup = map[string]*ChecksGroup{}
}

func TestGetResults(t *testing.T) {
	saveAndResetDBState(t)

	results := GetResults()
	assert.Empty(t, results)

	resultsDB["check-1"] = claim.Result{State: CheckResultPassed}
	resultsDB["check-2"] = claim.Result{State: CheckResultFailed}

	results = GetResults()
	assert.Len(t, results, 2)
	assert.Equal(t, CheckResultPassed, results["check-1"].State)
	assert.Equal(t, CheckResultFailed, results["check-2"].State)
}

func TestGetTotalTests(t *testing.T) {
	saveAndResetDBState(t)

	assert.Equal(t, 0, GetTotalTests())

	resultsDB["check-1"] = claim.Result{State: CheckResultPassed}
	resultsDB["check-2"] = claim.Result{State: CheckResultFailed}
	resultsDB["check-3"] = claim.Result{State: CheckResultSkipped}

	assert.Equal(t, 3, GetTotalTests())
}

func TestGetTestsCountByState(t *testing.T) {
	saveAndResetDBState(t)

	resultsDB["check-1"] = claim.Result{State: CheckResultPassed}
	resultsDB["check-2"] = claim.Result{State: CheckResultPassed}
	resultsDB["check-3"] = claim.Result{State: CheckResultFailed}
	resultsDB["check-4"] = claim.Result{State: CheckResultSkipped}

	assert.Equal(t, 2, GetTestsCountByState(CheckResultPassed))
	assert.Equal(t, 1, GetTestsCountByState(CheckResultFailed))
	assert.Equal(t, 1, GetTestsCountByState(CheckResultSkipped))
	assert.Equal(t, 0, GetTestsCountByState(CheckResultError))
}

func TestGetReconciledResults(t *testing.T) {
	saveAndResetDBState(t)

	resultsDB["check-1"] = claim.Result{State: CheckResultPassed}
	resultsDB["check-2"] = claim.Result{State: CheckResultFailed}

	reconciled := GetReconciledResults()
	assert.Len(t, reconciled, 2)
	assert.Equal(t, CheckResultPassed, reconciled["check-1"].State)
	assert.Equal(t, CheckResultFailed, reconciled["check-2"].State)

	// Verify it's a copy by modifying the returned map
	reconciled["check-3"] = claim.Result{State: CheckResultSkipped}
	assert.Len(t, resultsDB, 2)
}

func TestGetTestSuites(t *testing.T) {
	saveAndResetDBState(t)

	suites := GetTestSuites()
	assert.Empty(t, suites)

	resultsDB["check-1"] = claim.Result{State: CheckResultPassed}
	resultsDB["check-2"] = claim.Result{State: CheckResultFailed}

	suites = GetTestSuites()
	assert.Len(t, suites, 2)
	assert.ElementsMatch(t, []string{"check-1", "check-2"}, suites)
}

func TestInitLabelsExprEvaluator(t *testing.T) {
	saveAndResetDBState(t)

	tests := []struct {
		name      string
		filter    string
		expectErr bool
	}{
		{
			name:      "valid single label",
			filter:    "common",
			expectErr: false,
		},
		{
			name:      "valid comma-separated labels",
			filter:    "common,extended",
			expectErr: false,
		},
		{
			name:      "all expands to all tags",
			filter:    "all",
			expectErr: false,
		},
		{
			name:      "valid boolean expression",
			filter:    "common && !extended",
			expectErr: false,
		},
		{
			name:      "invalid expression",
			filter:    "&&&&",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitLabelsExprEvaluator(tt.filter)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, labelsExprEvaluator)
			}
		})
	}
}

func TestFilterCheckIDs(t *testing.T) {
	saveAndResetDBState(t)

	group := &ChecksGroup{
		name: "test-group",
		checks: []*Check{
			NewCheck("check-common", []string{"common"}),
			NewCheck("check-extended", []string{"extended"}),
			NewCheck("check-telco", []string{"telco"}),
		},
	}
	dbByGroup["test-group"] = group

	err := InitLabelsExprEvaluator("common")
	require.NoError(t, err)

	ids, err := FilterCheckIDs()
	require.NoError(t, err)
	assert.Equal(t, []string{"check-common"}, ids)

	err = InitLabelsExprEvaluator("common,extended")
	require.NoError(t, err)

	ids, err = FilterCheckIDs()
	require.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, "check-common")
	assert.Contains(t, ids, "check-extended")
}

func TestGetResultsSummary(t *testing.T) {
	saveAndResetDBState(t)

	net1 := NewCheck("net-1", []string{"test"})
	net1.Result = CheckResultPassed
	net2 := NewCheck("net-2", []string{"test"})
	net2.Result = CheckResultFailed
	net3 := NewCheck("net-3", []string{"test"})
	net3.Result = CheckResultSkipped
	net4 := NewCheck("net-4", []string{"test"})
	net4.Result = CheckResultPassed

	group := &ChecksGroup{
		name:   "networking",
		checks: []*Check{net1, net2, net3, net4},
	}
	dbByGroup["networking"] = group

	summary := getResultsSummary()
	require.Contains(t, summary, "networking")
	assert.Equal(t, 2, summary["networking"][PASSED])
	assert.Equal(t, 1, summary["networking"][FAILED])
	assert.Equal(t, 1, summary["networking"][SKIPPED])
}

func TestRecordCheckResultNotFound(t *testing.T) {
	saveAndResetDBState(t)

	check := NewCheck("non-existent-check-id", []string{"test"})
	recordCheckResult(check)

	assert.Empty(t, resultsDB)
}

func TestInitLabelsExprEvaluatorEval(t *testing.T) {
	saveAndResetDBState(t)

	err := InitLabelsExprEvaluator("common")
	require.NoError(t, err)

	assert.True(t, labelsExprEvaluator.Eval([]string{"common"}))
	assert.False(t, labelsExprEvaluator.Eval([]string{"extended"}))
	assert.True(t, labelsExprEvaluator.Eval([]string{"common", "extended"}))
}
