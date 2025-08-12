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

// ChecksGroup represents a collection of checks that can be executed with optional setup and teardown hooks.
//
// ChecksGroup holds a named group of Check pointers and optional callback functions
// that run before all checks, after all checks, before each individual check,
// and after each individual check. The group tracks the index of the currently
// running check so that hooks and abort handling can reference it.
//
// Fields:
//   name                identifies the group.
//   checks              slice of pointers to Check structs that will be executed.
//   currentRunningCheckIdx holds the index of the check currently being run,
//     or -1 if no check is active.
//   beforeAllFn, afterAllFn are optional functions that receive the entire
//     checks slice and return an error. They are called once at the start and
//     end of RunChecks respectively.
//   beforeEachFn, afterEachFn are optional functions that receive a single
//     Check pointer and return an error. They are called before and after each
//     individual check.
//
// Methods provide fluent configuration of these hooks and add checks to the group,
// while RunChecks executes all checks respecting skip logic, abort signals,
// and error handling defined by the callbacks.
type ChecksGroup struct {
	name   string
	checks []*Check

	beforeAllFn, afterAllFn func(checks []*Check) error

	beforeEachFn, afterEachFn func(check *Check) error

	currentRunningCheckIdx int
}

// NewChecksGroup creates or retrieves a ChecksGroup by name.
//
// It locks the global database mutex to ensure thread safety while accessing
// the map of existing groups. If a group with the given name already exists,
// it returns that instance; otherwise it constructs a new ChecksGroup, stores
// it in the map, and returns the newly created object. The returned pointer
// should be used for subsequent operations on that group.
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

// WithBeforeAllFn registers a function to run before all checks in the group.
//
// It accepts a callback that receives a slice of pointers to Check objects and
// returns an error if setup fails. The function is stored on the ChecksGroup
// instance and will be invoked automatically prior to executing any check
// within that group. The method returns the modified ChecksGroup for chaining.
func (group *ChecksGroup) WithBeforeAllFn(beforeAllFn func(checks []*Check) error) *ChecksGroup {
	group.beforeAllFn = beforeAllFn

	return group
}

// WithBeforeEachFn registers a function to be executed before each check in the group.
//
// It accepts a callback that receives a pointer to a Check and may return an error.
// The callback is stored on the ChecksGroup and will run for every check
// added to the group, allowing setup or validation logic to be applied
// consistently across all checks. The method returns the modified
// ChecksGroup to support chaining of configuration calls.
func (group *ChecksGroup) WithBeforeEachFn(beforeEachFn func(check *Check) error) *ChecksGroup {
	group.beforeEachFn = beforeEachFn

	return group
}

// WithAfterEachFn registers a function to be executed after each check in the group.
//
// It accepts a callback that receives a pointer to the Check being processed
// and may return an error. The function is stored on the ChecksGroup and
// will be called automatically after every individual check runs.
// The method returns the same ChecksGroup to allow chaining of configuration calls.
func (group *ChecksGroup) WithAfterEachFn(afterEachFn func(check *Check) error) *ChecksGroup {
	group.afterEachFn = afterEachFn

	return group
}

// WithAfterAllFn registers a function to be called after all checks in the group have run.
//
// It accepts a callback that receives the slice of checks executed in the group and may
// return an error. The method stores this function inside the ChecksGroup so it will
// be invoked automatically when the group's execution completes. The method returns
// the modified *ChecksGroup to allow chaining configuration calls.
func (group *ChecksGroup) WithAfterAllFn(afterAllFn func(checks []*Check) error) *ChecksGroup {
	group.afterAllFn = afterAllFn

	return group
}

// Add registers a check in the group.
//
// It appends the provided Check pointer to the group's internal slice,
// ensuring thread safety by locking the group's mutex during modification.
// No return value is produced; the operation modifies the receiver's state.
func (group *ChecksGroup) Add(check *Check) {
	dbLock.Lock()
	defer dbLock.Unlock()

	group.checks = append(group.checks, check)
}

// skipCheck marks a check as skipped and logs the action.
//
// It receives a pointer to a Check and an optional reason string.
// The function records the skipped status by calling SetResultSkipped,
// logs the skip with LogInfo, and outputs the result via printCheckResult.
func skipCheck(check *Check, reason string) {
	check.LogInfo("Skipping check %s, reason: %s", check.ID, reason)
	check.SetResultSkipped(reason)
	printCheckResult(check)
}

// skipAll marks a slice of checks as skipped and returns a cleanup function.
//
// skipAll records that each check in the provided slice should be marked as skipped
// with the supplied reason string. It calls skipCheck on every element and then
// returns a closure which, when invoked, can perform any necessary post‑processing.
// The returned function currently performs no additional work but preserves the
// pattern used for other group operations.
func skipAll(checks []*Check, reason string) {
	for _, check := range checks {
		skipCheck(check, reason)
	}
}

// onFailure handles a check failure scenario and records the error in the database.
//
// It receives an error message, a string identifier for the failed check,
// the group containing the check, the specific Check that failed,
// and any dependent checks that were impacted.
// The function logs the failure using Printf, marks the Check result as an error
// with SetResultError, and if the group's skip mode is set to SkipModeAll,
// it will invoke skipAll on all remaining checks in the group. Finally,
// it creates a new database entry for the failed check by calling New
// and returns any error that occurs during this process.
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

// runBeforeAllFn executes a group's beforeAll function and handles any panic that occurs.
//
// It receives the ChecksGroup to which the checks belong and a slice of Check pointers
// that will be executed. The function runs the group's beforeAllFn, capturing any
// panic as an error. If the function panics, it logs the stack trace and returns
// the recovered error. Otherwise, it simply returns nil. This helper is used to
// ensure that group-wide setup code does not cause a crash in the test runner.
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

// runAfterAllFn executes the after‑all callback of a ChecksGroup and handles any panics that occur.
//
// It receives a pointer to the group and a slice of checks that were executed.
// The function calls the group's afterAllFn, capturing errors and panics.
// If an error is returned or a panic occurs, it logs details using Debug
// and returns the resulting error. If no errors happen, nil is returned.
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

// runBeforeEachFn executes a before‑each function for a check group, handling errors and panics.
//
// runBeforeEachFn runs the before‑each hook associated with a checks group.
// It receives the parent group, the current check, and the list of sibling checks.
// The function logs debugging information, recovers from any panic that occurs
// in the before‑each callback, and returns an error if the callback fails or
// panics. On success it returns nil.
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

// runAfterEachFn executes an after‑each callback for a check and handles any panic that occurs.
//
// It receives the checks group, the current check, and a slice of all checks in the group.
// If the after‑each function panics, runAfterEachFn recovers and logs the stack trace,
// then returns an error describing the failure.  
// The returned error is used to mark the check as failed if it occurs during cleanup.
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

// shouldSkipCheck determines whether a check should be skipped.
//
// It examines the skip function of a Check, evaluates any panic that occurs,
// and collects error messages if the skip condition cannot be evaluated.
// The first return value indicates whether the check is to be skipped;
// the second return value contains any diagnostics produced during evaluation.
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

// runCheck executes a single check within its group context.
// It runs the check's Run method and handles any panic by converting it to an error.
// The function logs failures or skips based on the check result,
// updates internal state via onFailure callbacks,
// and returns any error that occurred during execution.
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

// RunChecks executes all checks in the group that match a label expression filter, managing setup and teardown around each check and handling errors or panics gracefully.
//
// It first calls the group's BeforeAll function to prepare any shared state.
// Then it iterates over every Check in the group, performing these steps for each:
//   1. Calls BeforeEach to set up per‑check environment (e.g., refreshing variables).
//   2. Invokes SkipCheckFn to determine if the check should be skipped; if so, the actual Run is omitted.
//   3. If not skipped, calls the Check's Run method which performs the CNF certification requirement test.
//   4. Calls AfterEach after each individual check finishes.
//
// After all checks have been processed, AfterAll is called to clean up shared resources.
//
// The function returns a slice of errors collected during execution and an integer count of
// checks that were actually run (excluding those skipped). Errors are recorded for any panic or
// failure in BeforeAll, BeforeEach, Run, or AfterEach. A cancel channel can be used to abort
// the process prematurely; if closed, subsequent checks will not execute. The string channel
// is used internally for logging progress and status messages.
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

// OnAbort handles the aborting of a check within a ChecksGroup.
//
// It receives an error message string, logs the abort event,
// evaluates any skip or abort conditions, updates the check's
// result status to either Skipped or Aborted accordingly,
// and prints the final result for that check. The function
// returns an error if any step in processing the abort fails.
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

// RecordChecksResults returns a function that records the results of all checks in this group.
//
// The returned function performs logging via Info and invokes
// recordCheckResult for each individual check, storing the outcomes
// in the package‑wide database. No parameters are required and no
// value is returned by the closure itself.
func (group *ChecksGroup) RecordChecksResults() {
	log.Info("Recording checks results of group %s", group.name)
	for _, check := range group.checks {
		recordCheckResult(check)
	}
}
