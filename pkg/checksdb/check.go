package checksdb

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
)

const (
	CheckResultPassed  = "passed"
	CheckResultSkipped = "skipped"
	CheckResultFailed  = "failed"
	CheckResultError   = "error"
	CheckResultAborted = "aborted"
)

type CheckResult string

func (cr CheckResult) String() string {
	return string(cr)
}

type Check struct {
	mutex  sync.Mutex
	ID     string
	Labels []string

	BeforeCheckFn, AfterCheckFn func(check *Check) error
	CheckFn                     func(check *Check) error

	SkipCheckFn func() (skip bool, reason string)

	Result         CheckResult
	CapturedOutput string
	FailureReason  string

	StartTime, EndTime time.Time
	Timeout            time.Duration
	Error              error
}

func NewCheck(id string, labels []string) *Check {
	return &Check{
		ID:     id,
		Labels: labels,
		Result: CheckResultPassed,
	}
}

func (check *Check) WithCheckFn(checkFn func(check *Check) error) *Check {
	if check.Error != nil {
		return check
	}

	check.CheckFn = checkFn
	return check
}

func (check *Check) WithBeforeCheckFn(beforeCheckFn func(check *Check) error) *Check {
	if check.Error != nil {
		return check
	}

	check.BeforeCheckFn = beforeCheckFn
	return check
}

func (check *Check) WithAfterCheckFn(afterCheckFn func(check *Check) error) *Check {
	if check.Error != nil {
		return check
	}

	check.AfterCheckFn = afterCheckFn
	return check
}

func (check *Check) WithSkipCheckFn(skipCheckFn func() (skip bool, reason string)) *Check {
	if check.Error != nil {
		return check
	}

	check.SkipCheckFn = skipCheckFn

	return check
}

func (check *Check) WithTimeout(duration time.Duration) *Check {
	if check.Error != nil {
		return check
	}

	check.Timeout = duration

	return check
}

func (check *Check) SetResult(compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	if check.Result == CheckResultAborted {
		return
	}

	resultObjectsStr, err := testhelper.ResultObjectsToString(compliantObjects, nonCompliantObjects)
	if err != nil {
		logrus.Errorf("Failed to get result objects string for check %s: %v", check.ID, err)
	}

	check.CapturedOutput = resultObjectsStr

	// If an error/panic happened before, do not change the result.
	if check.Result == CheckResultError {
		return
	}

	if len(nonCompliantObjects) > 0 {
		check.Result = CheckResultFailed
		check.FailureReason = resultObjectsStr
	} else if len(compliantObjects) == 0 {
		// Mark this check as skipped.
		logrus.Warnf("Check %s marked as skipped as both compliant and non-compliant objects lists are empty.", check.ID)
		check.FailureReason = "Compliant and non-compliant objects lists are empty."
		check.Result = CheckResultSkipped
	}
}

func (check *Check) SetResultFailed(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	if check.Result == CheckResultAborted {
		return
	}

	check.Result = CheckResultFailed
	check.FailureReason = reason
}

func (check *Check) SetResultSkipped(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	if check.Result == CheckResultAborted {
		return
	}

	check.Result = CheckResultSkipped
	check.FailureReason = reason
}

func (check *Check) SetResultError(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	if check.Result == CheckResultAborted {
		return
	}

	if check.Result == CheckResultError {
		logrus.Warnf("Check %s result was already marked as error.", check.ID)
		return
	}
	check.Result = CheckResultError
	check.FailureReason = reason
}

func (check *Check) SetResultAborted(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	check.Result = CheckResultAborted
	check.FailureReason = reason
}

func (check *Check) Run() error {
	if check == nil {
		return fmt.Errorf("check is a nil pointer")
	}

	if check.Error != nil {
		return fmt.Errorf("unable to run due to a previously existing error: %v", check.Error)
	}

	check.StartTime = time.Now()
	defer func() {
		check.EndTime = time.Now()
	}()

	logrus.Infof("RUNNING CHECK: %s (labels: %v)", check.ID, check.Labels)
	if check.BeforeCheckFn != nil {
		if err := check.BeforeCheckFn(check); err != nil {
			return fmt.Errorf("check %s failed in before check function: %v", check.ID, err)
		}
	}

	if err := check.CheckFn(check); err != nil {
		return fmt.Errorf("check %s failed in check function: %v", check.ID, err)
	}

	if check.AfterCheckFn != nil {
		if err := check.AfterCheckFn(check); err != nil {
			return fmt.Errorf("check %s failed in after check function: %v", check.ID, err)
		}
	}

	return nil
}
