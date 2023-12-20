package checksdb

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/test-network-function/cnf-certification-test/internal/cli"
	"github.com/test-network-function/cnf-certification-test/internal/log"
)

const (
	checkIdxNone = -1
)

type ChecksGroup struct {
	name   string
	checks []*Check

	beforeAllFn, afterAllFn func(checks []*Check) error

	beforeEachFn, afterEachFn func(check *Check) error

	currentRunningCheckIdx int
}

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

func (group *ChecksGroup) WithBeforeAllFn(beforeAllFn func(checks []*Check) error) *ChecksGroup {
	group.beforeAllFn = beforeAllFn

	return group
}

func (group *ChecksGroup) WithBeforeEachFn(beforeEachFn func(check *Check) error) *ChecksGroup {
	group.beforeEachFn = beforeEachFn

	return group
}

func (group *ChecksGroup) WithAfterEachFn(afterEachFn func(check *Check) error) *ChecksGroup {
	group.afterEachFn = afterEachFn

	return group
}

func (group *ChecksGroup) WithAfterAllFn(afterAllFn func(checks []*Check) error) *ChecksGroup {
	group.afterAllFn = afterAllFn

	return group
}

func (group *ChecksGroup) Add(check *Check) {
	dbLock.Lock()
	defer dbLock.Unlock()

	group.checks = append(group.checks, check)
}

func skipCheck(check *Check, reason string) {
	check.LogInfo("Skipping check %s, reason: %s", check.ID, reason)

	check.SetResultSkipped(reason)
}

func skipAll(checks []*Check, reason string) {
	for _, check := range checks {
		skipCheck(check, reason)
		printCheckResult(check)
	}
}

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

func runCheck(check *Check, group *ChecksGroup, remainingChecks []*Check) (err error) {
	check.LogInfo("Running check")
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

// Runs all the checks in the group whose labels match the label expression filter.
//  1. Calls group.BeforeAll(). Then, for each Check in the group:
//  2. Calls group.BeforeEach()  -> normally used to get/refresh the test environment variable.
//  3. Calls check.SkipCheckFn() -> if true, skip the check.Run() (step 4)
//  4. Calls check.Run() -> Will call the actual CNF Cert requirement check function.
//  5. Calls group.AfterEach()
//  6. Calls group.AfterAll()
//
// Issues/errors/panics:
//   - BeforeAll panic/error: Set first check as error. Run AfterAll()
//   - BeforeEach panic/error: Set check as error and skip remaining. Skip check.Run(), run AfterEach + AfterAll.
//   - Check.Run() panic/error:  Set check as panicked. Run AfterEach + AfterAll
//   - AfterEach panic: Set check as error.
//
//nolint:funlen
func (group *ChecksGroup) RunChecks(labelsExpr string, stopChan <-chan bool, abortChan chan string) (errs []error, failedChecks int) {
	log.Info("Running group %q checks.", group.name)
	fmt.Printf("Running suite %s\n", strings.ToUpper(group.name))

	labelsExprEvaluator, err := NewLabelsExprEvaluator(labelsExpr)
	if err != nil {
		return []error{fmt.Errorf("invalid labels expression: %v", err)}, 0
	}

	// Get checks to run based on the label expr.
	checks := []*Check{}
	for _, check := range group.checks {
		if !labelsExprEvaluator.Eval(check.Labels) {
			skipCheck(check, "no matching labels")
			printCheckResult(check)
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

func (group *ChecksGroup) OnAbort(labelsExpr, abortReason string) error {
	labelsExprEvaluator, err := NewLabelsExprEvaluator(labelsExpr)
	if err != nil {
		return fmt.Errorf("invalid labels expression: %v", err)
	}

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

func (group *ChecksGroup) RecordChecksResults() {
	log.Info("Recording checks results of group %s", group.name)
	for _, check := range group.checks {
		recordCheckResult(check)
	}
}
