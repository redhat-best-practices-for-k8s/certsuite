package checksdb

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/cli"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

const (
	checkIdxNone = -1
)

// ChecksGroup Holds a collection of checks and orchestrates their execution
//
// This structure stores the group's name, the list of checks to run, and
// optional callback functions for before/after all and before/after each check.
// It tracks which check is currently executing to handle aborts or failures
// correctly. The group provides methods to add checks, run them with support
// for labeling, and record results.
type ChecksGroup struct {
	name   string
	checks []*Check

	beforeAllFn, afterAllFn func(checks []*Check) error

	beforeEachFn, afterEachFn func(check *Check) error

	currentRunningCheckIdx int
}

// NewChecksGroup creates or retrieves a checks group by name
//
// This function locks the global database, ensuring thread safety while
// accessing the map of groups. It initializes the map if necessary, then looks
// up an existing group with the given key. If found it returns that instance;
// otherwise it constructs a new ChecksGroup with default fields, stores it in
// the map, and returns it.
func NewChecksGroup(groupName string) *ChecksGroup {
	dbLock.Lock()
	defer dbLock.Unlock()

	if dbByGroup == nil {
		dbByGroup = map[string]*ChecksGroup{}
	}

	group, exists := dbByGroup[groupName]
	if exists {
		return group
	}

	group = &ChecksGroup{
		name:                   groupName,
		checks:                 []*Check{},
		currentRunningCheckIdx: checkIdxNone,
	}
	dbByGroup[groupName] = group

	return group
}

// ChecksGroup.WithBeforeAllFn Registers a function to run before all checks
//
// This method assigns the provided callback to the group, which will be
// executed with the slice of checks prior to any other operations. It returns
// the modified group for chaining purposes.
func (group *ChecksGroup) WithBeforeAllFn(beforeAllFn func(checks []*Check) error) *ChecksGroup {
	group.beforeAllFn = beforeAllFn

	return group
}

// ChecksGroup.WithBeforeEachFn Assigns a callback to execute prior to each check
//
// This method accepts a function that takes a check pointer and may return an
// error. It stores this function in the group's internal field so it will be
// invoked before each individual check runs. The group instance is returned,
// allowing further chained configuration calls.
func (group *ChecksGroup) WithBeforeEachFn(beforeEachFn func(check *Check) error) *ChecksGroup {
	group.beforeEachFn = beforeEachFn

	return group
}

// ChecksGroup.WithAfterEachFn Assigns a function that runs after every individual check
//
// This method stores the provided function as the group's post‑check hook,
// ensuring it is invoked with a reference to each Check object once the check
// completes. The stored callback can modify or inspect the check before the
// group continues processing. It returns the same ChecksGroup instance for
// chaining.
func (group *ChecksGroup) WithAfterEachFn(afterEachFn func(check *Check) error) *ChecksGroup {
	group.afterEachFn = afterEachFn

	return group
}

// ChecksGroup.WithAfterAllFn Assigns a callback to execute after all checks complete
//
// This method stores the supplied function in the ChecksGroup so it will be
// called with the list of executed checks once processing is finished. The
// stored function can perform cleanup or result aggregation. It returns the
// same group instance, allowing method chaining.
func (group *ChecksGroup) WithAfterAllFn(afterAllFn func(checks []*Check) error) *ChecksGroup {
	group.afterAllFn = afterAllFn

	return group
}

// ChecksGroup.Add Adds a check to the group
//
// This method acquires a global lock, appends the provided check to the group's
// internal slice, and then releases the lock. It ensures thread‑safe
// modification of the checks collection while keeping the operation simple and
// efficient.
func (group *ChecksGroup) Add(check *Check) {
	dbLock.Lock()
	defer dbLock.Unlock()

	group.checks = append(group.checks, check)
}

// skipCheck Marks a check as skipped with a reason
//
// This function records an informational message indicating that the specified
// check will not be executed due to the supplied reason. It then updates the
// check’s status to skipped and displays the outcome using the standard
// output routine.
func skipCheck(check *Check, reason string) {
	check.LogInfo("Skipping check %s, reason: %s", check.ID, reason)
	check.SetResultSkipped(reason)
	printCheckResult(check)
}

// skipAll marks all remaining checks as skipped with a given reason
//
// This routine iterates over a slice of check objects, calling an internal
// helper for each one to log the skip action, set its result state to skipped,
// and output its status. The provided reason string is passed unchanged to
// every check so that downstream reporting can identify why the checks were not
// executed. No value is returned; the function simply updates each check's
// internal state.
func skipAll(checks []*Check, reason string) {
	for _, check := range checks {
		skipCheck(check, reason)
	}
}

// onFailure Handles a failure during group or check execution
//
// When a before/after or check function fails, this routine marks the current
// check as an error with a descriptive message. It then skips all remaining
// checks in the same group using a concise skip reason. Finally it returns a
// generic error that indicates which failure type occurred.
func onFailure(failureType, failureMsg string, group *ChecksGroup, currentCheck *Check, remainingChecks []*Check) error {
	// Set current Check's result as error.
	fmt.Printf("\r[ %s ] %-60s\n", cli.CheckResultTagError, currentCheck.ID)
	currentCheck.SetResultError(failureType + ": " + failureMsg)
	// Set the remaining checks as skipped, using a simplified reason msg.
	reason := "group " + group.name + " " + failureType
	skipAll(remainingChecks, reason)
	// Return generic error using the reason.
	return errors.New(reason)
}

// runBeforeAllFn Executes a group-wide setup routine before any checks run
//
// This function calls the optional beforeAllFn defined on a ChecksGroup,
// passing all checks to it. If the function panics or returns an error, the
// first check is marked as failed and all remaining checks are skipped with an
// explanatory reason. No other actions occur if beforeAllFn is nil.
func runBeforeAllFn(group *ChecksGroup, checks []*Check) (err error) {
	log.Debug("GROUP %s - Running beforeAll", group.name)
	if group.beforeAllFn == nil {
		return nil
	}

	firstCheck := checks[0]
	defer func() {
		if r := recover(); r != nil {
			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())
			log.Error("Panic while running beforeAll function:\n%v", stackTrace)
			// Set first check's result as error and skip the remaining ones.
			err = onFailure("beforeAll function panicked", "\n:"+stackTrace, group, firstCheck, checks)
		}
	}()

	if err := group.beforeAllFn(checks); err != nil {
		log.Error("Unexpected error while running beforeAll function: %v", err)
		// Set first check's result as error and skip the remaining ones.
		return onFailure("beforeAll function unexpected error", err.Error(), group, firstCheck, checks)
	}

	return nil
}

// runAfterAllFn Executes the group's final cleanup routine
//
// When a checks group has finished running all its checks, this function
// invokes any registered afterAll hook with the entire list of checks. It logs
// the start and handles both panics and returned errors by marking the last
// executed check as failed and preventing further actions. The result is an
// error if the cleanup fails; otherwise nil.
func runAfterAllFn(group *ChecksGroup, checks []*Check) (err error) {
	log.Debug("GROUP %s - Running afterAll", group.name)

	if group.afterAllFn == nil {
		return nil
	}

	lastCheck := checks[len(checks)-1]
	zeroRemainingChecks := []*Check{}
	defer func() {
		if r := recover(); r != nil {
			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())
			log.Error("Panic while running afterAll function:\n%v", stackTrace)
			// Set last check's result as error, no need to skip anyone.
			err = onFailure("afterAll function panicked", "\n: "+stackTrace, group, lastCheck, zeroRemainingChecks)
		}
	}()

	if err := group.afterAllFn(group.checks); err != nil {
		log.Error("Unexpected error while running afterAll function: %v", err.Error())
		// Set last check's result as error, no need to skip anyone.
		return onFailure("afterAll function unexpected error", err.Error(), group, lastCheck, zeroRemainingChecks)
	}

	return nil
}

// runBeforeEachFn Executes a group’s beforeEach hook for a specific check
//
// This function runs the optional beforeEachFn defined on a ChecksGroup,
// passing it the current Check. It captures panics or returned errors, logs
// diagnostic information, and records the failure by marking the check as
// errored and skipping subsequent checks. If no issues occur, the function
// simply returns nil.
func runBeforeEachFn(group *ChecksGroup, check *Check, remainingChecks []*Check) (err error) {
	log.Debug("GROUP %s - Running beforeEach for check %s", group.name, check.ID)
	if group.beforeEachFn == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())
			log.Error("Panic while running beforeEach function:\n%v", stackTrace)
			// Set last check's result as error, no need to skip anyone.
			err = onFailure("beforeEach function panicked", "\n: "+stackTrace, group, check, remainingChecks)
		}
	}()

	if err := group.beforeEachFn(check); err != nil {
		log.Error("Unexpected error while running beforeEach function:\n%v", err.Error())
		// Set last check's result as error, no need to skip anyone.
		return onFailure("beforeEach function unexpected error", err.Error(), group, check, remainingChecks)
	}

	return nil
}

// runAfterEachFn Handles post‑check cleanup and error reporting
//
// This routine runs a group's afterEach function for each check, logging its
// start and capturing any panic or returned error. If the function panics, it
// logs the stack trace and marks the current check as failed without skipping
// subsequent checks. On a normal error, it reports the issue, sets the check
// result to an error state, and returns that error.
func runAfterEachFn(group *ChecksGroup, check *Check, remainingChecks []*Check) (err error) {
	log.Debug("GROUP %s - Running afterEach for check %s", group.name, check.ID)

	if group.afterEachFn == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())
			log.Error("Panic while running afterEach function:\n%v", stackTrace)
			// Set last check's result as error, no need to skip anyone.
			err = onFailure("afterEach function panicked", "\n: "+stackTrace, group, check, remainingChecks)
		}
	}()

	if err := group.afterEachFn(check); err != nil {
		log.Error("Unexpected error while running afterEach function:\n%v", err.Error())
		// Set last check's result as error, no need to skip anyone.
		return onFailure("afterEach function unexpected error", err.Error(), group, check, remainingChecks)
	}

	return nil
}

// shouldSkipCheck decides whether a check should be skipped based on its skip functions
//
// The function evaluates each user-provided skip function, collecting any
// reasons for skipping. If any reason is found, it applies the check's SkipMode
// policy: SkipModeAny skips if at least one reason exists, while SkipModeAll
// requires all skip functions to indicate a skip. The function also recovers
// from panics in skip functions, logs an error, and treats that as a skip with
// a panic reason.
func shouldSkipCheck(check *Check) (skip bool, reasons []string) {
	if len(check.SkipCheckFns) == 0 {
		return false, []string{}
	}

	// Short-circuit
	if len(check.SkipCheckFns) == 0 {
		return false, []string{}
	}

	// Save the skipFn index in case it panics so it can be used in the log trace.
	currentSkipFnIndex := 0

	defer func() {
		if r := recover(); r != nil {
			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())
			check.LogError("Skip check function (idx=%d) panic'ed: %s", currentSkipFnIndex, stackTrace)
			skip = true
			reasons = []string{fmt.Sprintf("skipCheckFn (idx=%d) panic:\n%s", currentSkipFnIndex, stackTrace)}
		}
	}()

	// Call all the skip functions first.
	for _, skipFn := range check.SkipCheckFns {
		if skip, reason := skipFn(); skip {
			reasons = append(reasons, reason)
		}
		currentSkipFnIndex++
	}

	// If none of the skipFn returned true, exit now.
	if len(reasons) == 0 {
		return false, []string{}
	}

	// Now we need to check the skipMode for this check.
	switch check.SkipMode {
	case SkipModeAny:
		return true, reasons
	case SkipModeAll:
		// Only skip if all the skipFn returned true.
		if len(reasons) == len(check.SkipCheckFns) {
			return true, reasons
		}
		return false, []string{}
	}

	return false, []string{}
}

// runCheck Executes a check with error handling and panic recovery
//
// The function runs the provided check, capturing any panics or errors that
// occur during its execution. If a panic is detected, it distinguishes between
// an intentional abort and unexpected failures, logs detailed information, and
// marks subsequent checks as skipped. Successful completion returns nil, while
// any failure results in an error describing the issue.
func runCheck(check *Check, group *ChecksGroup, remainingChecks []*Check) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Don't do anything in case the check was manually aborted by check.Abort().
			if msg, ok := r.(AbortPanicMsg); ok {
				log.Warn("Check was manually aborted, msg: %v", msg)
				err = fmt.Errorf("%v", msg)
				return
			}

			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())

			check.LogError("Panic while running check %s function:\n%v", check.ID, stackTrace)
			err = onFailure(fmt.Sprintf("check %s function panic", check.ID), stackTrace, group, check, remainingChecks)
		}
	}()

	if err := check.Run(); err != nil {
		check.LogError("Unexpected error while running check %s function: %v", check.ID, err.Error())
		return onFailure(fmt.Sprintf("check %s function unexpected error", check.ID), err.Error(), group, check, remainingChecks)
	}

	return nil
}

// ChecksGroup.RunChecks Executes a filtered set of checks with lifecycle hooks
//
// The method gathers checks whose labels match the group’s filter, then runs
// them in order while invoking BeforeAll, BeforeEach, AfterEach, and AfterAll
// callbacks. It handles skipping logic, abort signals, and panics by recording
// errors or marking checks as skipped/failed. The function returns any
// collected errors and a count of failed checks.
//
//nolint:funlen
func (group *ChecksGroup) RunChecks(stopChan <-chan bool, abortChan chan string) (errs []error, failedChecks int) {
	log.Info("Running group %q checks.", group.name)
	fmt.Printf("Running suite %s\n", strings.ToUpper(group.name))

	// Get checks to run based on the label expr.
	checks := []*Check{}
	for _, check := range group.checks {
		if !labelsExprEvaluator.Eval(check.Labels) {
			skipCheck(check, "no matching labels")
			continue
		}
		checks = append(checks, check)
	}

	if len(checks) == 0 {
		return nil, 0
	}

	// Run afterAllFn always, no matter previous panics/crashes.
	defer func() {
		if err := runAfterAllFn(group, checks); err != nil {
			errs = append(errs, err)
		}
	}()

	if err := runBeforeAllFn(group, checks); err != nil {
		errs = append(errs, err)
		return errs, 0
	}

	log.Info("Checks to run: %d (group's total=%d)", len(checks), len(group.checks))
	group.currentRunningCheckIdx = 0
	for i, check := range checks {
		// Fast stop in case the stop (abort/timeout) signal received.
		select {
		case <-stopChan:
			return nil, 0
		default:
		}

		// Create a remainingChecks list excluding the current check.
		remainingChecks := []*Check{}
		if i+1 < len(checks) {
			remainingChecks = checks[i+1:]
		}

		if err := runBeforeEachFn(group, check, remainingChecks); err != nil {
			errs = []error{err}
		}

		if len(errs) == 0 {
			// Should we skip this check?
			skip, reasons := shouldSkipCheck(check)
			if skip {
				skipCheck(check, strings.Join(reasons, ", "))
			} else {
				check.SetAbortChan(abortChan) // Set the abort channel for the check.
				err := runCheck(check, group, remainingChecks)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
		// afterEach func must run even if the check was skipped or panicked/unexpected error.
		if err := runAfterEachFn(group, check, remainingChecks); err != nil {
			errs = append(errs, err)
		}

		// Don't run more checks if any of beforeEach, the checkFn or afterEach functions errored/panicked.
		if len(errs) > 0 {
			break
		}

		// Increment the failed checks counter.
		if check.Result.String() == CheckResultFailed {
			failedChecks++
		}

		group.currentRunningCheckIdx++
	}

	return errs, failedChecks
}

// ChecksGroup.OnAbort Handles a group’s abort by setting check results accordingly
//
// When an abort occurs, this method iterates over all checks in the group.
// Checks that do not match labels are marked as skipped with a label reason. If
// no check had started yet, every remaining check is skipped with the abort
// reason; otherwise the currently running check is marked aborted and
// subsequent checks are skipped. Each result is printed immediately.
func (group *ChecksGroup) OnAbort(abortReason string) error {
	// If this wasn't the group with the aborted check.
	if group.currentRunningCheckIdx == checkIdxNone {
		fmt.Printf("Skipping checks from suite %s\n", strings.ToUpper(group.name))
	}

	for i, check := range group.checks {
		if !labelsExprEvaluator.Eval(check.Labels) {
			check.SetResultSkipped("not matching labels")
			continue
		}

		// If none of this group's checks was running yet, skip all.
		if group.currentRunningCheckIdx == checkIdxNone {
			check.SetResultSkipped(abortReason)
			continue
		}

		// Abort the check that was running when it was aborted and skip the rest.
		if i == group.currentRunningCheckIdx {
			check.SetResultAborted(abortReason)
		} else if i > group.currentRunningCheckIdx {
			check.SetResultSkipped(abortReason)
		}

		printCheckResult(check)
	}

	return nil
}

// ChecksGroup.RecordChecksResults Logs each check result and stores it in the results database
//
// The method iterates over all checks in the group, invoking a helper that logs
// information about the test ID, state, and duration. For each check, it
// records the outcome in a shared map keyed by the test identifier, including
// metadata such as timestamps, skip reasons, and catalog references. This
// ensures that results are persisted for later reporting or further processing.
func (group *ChecksGroup) RecordChecksResults() {
	log.Info("Recording checks results of group %s", group.name)
	for _, check := range group.checks {
		recordCheckResult(check)
	}
}
