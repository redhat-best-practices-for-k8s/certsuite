package checksdb

import (
	"errors"
	"testing"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	cr := CheckResult("passed")
	assert.Equal(t, "passed", cr.String())
}

func TestNewCheck(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	assert.NotNil(t, check)

	assert.Equal(t, "myID", check.ID)
	assert.Equal(t, []string{"label1", "label2"}, check.Labels)
}

func TestSetAbortChan(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})
	abortChan := make(chan string)

	check.SetAbortChan(abortChan)

	assert.Equal(t, abortChan, check.abortChan)
}

func TestGetLogs(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	logs := check.GetLogs()

	assert.NotNil(t, logs)
}

func TestGetLogger(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	logger := check.GetLogger()

	assert.NotNil(t, logger)
}

func TestWithCheckFn(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	check.WithCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.CheckFn)

	check.Error = errors.New("this is an error")
	check.WithCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.CheckFn)
	assert.Equal(t, "this is an error", check.Error.Error())
}

func TestWithBeforeCheckFn(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	check.WithBeforeCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.BeforeCheckFn)

	check.Error = errors.New("this is an error")
	check.WithBeforeCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.BeforeCheckFn)
	assert.Equal(t, "this is an error", check.Error.Error())
}

func TestWithAfterCheckFn(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	check.WithAfterCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.AfterCheckFn)

	check.Error = errors.New("this is an error")
	check.WithAfterCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.AfterCheckFn)
	assert.Equal(t, "this is an error", check.Error.Error())
}

func TestWithSkipCheckFn(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	check.WithSkipCheckFn(func() (skip bool, reason string) {
		return false, ""
	})

	assert.Len(t, check.SkipCheckFns, 1)

	check.WithSkipCheckFn(func() (skip bool, reason string) {
		return false, ""
	})

	assert.Len(t, check.SkipCheckFns, 2)
}

func TestWithSkipModeAny(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	// Test the default value, which is SkipModeAny
	assert.Equal(t, SkipModeAny, check.SkipMode)

	check.WithSkipModeAny()

	assert.Equal(t, SkipModeAny, check.SkipMode)
}

func TestWithSkipModeAll(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	// Test the default value, which is SkipModeAny
	assert.Equal(t, SkipModeAny, check.SkipMode)

	check.WithSkipModeAll()

	assert.Equal(t, SkipModeAll, check.SkipMode)
}

func TestWithTimeout(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	check.WithTimeout(10)

	assert.Equal(t, time.Duration(10), check.Timeout)
}

func TestSetResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialResult  CheckResult
		compliant      []*testhelper.ReportObject
		nonCompliant   []*testhelper.ReportObject
		expectedResult CheckResult
	}{
		{
			name:          "passed with only compliant objects",
			initialResult: CheckResultPassed,
			compliant: []*testhelper.ReportObject{
				testhelper.NewPodReportObject("ns", "pod1", "ok", true),
			},
			nonCompliant:   nil,
			expectedResult: CheckResultPassed,
		},
		{
			name:          "failed with non-compliant objects",
			initialResult: CheckResultPassed,
			compliant: []*testhelper.ReportObject{
				testhelper.NewPodReportObject("ns", "pod1", "ok", true),
			},
			nonCompliant: []*testhelper.ReportObject{
				testhelper.NewPodReportObject("ns", "pod2", "bad", false),
			},
			expectedResult: CheckResultFailed,
		},
		{
			name:           "skipped when both lists empty",
			initialResult:  CheckResultPassed,
			compliant:      nil,
			nonCompliant:   nil,
			expectedResult: CheckResultSkipped,
		},
		{
			name:          "no change when already aborted",
			initialResult: CheckResultAborted,
			compliant: []*testhelper.ReportObject{
				testhelper.NewPodReportObject("ns", "pod1", "ok", true),
			},
			nonCompliant:   nil,
			expectedResult: CheckResultAborted,
		},
		{
			name:          "no change from error to failed",
			initialResult: CheckResultError,
			nonCompliant: []*testhelper.ReportObject{
				testhelper.NewPodReportObject("ns", "pod1", "bad", false),
			},
			expectedResult: CheckResultError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := NewCheck("test-set-result", []string{"test"})
			check.Result = tt.initialResult
			check.SetResult(tt.compliant, tt.nonCompliant)
			assert.Equal(t, tt.expectedResult, check.Result)
		})
	}
}

func TestSetResultSkipped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialResult  CheckResult
		expectedResult CheckResult
	}{
		{
			name:           "sets skipped from passed",
			initialResult:  CheckResultPassed,
			expectedResult: CheckResultSkipped,
		},
		{
			name:           "no change when already aborted",
			initialResult:  CheckResultAborted,
			expectedResult: CheckResultAborted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := NewCheck("test-skip", []string{"test"})
			check.Result = tt.initialResult
			check.SetResultSkipped("test reason")
			assert.Equal(t, tt.expectedResult, check.Result)
		})
	}
}

func TestSetResultError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialResult  CheckResult
		expectedResult CheckResult
	}{
		{
			name:           "sets error from passed",
			initialResult:  CheckResultPassed,
			expectedResult: CheckResultError,
		},
		{
			name:           "no change when already aborted",
			initialResult:  CheckResultAborted,
			expectedResult: CheckResultAborted,
		},
		{
			name:           "no change when already error",
			initialResult:  CheckResultError,
			expectedResult: CheckResultError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := NewCheck("test-error", []string{"test"})
			check.Result = tt.initialResult
			check.SetResultError("test error reason")
			assert.Equal(t, tt.expectedResult, check.Result)
		})
	}
}

func TestSetResultAborted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		initialResult CheckResult
	}{
		{
			name:          "sets aborted from passed",
			initialResult: CheckResultPassed,
		},
		{
			name:          "overrides error",
			initialResult: CheckResultError,
		},
		{
			name:          "overrides failed",
			initialResult: CheckResultFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := NewCheck("test-abort", []string{"test"})
			check.Result = tt.initialResult
			check.SetResultAborted("test abort reason")
			assert.Equal(t, CheckResultAborted, check.Result.String())
		})
	}
}

func TestLogMethods(t *testing.T) {
	t.Parallel()

	check := NewCheck("test-log", []string{"test"})

	check.LogInfo("info message %d", 2)
	check.LogWarn("warn message %d", 3)
	check.LogError("error message %d", 4)

	logs := check.GetLogs()
	assert.Contains(t, logs, "info message 2")
	assert.Contains(t, logs, "warn message 3")
	assert.Contains(t, logs, "error message 4")
}

func TestRun(t *testing.T) {
	// Not parallel: Check.Run() calls cli.PrintCheckRunning/PrintCheckPassed
	// which use package-level channels unsafe for concurrent access.

	tests := []struct {
		name        string
		setup       func() *Check
		expectErr   bool
		errContains string
	}{
		{
			name: "successful run with all hooks",
			setup: func() *Check {
				check := NewCheck("test-run", []string{"test"})
				check.WithBeforeCheckFn(func(c *Check) error { return nil })
				check.WithCheckFn(func(c *Check) error { return nil })
				check.WithAfterCheckFn(func(c *Check) error { return nil })
				return check
			},
			expectErr: false,
		},
		{
			name: "successful run without optional hooks",
			setup: func() *Check {
				check := NewCheck("test-run-minimal", []string{"test"})
				check.WithCheckFn(func(c *Check) error { return nil })
				return check
			},
			expectErr: false,
		},
		{
			name: "error when check has pre-existing error",
			setup: func() *Check {
				check := NewCheck("test-run-preerr", []string{"test"})
				check.Error = errors.New("pre-existing error")
				return check
			},
			expectErr:   true,
			errContains: "previously existing error",
		},
		{
			name: "error from beforeCheckFn",
			setup: func() *Check {
				check := NewCheck("test-run-before", []string{"test"})
				check.WithBeforeCheckFn(func(c *Check) error { return errors.New("before failed") })
				check.WithCheckFn(func(c *Check) error { return nil })
				return check
			},
			expectErr:   true,
			errContains: "before check function",
		},
		{
			name: "error from checkFn",
			setup: func() *Check {
				check := NewCheck("test-run-checkfn", []string{"test"})
				check.WithCheckFn(func(c *Check) error { return errors.New("check failed") })
				return check
			},
			expectErr:   true,
			errContains: "check function",
		},
		{
			name: "error from afterCheckFn",
			setup: func() *Check {
				check := NewCheck("test-run-after", []string{"test"})
				check.WithCheckFn(func(c *Check) error { return nil })
				check.WithAfterCheckFn(func(c *Check) error { return errors.New("after failed") })
				return check
			},
			expectErr:   true,
			errContains: "after check function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := tt.setup()
			err := check.Run()
			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunNilCheck(t *testing.T) {
	t.Parallel()

	var check *Check
	err := check.Run()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil pointer")
}

func TestRunSetsTimestamps(t *testing.T) {
	// Not parallel: same cli channel race as TestRun.

	check := NewCheck("test-timestamps", []string{"test"})
	check.WithCheckFn(func(c *Check) error { return nil })

	before := time.Now()
	err := check.Run()
	after := time.Now()

	require.NoError(t, err)
	assert.True(t, !check.StartTime.Before(before))
	assert.True(t, !check.EndTime.After(after))
	assert.True(t, !check.EndTime.Before(check.StartTime))
}

func TestAbort(t *testing.T) {
	t.Parallel()

	check := NewCheck("test-abort-fn", []string{"test"})
	abortChan := make(chan string, 1)
	check.SetAbortChan(abortChan)

	assert.Panics(t, func() {
		check.Abort("abort reason")
	})

	msg := <-abortChan
	assert.Contains(t, msg, "abort reason")
}
