package checksdb

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/cli"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
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

// CheckResult.String
// Returns a human‑readable string representation of the check result.
//
// It converts the CheckResult value to its corresponding string form,
// such as "PASSED", "FAILED", "SKIPPED", or "ERROR". The returned string
// is used for reporting and logging of test outcomes.
func (cr CheckResult) String() string {
	return string(cr)
}

// Check represents a single test check that can be executed, logged and recorded.
//
// It holds the check's metadata such as ID, labels, timeout, and execution
// times. The functional fields BeforeCheckFn, CheckFn, AfterCheckFn allow
// custom setup, execution and teardown logic. SkipCheckFns determine if the
// check should run based on dynamic conditions. Results are stored in Result,
// and any errors or abort signals are handled via Abort and related methods.
// Logging is performed through an embedded logger, and captured output is
// available via CapturedOutput. The struct also supports concurrent access
// protection with a mutex.
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

// NewCheck constructs a new check instance.
//
// It receives the name of the check and an optional list of tag strings,
// initializes a Check object with these values, attaches a multi‑logger
// obtained from GetMultiLogger, and applies any additional configuration
// via With before returning the configured *Check.
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

// Abort signals that the current check should terminate immediately and cause a panic with a provided message.
//
// The returned function, when called, locks the database mutex,
// panics using the supplied message to indicate an abort condition,
// and then unlocks the mutex. It is intended to be invoked from within
// a check implementation to stop execution early while ensuring
// proper synchronization and error reporting.
func (check *Check) Abort(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	abortMsg := check.ID + " issued non-graceful abort: " + reason

	check.abortChan <- abortMsg
	panic(AbortPanicMsg(abortMsg))
}

// SetAbortChan assigns a channel used to signal abortion of the check and
// returns a function that can be called to close that channel.
//
// The returned function accepts a channel of strings, which should carry an
// abort reason or message. When invoked, it closes the provided channel,
// allowing any goroutine listening on the original abort channel to detect
// the termination signal and halt processing accordingly. This helper simplifies
// cleanup of abortion signals when a check is cancelled or fails early.
func (check *Check) SetAbortChan(abortChan chan string) {
	check.abortChan = abortChan
}

// LogDebug logs a debug message for the check and returns nil.
//
// It accepts a format string and optional arguments, similar to fmt.Printf,
// and writes the formatted output using the Check's internal logging
// mechanism. The function is intended for verbose debugging of individual
// checks and does not affect test results or control flow. The return value
// is always nil, allowing callers to ignore it if desired.
func (check *Check) LogDebug(msg string, args ...any) {
	log.Logf(check.logger, log.LevelDebug, msg, args...)
}

// LogInfo logs informational messages related to a check.
//
// It accepts a format string and optional arguments, forwards them to the
// underlying Logf function, and returns a no-op function that can be used
// as a deferable cleanup if desired. The returned function performs no
// action when called. This method is intended for use within check
// implementations to emit non-critical information about test execution.
func (check *Check) LogInfo(msg string, args ...any) {
	log.Logf(check.logger, log.LevelInfo, msg, args...)
}

// LogWarn logs a warning message associated with the check.
//
// It accepts a format string and optional arguments, which are passed to
// Logf to produce the final log entry. The returned function is intended to be
// deferred so that any cleanup actions or additional logging can occur after
// the current function completes. This method does not return a value; it only
// schedules the warning for later emission.
func (check *Check) LogWarn(msg string, args ...any) {
	log.Logf(check.logger, log.LevelWarn, msg, args...)
}

// LogError logs an error message for the check and returns a function that records
// the failure in the database.
//
// It accepts a format string and optional arguments, similar to fmt.Printf.
// The returned function should be called when the check has finished executing,
// and it will store the error details in the results database.
func (check *Check) LogError(msg string, args ...any) {
	log.Logf(check.logger, log.LevelError, msg, args...)
}

// LogFatal logs a fatal error message and exits the program.
//
// It formats the provided message using fmt.Sprintf with any additional arguments,
// writes the formatted output to standard error, and then terminates execution
// by calling os.Exit(1). The function does not return normally; it is intended
// for unrecoverable errors within a Check.
func (check *Check) LogFatal(msg string, args ...any) {
	log.Logf(check.logger, log.LevelFatal, msg, args...)
	fmt.Fprintf(os.Stderr, "\nFATAL: "+msg+"\n", args...)
	os.Exit(1)
}

// GetLogs retrieves the accumulated log output for a Check.
//
// The returned string contains all messages that have been recorded
// during the execution of this check. It is produced by calling the
// internal String method on the Check instance, which formats the
// stored log entries into a single string.
func (check *Check) GetLogs() string {
	return check.logArchive.String()
}

// GetLogger returns the logger associated with this check.
//
// The returned *log.Logger is used to emit log messages during the execution of
// the check. It allows callers to inspect or configure logging behaviour specific
// to a Check instance. If no custom logger has been set, the default logger for
// the package is returned.
func (check *Check) GetLogger() *log.Logger {
	return check.logger
}

// WithCheckFn registers a custom check function to be executed when the Check is run.
//
// It takes a function that receives a pointer to the current Check and returns an error.
// The provided function can perform arbitrary validation logic and set the Check's result or reason.
// The method returns the same Check instance, allowing method chaining.
func (check *Check) WithCheckFn(checkFn func(check *Check) error) *Check {
	if check.Error != nil {
		return check
	}

	check.CheckFn = checkFn
	return check
}

// WithBeforeCheckFn registers a function to be executed before the check runs.
//
// It takes a callback that receives the current Check instance and may return an error.
// The callback is stored in the Check object and will be invoked automatically prior
// to running the main check logic, allowing pre‑execution setup or validation.  
// The method returns the modified Check for chaining.
func (check *Check) WithBeforeCheckFn(beforeCheckFn func(check *Check) error) *Check {
	if check.Error != nil {
		return check
	}

	check.BeforeCheckFn = beforeCheckFn
	return check
}

// WithAfterCheckFn registers a function to be executed after the check finishes.
//
// It takes a callback that receives the Check instance and may return an error.
// The callback is stored on the Check and will run after the main check logic,
// allowing post‑processing or cleanup steps. The method returns the modified
// Check so calls can be chained.
func (check *Check) WithAfterCheckFn(afterCheckFn func(check *Check) error) *Check {
	if check.Error != nil {
		return check
	}

	check.AfterCheckFn = afterCheckFn
	return check
}

// WithSkipCheckFn registers one or more skip check functions for a Check.
//
// It appends the provided functions to the Check's internal list of
// skip-check callbacks. Each callback should return a boolean indicating
// whether the check should be skipped and an optional reason string.
// The returned *Check allows chaining additional configuration calls.
func (check *Check) WithSkipCheckFn(skipCheckFn ...func() (skip bool, reason string)) *Check {
	if check.Error != nil {
		return check
	}

	check.SkipCheckFns = append(check.SkipCheckFns, skipCheckFn...)

	return check
}

// WithSkipModeAny sets the skip mode to any, which is the default.
//
// It returns a pointer to the Check with its SkipMode set to SkipModeAny,
// allowing callers to explicitly specify that a check may be skipped if
// any of its conditions are not met. This modifier exists for completeness
// but does not alter behavior when already using the default mode.
func (check *Check) WithSkipModeAny() *Check {
	if check.Error != nil {
		return check
	}

	check.SkipMode = SkipModeAny

	return check
}

// WithSkipModeAll sets the check to skip all sub‑checks if any fail.
//
// It modifies the Check instance so that when executed, it will run each
// associated sub‑check and immediately stop further execution if one of them
// fails or errors. The function returns a pointer to the modified Check,
// allowing method chaining.
func (check *Check) WithSkipModeAll() *Check {
	if check.Error != nil {
		return check
	}

	check.SkipMode = SkipModeAll

	return check
}

// WithTimeout sets the maximum duration allowed for the check to run.
//
// It returns a new Check instance configured with the supplied time.Duration.
// The returned check will enforce this timeout when executed, ensuring that
// long‑running or hanging checks do not block the test suite indefinitely.
func (check *Check) WithTimeout(duration time.Duration) *Check {
	if check.Error != nil {
		return check
	}

	check.Timeout = duration

	return check
}

// SetResult records the result of a check and updates internal state.
//
// It locks the checks database, converts the passed report objects to
// string representations, logs any errors or warnings from those reports,
// and stores the final status strings in the Check instance. The function
// takes two slices of ReportObject: one for successful results and one
// for failed results. After processing it releases the lock.
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

// SetResultSkipped marks a check as skipped and records the reason.
//
// It locks the internal database mutex, updates the check’s result to SKIPPED,
// stores the supplied skip reason string, and then unlocks the mutex.
// The function returns immediately after setting the state; it does not
// perform any additional logic or error handling.
func (check *Check) SetResultSkipped(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	if check.Result == CheckResultAborted {
		return
	}

	check.Result = CheckResultSkipped
	check.skipReason = reason
}

// SetResultError records an error result for a check and logs a warning.
//
// It accepts an error message string, locks the checks database,
// sets the check's Result to CheckResultError with the provided message,
// updates the Status to FAILED, increments the error count, and then unlocks.
// Finally it logs the warning using LogWarn.
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

// SetResultAborted marks a check as aborted and returns an empty result string.
//
// It takes the reason string, locks the checks database,
// sets the check's Result field to CheckResultAborted,
// stores the reason in the Message field, and unlocks the database.
// The function returns an empty string; callers can ignore it.
func (check *Check) SetResultAborted(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	check.Result = CheckResultAborted
	check.skipReason = reason
}

// Run executes the check, invoking its lifecycle callbacks and recording the outcome.
//
// It performs the following steps: logs that the check is running,
// records start time, calls BeforeCheckFn if present, runs the main CheckFn,
// handles any error returned by it, then calls AfterCheckFn.
// The result (passed, failed, aborted, or skipped) is logged and stored.
// If an error occurs during any phase, Run returns that error.
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

// printCheckResult displays the result of a check.
//
// It inspects the Check's Result field and calls one of the
// PrintCheck* functions to output an appropriate message.
// No value is returned; the function only performs side effects.
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
