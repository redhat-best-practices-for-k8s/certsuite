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

// CheckResult.String Converts the result code to a readable string
//
// This method casts the underlying CheckResult type, which is an alias of
// string, into a standard Go string. It returns the textual representation of
// the check outcome, such as "Passed", "Failed" or "Skipped". The conversion
// allows callers to use the result in logs and comparisons without needing to
// know the internal type.
func (cr CheckResult) String() string {
	return string(cr)
}

// Check Represents an individual compliance check
//
// This type holds configuration, state, and results for a single test. It
// tracks identifiers, labels, timing, timeouts, and any error that occurs
// during execution. The struct also contains optional functions to run before,
// after, or as the main check logic, along with mechanisms for skipping,
// aborting, and logging output.
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

// NewCheck Creates a new check instance
//
// This function constructs a Check object with the provided identifier and
// label set. It assigns an initial passed result status, creates a string
// builder for log storage, and attaches a multi‑logger that records events
// specific to this check. The fully initialized Check is then returned as a
// pointer.
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

// Check.Abort Aborts a check immediately with an error message
//
// The method locks the check’s mutex, constructs a descriptive abort message
// using the check ID and the supplied reason, sends this message on the abort
// channel, then panics to terminate execution. It is used to halt a check that
// encounters a non‑graceful failure condition.
func (check *Check) Abort(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	abortMsg := check.ID + " issued non-graceful abort: " + reason

	check.abortChan <- abortMsg
	panic(AbortPanicMsg(abortMsg))
}

// Check.SetAbortChan Assigns a channel to signal check abortion
//
// This method records the supplied channel into the check instance so that the
// check can listen for abort signals during execution. It performs a simple
// field assignment and does not return any value. The stored channel is later
// used by other parts of the framework to terminate the check prematurely when
// needed.
func (check *Check) SetAbortChan(abortChan chan string) {
	check.abortChan = abortChan
}

// Check.LogDebug logs a debug message with optional formatting
//
// This method sends a formatted string to the check's logger at the debug
// level, allowing additional arguments for interpolation. It forwards the call
// to an internal logging helper that determines if the debug level is enabled
// before emitting the record. No value is returned.
func (check *Check) LogDebug(msg string, args ...any) {
	log.Logf(check.logger, log.LevelDebug, msg, args...)
}

// Check.LogInfo Logs an informational message for a check
//
// This method forwards the supplied format string and arguments to a logging
// helper, tagging the output with the Info level. It uses the check's internal
// logger if available or falls back to a default logger. The function does not
// return any value; it simply emits the formatted log entry.
func (check *Check) LogInfo(msg string, args ...any) {
	log.Logf(check.logger, log.LevelInfo, msg, args...)
}

// Check.LogWarn logs a warning message for the check
//
// The method formats a message with optional arguments and forwards it to the
// internal logger at the warning level. It does not alter any state of the
// Check instance, only records diagnostic information that can be inspected
// later.
func (check *Check) LogWarn(msg string, args ...any) {
	log.Logf(check.logger, log.LevelWarn, msg, args...)
}

// Check.LogError logs an error message for the check
//
// This method sends a formatted string and optional arguments to the logging
// system at the error level, associating the log with the specific check
// instance. It uses the check's logger field or falls back to a default if nil.
// The function does not return any value.
func (check *Check) LogError(msg string, args ...any) {
	log.Logf(check.logger, log.LevelError, msg, args...)
}

// Check.LogFatal Logs a fatal message and terminates the program
//
// The method records a fatal log entry using the provided logger, prints the
// message to standard error prefixed with "FATAL:", and then exits the process
// with status code 1. It accepts a format string and optional arguments, which
// are passed to both the logger and the formatted output.
func (check *Check) LogFatal(msg string, args ...any) {
	log.Logf(check.logger, log.LevelFatal, msg, args...)
	fmt.Fprintf(os.Stderr, "\nFATAL: "+msg+"\n", args...)
	os.Exit(1)
}

// Check.GetLogs Retrieves stored log output
//
// This method returns the accumulated log data for a check as a single string.
// The logs are gathered during the check's execution and stored in an internal
// buffer, which this function simply exposes to callers such as reporting or
// result recording functions.
func (check *Check) GetLogs() string {
	return check.logArchive.String()
}

// Check.GetLogger Provides access to the check's logger
//
// The method returns the logger stored in the Check instance, allowing callers
// to log messages related to that specific check. It does not modify the state
// and simply exposes the internal logger pointer.
func (check *Check) GetLogger() *log.Logger {
	return check.logger
}

// Check.WithCheckFn Assigns a new check function only when no previous error exists
//
// This method first checks whether the Check instance already contains an
// error; if so, it returns the instance unchanged. Otherwise, it assigns the
// provided function to the CheckFn field and then returns the modified instance
// for chaining.
func (check *Check) WithCheckFn(checkFn func(check *Check) error) *Check {
	if check.Error != nil {
		return check
	}

	check.CheckFn = checkFn
	return check
}

// Check.WithBeforeCheckFn Assigns a custom function to run before the main check
//
// The method accepts a callback that receives the current Check instance and
// may return an error. If the Check already contains an error, it skips
// assignment and returns the Check unchanged; otherwise, it stores the callback
// in BeforeCheckFn and returns the same Check pointer for chaining.
func (check *Check) WithBeforeCheckFn(beforeCheckFn func(check *Check) error) *Check {
	if check.Error != nil {
		return check
	}

	check.BeforeCheckFn = beforeCheckFn
	return check
}

// Check.WithAfterCheckFn Sets a callback to run after the check completes
//
// The method attaches a function that will be invoked once the check finishes,
// provided no error has already occurred. It stores the supplied function in
// the AfterCheckFn field of the Check instance and returns the same instance
// for chaining.
func (check *Check) WithAfterCheckFn(afterCheckFn func(check *Check) error) *Check {
	if check.Error != nil {
		return check
	}

	check.AfterCheckFn = afterCheckFn
	return check
}

// Check.WithSkipCheckFn Adds functions that decide whether a test should be skipped
//
// When called, this method appends one or more supplied functions to the
// receiver's list of skip-check callbacks, but only if no previous error has
// been recorded on the Check instance. Each added function returns a boolean
// indicating whether skipping is required and an optional reason string. The
// updated Check pointer is then returned for chaining.
func (check *Check) WithSkipCheckFn(skipCheckFn ...func() (skip bool, reason string)) *Check {
	if check.Error != nil {
		return check
	}

	check.SkipCheckFns = append(check.SkipCheckFns, skipCheckFn...)

	return check
}

// Check.WithSkipModeAny sets the check to always skip when appropriate
//
// This method changes the internal skip mode of a check to allow it to be
// skipped under any circumstance that matches the default logic. If an error is
// already present on the check, the call becomes a no‑op and simply returns
// the existing instance. Otherwise it assigns SkipModeAny to the check and
// returns the updated object for chaining.
func (check *Check) WithSkipModeAny() *Check {
	if check.Error != nil {
		return check
	}

	check.SkipMode = SkipModeAny

	return check
}

// Check.WithSkipModeAll enables all-skip mode
//
// This method changes a check's configuration so that it will skip any
// remaining steps or validations, effectively marking the check as fully
// skipped. It first verifies that no error has already been recorded on the
// check; if an error exists, it returns immediately without modifying the
// state. When successful, it assigns the SkipModeAll constant to the check and
// returns the modified check for further chaining.
func (check *Check) WithSkipModeAll() *Check {
	if check.Error != nil {
		return check
	}

	check.SkipMode = SkipModeAll

	return check
}

// Check.WithTimeout assigns a timeout value to the check
//
// If the check has not already encountered an error, this method updates its
// Timeout field with the supplied duration and returns the modified check for
// chaining. If an error is present, it simply returns the check unchanged so
// that subsequent operations are skipped.
func (check *Check) WithTimeout(duration time.Duration) *Check {
	if check.Error != nil {
		return check
	}

	check.Timeout = duration

	return check
}

// Check.SetResult stores compliance results for a check
//
// This method records the list of compliant and non‑compliant objects for a
// check, converting them into a JSON string that is kept in the details field.
// It locks the check’s mutex to ensure thread safety, skips any changes if
// the check has already been aborted or errored, and updates the result status
// based on whether there are failures or no objects at all. Errors during
// serialization are logged as error messages.
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

// Check.SetResultSkipped Marks a check as skipped with an optional reason
//
// When invoked, this method acquires the check’s mutex to ensure thread
// safety, then sets the result status to skipped unless the check has already
// been aborted. It records the provided reason for skipping, which can be used
// for reporting or debugging purposes.
func (check *Check) SetResultSkipped(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	if check.Result == CheckResultAborted {
		return
	}

	check.Result = CheckResultSkipped
	check.skipReason = reason
}

// Check.SetResultError Marks a check as failed with an error reason
//
// This method locks the check’s mutex, verifies that it has not already been
// aborted or marked as an error, then sets its result to error and records the
// supplied reason. If the check is already in an error state, a warning log is
// emitted instead of changing the state.
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

// Check.SetResultAborted Marks a check as aborted with a reason
//
// This method records that the check has been aborted, setting its result state
// accordingly. It stores the provided abort reason for later reference and
// protects the update with a mutex to ensure thread safety.
func (check *Check) SetResultAborted(reason string) {
	check.mutex.Lock()
	defer check.mutex.Unlock()

	check.Result = CheckResultAborted
	check.skipReason = reason
}

// Check.Run Runs a check through its pre‑check, main, and post‑check stages
//
// The method first validates the receiver and any prior errors, then signals
// that the check is starting and records timestamps. It executes an optional
// before function, followed by the core check function, and finally an optional
// after function, each returning an error if they fail. If all stages succeed,
// it prints the final result based on the check's outcome and returns nil;
// otherwise it propagates the encountered error.
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

// printCheckResult Displays the final status of a check
//
// The function examines the result field of a check object and calls an
// appropriate CLI helper to print a formatted message indicating pass, fail,
// skip, abort or error. It uses the check's ID and any skip reason when
// relevant, ensuring that the output line is cleared before printing.
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
