package checksdb

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChecksGroup(t *testing.T) {
	saveAndResetDBState(t)

	group := NewChecksGroup("test-suite")
	require.NotNil(t, group)
	assert.Equal(t, "test-suite", group.name)
	assert.Empty(t, group.checks)
	assert.Equal(t, checkIdxNone, group.currentRunningCheckIdx)

	// Same name returns existing group
	group2 := NewChecksGroup("test-suite")
	assert.Same(t, group, group2)

	// Different name creates new group
	group3 := NewChecksGroup("other-suite")
	assert.NotSame(t, group, group3)
}

func TestChecksGroupBuilderMethods(t *testing.T) {
	saveAndResetDBState(t)

	group := NewChecksGroup("builder-test")

	beforeAllFn := func(checks []*Check) error { return nil }
	afterAllFn := func(checks []*Check) error { return nil }
	beforeEachFn := func(check *Check) error { return nil }
	afterEachFn := func(check *Check) error { return nil }

	result := group.WithBeforeAllFn(beforeAllFn)
	assert.Same(t, group, result)
	assert.NotNil(t, group.beforeAllFn)

	result = group.WithAfterAllFn(afterAllFn)
	assert.Same(t, group, result)
	assert.NotNil(t, group.afterAllFn)

	result = group.WithBeforeEachFn(beforeEachFn)
	assert.Same(t, group, result)
	assert.NotNil(t, group.beforeEachFn)

	result = group.WithAfterEachFn(afterEachFn)
	assert.Same(t, group, result)
	assert.NotNil(t, group.afterEachFn)
}

func TestAddAndResetChecks(t *testing.T) {
	saveAndResetDBState(t)

	group := NewChecksGroup("add-test")

	check1 := NewCheck("check-1", []string{"test"})
	check2 := NewCheck("check-2", []string{"test"})

	group.Add(check1)
	assert.Len(t, group.checks, 1)

	group.Add(check2)
	assert.Len(t, group.checks, 2)

	group.ResetChecks()
	assert.Empty(t, group.checks)
}

func TestSkipCheckSetsSkipped(t *testing.T) {
	t.Parallel()

	check := NewCheck("skip-test", []string{"test"})
	skipCheck(check, "test reason")
	assert.Equal(t, CheckResultSkipped, check.Result.String())
}

func TestSkipAllSetsAllSkipped(t *testing.T) {
	t.Parallel()

	checks := []*Check{
		NewCheck("skip-1", []string{"test"}),
		NewCheck("skip-2", []string{"test"}),
		NewCheck("skip-3", []string{"test"}),
	}

	skipAll(checks, "batch skip")

	for _, c := range checks {
		assert.Equal(t, CheckResultSkipped, c.Result.String())
	}
}

func TestOnFailure(t *testing.T) {
	saveAndResetDBState(t)

	group := NewChecksGroup("failure-test")

	current := NewCheck("current", []string{"test"})
	remaining := []*Check{
		NewCheck("remaining-1", []string{"test"}),
		NewCheck("remaining-2", []string{"test"}),
	}

	err := onFailure("panic", "something crashed", group, current, remaining)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failure-test")

	assert.Equal(t, CheckResultError, current.Result.String())
	assert.Equal(t, CheckResultSkipped, remaining[0].Result.String())
	assert.Equal(t, CheckResultSkipped, remaining[1].Result.String())
}

func TestShouldSkipCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func() *Check
		expectSkip  bool
		expectCount int
	}{
		{
			name: "no skip functions returns false",
			setup: func() *Check {
				return NewCheck("no-skip-fns", []string{"test"})
			},
			expectSkip: false,
		},
		{
			name: "skipModeAny with one skip returns true",
			setup: func() *Check {
				c := NewCheck("any-one-skip", []string{"test"})
				c.WithSkipCheckFn(
					func() (bool, string) { return true, "reason1" },
					func() (bool, string) { return false, "" },
				)
				return c
			},
			expectSkip:  true,
			expectCount: 1,
		},
		{
			name: "skipModeAll with only some returning true",
			setup: func() *Check {
				c := NewCheck("all-partial", []string{"test"})
				c.WithSkipModeAll()
				c.WithSkipCheckFn(
					func() (bool, string) { return true, "reason1" },
					func() (bool, string) { return false, "" },
				)
				return c
			},
			expectSkip: false,
		},
		{
			name: "skipModeAll with all returning true",
			setup: func() *Check {
				c := NewCheck("all-true", []string{"test"})
				c.WithSkipModeAll()
				c.WithSkipCheckFn(
					func() (bool, string) { return true, "reason1" },
					func() (bool, string) { return true, "reason2" },
				)
				return c
			},
			expectSkip:  true,
			expectCount: 2,
		},
		{
			name: "none return true",
			setup: func() *Check {
				c := NewCheck("none-true", []string{"test"})
				c.WithSkipCheckFn(
					func() (bool, string) { return false, "" },
					func() (bool, string) { return false, "" },
				)
				return c
			},
			expectSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := tt.setup()
			skip, reasons := shouldSkipCheck(check)
			assert.Equal(t, tt.expectSkip, skip)
			if tt.expectSkip {
				assert.Len(t, reasons, tt.expectCount)
			}
		})
	}
}

func TestRunChecksOrchestration(t *testing.T) {
	saveAndResetDBState(t)

	var callOrder []string

	err := InitLabelsExprEvaluator("test")
	require.NoError(t, err)

	group := NewChecksGroup("orchestration")
	group.WithBeforeAllFn(func(checks []*Check) error {
		callOrder = append(callOrder, "beforeAll")
		return nil
	})
	group.WithBeforeEachFn(func(check *Check) error {
		callOrder = append(callOrder, "beforeEach:"+check.ID)
		return nil
	})
	group.WithAfterEachFn(func(check *Check) error {
		callOrder = append(callOrder, "afterEach:"+check.ID)
		return nil
	})
	group.WithAfterAllFn(func(checks []*Check) error {
		callOrder = append(callOrder, "afterAll")
		return nil
	})

	check1 := NewCheck("check-1", []string{"test"})
	check1.WithCheckFn(func(c *Check) error {
		callOrder = append(callOrder, "check:"+c.ID)
		return nil
	})
	group.Add(check1)

	check2 := NewCheck("check-2", []string{"test"})
	check2.WithCheckFn(func(c *Check) error {
		callOrder = append(callOrder, "check:"+c.ID)
		return nil
	})
	group.Add(check2)

	stopChan := make(chan bool, 1)
	abortChan := make(chan string, 1)

	errs, failedChecks := group.RunChecks(stopChan, abortChan)

	assert.Empty(t, errs)
	assert.Equal(t, 0, failedChecks)
	assert.Equal(t, []string{
		"beforeAll",
		"beforeEach:check-1", "check:check-1", "afterEach:check-1",
		"beforeEach:check-2", "check:check-2", "afterEach:check-2",
		"afterAll",
	}, callOrder)
}

func TestRunChecksSkipsByLabel(t *testing.T) {
	saveAndResetDBState(t)

	err := InitLabelsExprEvaluator("common")
	require.NoError(t, err)

	group := NewChecksGroup("label-skip")

	matching := NewCheck("matching-check", []string{"common"})
	matching.WithCheckFn(func(c *Check) error { return nil })
	group.Add(matching)

	nonMatching := NewCheck("non-matching", []string{"extended"})
	nonMatching.WithCheckFn(func(c *Check) error { return nil })
	group.Add(nonMatching)

	stopChan := make(chan bool, 1)
	abortChan := make(chan string, 1)

	errs, _ := group.RunChecks(stopChan, abortChan)

	assert.Empty(t, errs)
	assert.Equal(t, CheckResultPassed, matching.Result.String())
	assert.Equal(t, CheckResultSkipped, nonMatching.Result.String())
}

func TestRunChecksCountsFailures(t *testing.T) {
	saveAndResetDBState(t)

	err := InitLabelsExprEvaluator("test")
	require.NoError(t, err)

	group := NewChecksGroup("fail-count")

	passing := NewCheck("pass-check", []string{"test"})
	passing.WithCheckFn(func(c *Check) error {
		return nil
	})
	group.Add(passing)

	failing := NewCheck("fail-check", []string{"test"})
	failing.WithCheckFn(func(c *Check) error {
		c.Result = CheckResultFailed
		return nil
	})
	group.Add(failing)

	stopChan := make(chan bool, 1)
	abortChan := make(chan string, 1)

	errs, failedChecks := group.RunChecks(stopChan, abortChan)
	assert.Empty(t, errs)
	assert.Equal(t, 1, failedChecks)
}

func TestRunChecksBeforeAllError(t *testing.T) {
	saveAndResetDBState(t)

	err := InitLabelsExprEvaluator("test")
	require.NoError(t, err)

	group := NewChecksGroup("before-all-err")
	group.WithBeforeAllFn(func(checks []*Check) error {
		return errors.New("beforeAll failed")
	})

	check1 := NewCheck("check-1", []string{"test"})
	check1.WithCheckFn(func(c *Check) error { return nil })
	group.Add(check1)

	stopChan := make(chan bool, 1)
	abortChan := make(chan string, 1)

	errs, _ := group.RunChecks(stopChan, abortChan)
	require.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Error(), "beforeAll")
}

func TestOnAbort(t *testing.T) {
	saveAndResetDBState(t)

	err := InitLabelsExprEvaluator("test")
	require.NoError(t, err)

	group := &ChecksGroup{
		name:                   "abort-test",
		currentRunningCheckIdx: 1,
		checks: []*Check{
			NewCheck("done-check", []string{"test"}),
			NewCheck("running-check", []string{"test"}),
			NewCheck("pending-check", []string{"test"}),
		},
	}

	err = group.OnAbort("test abort")
	assert.NoError(t, err)

	// Check at index 1 (running) should be aborted
	assert.Equal(t, CheckResultAborted, group.checks[1].Result.String())
	// Check at index 2 (pending) should be skipped
	assert.Equal(t, CheckResultSkipped, group.checks[2].Result.String())
}

func TestOnAbortNoRunningCheck(t *testing.T) {
	saveAndResetDBState(t)

	err := InitLabelsExprEvaluator("test")
	require.NoError(t, err)

	group := &ChecksGroup{
		name:                   "abort-none",
		currentRunningCheckIdx: checkIdxNone,
		checks: []*Check{
			NewCheck("check-1", []string{"test"}),
			NewCheck("check-2", []string{"test"}),
		},
	}

	err = group.OnAbort("full abort")
	assert.NoError(t, err)

	assert.Equal(t, CheckResultSkipped, group.checks[0].Result.String())
	assert.Equal(t, CheckResultSkipped, group.checks[1].Result.String())
}

func TestRecordChecksResultsNotFound(t *testing.T) {
	saveAndResetDBState(t)

	group := &ChecksGroup{
		name: "record-test",
		checks: []*Check{
			NewCheck("unknown-check-1", []string{"test"}),
			NewCheck("unknown-check-2", []string{"test"}),
		},
	}

	group.RecordChecksResults()

	// These check IDs won't be in TestIDToClaimID, so nothing should be recorded
	assert.Empty(t, resultsDB)
}

func TestRunChecksEmptyGroup(t *testing.T) {
	saveAndResetDBState(t)

	err := InitLabelsExprEvaluator("test")
	require.NoError(t, err)

	group := NewChecksGroup("empty-group")

	stopChan := make(chan bool, 1)
	abortChan := make(chan string, 1)

	errs, failedChecks := group.RunChecks(stopChan, abortChan)
	assert.Empty(t, errs)
	assert.Equal(t, 0, failedChecks)
}

func TestRunChecksWithSkipFn(t *testing.T) {
	saveAndResetDBState(t)

	err := InitLabelsExprEvaluator("test")
	require.NoError(t, err)

	group := NewChecksGroup("skip-fn-test")

	skippable := NewCheck("skippable", []string{"test"})
	skippable.WithCheckFn(func(c *Check) error {
		t.Fatal("skipped check should not run")
		return nil
	})
	skippable.WithSkipCheckFn(func() (bool, string) {
		return true, "skip reason"
	})
	group.Add(skippable)

	stopChan := make(chan bool, 1)
	abortChan := make(chan string, 1)

	errs, _ := group.RunChecks(stopChan, abortChan)
	assert.Empty(t, errs)
	assert.Equal(t, CheckResultSkipped, skippable.Result.String())
}
