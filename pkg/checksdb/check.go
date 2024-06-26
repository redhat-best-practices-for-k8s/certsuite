package checksdb

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/test-network-function/cnf-certification-test/internal/cli"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
)

const (
	CheckResultPassed  = "passed"
	CheckResultSkipped = "skipped"
	CheckResultFailed  = "failed"
	CheckResultError   = "error"
	CheckResultAborted = "aborted"
)

type skipMode int

const (
	SkipModeAny skipMode = iota
	SkipModeAll
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

	SkipCheckFns []func() (skip bool, reason string)
	SkipMode     skipMode

	Result         CheckResult
	CapturedOutput string
	details        string
	skipReason     string

	logger     *log.Logger
	logArchive *strings.Builder

	StartTime, EndTime time.Time
	Timeout            time.Duration
	Error              error
	abortChan          chan string
}

func NewCheck(id string, labels []string) *Check {
	check := &Check{
		ID:         id,
		Labels:     labels,
		Result:     CheckResultPassed,
		logArchive: &strings.Builder{},
	}

	check.logger = log.GetMultiLogger(check.logArchive, cli.CliCheckLogSniffer).With("check", check.ID)

	return check
}

func (check *Check) Abort(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	abortMsg := check.ID + " issued non-graceful abort: " + reason

	check.abortChan <- abortMsg
	panic(AbortPanicMsg(abortMsg))
}

func (check *Check) SetAbortChan(abortChan chan string) {
	check.abortChan = abortChan
}

func (check *Check) LogDebug(msg string, args ...any) {
	log.Logf(check.logger, log.LevelDebug, msg, args...)
}

func (check *Check) LogInfo(msg string, args ...any) {
	log.Logf(check.logger, log.LevelInfo, msg, args...)
}

func (check *Check) LogWarn(msg string, args ...any) {
	log.Logf(check.logger, log.LevelWarn, msg, args...)
}

func (check *Check) LogError(msg string, args ...any) {
	log.Logf(check.logger, log.LevelError, msg, args...)
}

func (check *Check) LogFatal(msg string, args ...any) {
	log.Logf(check.logger, log.LevelFatal, msg, args...)
	fmt.Fprintf(os.Stderr, "\nFATAL: "+msg+"\n", args...)
	os.Exit(1)
}

func (check *Check) GetLogs() string {
	return check.logArchive.String()
}

func (check *Check) GetLogger() *log.Logger {
	return check.logger
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

func (check *Check) WithSkipCheckFn(skipCheckFn ...func() (skip bool, reason string)) *Check {
	if check.Error != nil {
		return check
	}

	check.SkipCheckFns = append(check.SkipCheckFns, skipCheckFn...)

	return check
}

// This modifier is provided for the sake of completeness, but it's not necessary to use it,
// as the SkipModeAny is the default skip mode.
func (check *Check) WithSkipModeAny() *Check {
	if check.Error != nil {
		return check
	}

	check.SkipMode = SkipModeAny

	return check
}

func (check *Check) WithSkipModeAll() *Check {
	if check.Error != nil {
		return check
	}

	check.SkipMode = SkipModeAll

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
		check.LogError("Failed to get result objects string for check %s: %v", check.ID, err)
	}

	check.details = resultObjectsStr

	// If an error/panic happened before, do not change the result.
	if check.Result == CheckResultError {
		return
	}

	if len(nonCompliantObjects) > 0 {
		check.Result = CheckResultFailed
		check.skipReason = ""
	} else if len(compliantObjects) == 0 {
		// Mark this check as skipped.
		check.LogWarn("Check %s marked as skipped as both compliant and non-compliant objects lists are empty.", check.ID)
		check.skipReason = "compliant and non-compliant objects lists are empty"
		check.Result = CheckResultSkipped
	}
}

func (check *Check) SetResultSkipped(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	if check.Result == CheckResultAborted {
		return
	}

	check.Result = CheckResultSkipped
	check.skipReason = reason
}

func (check *Check) SetResultError(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	if check.Result == CheckResultAborted {
		return
	}

	if check.Result == CheckResultError {
		check.LogWarn("Check %s result was already marked as error.", check.ID)
		return
	}
	check.Result = CheckResultError
	check.skipReason = reason
}

func (check *Check) SetResultAborted(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	check.Result = CheckResultAborted
	check.skipReason = reason
}

func (check *Check) Run() error {
	if check == nil {
		return fmt.Errorf("check is a nil pointer")
	}

	if check.Error != nil {
		return fmt.Errorf("unable to run due to a previously existing error: %v", check.Error)
	}

	cli.PrintCheckRunning(check.ID)

	check.StartTime = time.Now()
	defer func() {
		check.EndTime = time.Now()
	}()

	check.LogInfo("Running check (labels: %v)", check.Labels)
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

	printCheckResult(check)

	return nil
}

func printCheckResult(check *Check) {
	switch check.Result {
	case CheckResultPassed:
		cli.PrintCheckPassed(check.ID)
	case CheckResultFailed:
		cli.PrintCheckFailed(check.ID)
	case CheckResultSkipped:
		cli.PrintCheckSkipped(check.ID, check.skipReason)
	case CheckResultAborted:
		cli.PrintCheckAborted(check.ID, check.skipReason)
	case CheckResultError:
		cli.PrintCheckErrored(check.ID)
	}
}
