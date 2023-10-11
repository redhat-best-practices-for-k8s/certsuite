package checksdb

import (
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

type ChecksGroup struct {
	name   string
	checks []*Check

	beforAllFn, afterAllFn func(checks []*Check) error

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
		name:   groupName,
		checks: []*Check{},
	}
	dbByGroup[groupName] = group

	return group
}

func (g *ChecksGroup) WithBeforeAllFn(beforAllFn func(checks []*Check) error) *ChecksGroup {
	g.beforAllFn = beforAllFn

	return g
}

func (g *ChecksGroup) WithBeforeEachFn(beforeEachFn func(check *Check) error) *ChecksGroup {
	g.beforeEachFn = beforeEachFn

	return g
}

func (g *ChecksGroup) WithAfterEachFn(afterEachFn func(check *Check) error) *ChecksGroup {
	g.afterEachFn = afterEachFn

	return g
}

func (g *ChecksGroup) WithAfterAllFn(afterAllFn func(checks []*Check) error) *ChecksGroup {
	g.afterAllFn = afterAllFn

	return g
}

func (g *ChecksGroup) Add(check *Check) {
	dbLock.Lock()
	defer dbLock.Unlock()

	g.checks = append(g.checks, check)
}

func skipCheck(check *Check, reason string) {
	logrus.Infof("Skipping check %s, reason: %s", check.ID, reason)

	check.SetResultSkipped(reason)
	// recordCheckResult(check)
}

func skipAll(checks []*Check, reason string) {
	for _, check := range checks {
		skipCheck(check, reason)
	}
}

func onFailure(failureType, failureMsg string, group *ChecksGroup, currentCheck *Check, remainingChecks []*Check) error {
	// Set current Check's result as error.
	currentCheck.SetResultError(failureType + ": " + failureMsg)
	// Set the remaining checks as skipped, using a simplified reason msg.
	reason := "group " + group.name + " " + failureType
	skipAll(remainingChecks, reason)
	// Return generic error using the reason.
	return errors.New(reason)
}

func runBeforeAllFn(group *ChecksGroup, checks []*Check) (err error) {
	logrus.Tracef("GROUP %s - Running beforeAll", group.name)
	if group.beforAllFn == nil {
		return nil
	}

	firstCheck := checks[0]
	defer func() {
		if r := recover(); r != nil {
			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())
			logrus.Errorf("Panic while running beforeAll function:\n%v", stackTrace)
			// Set first check's result as error and skip the remaining ones.
			err = onFailure("beforeAll function panicked", "\n:"+stackTrace, group, firstCheck, checks)
		}
	}()

	if err := group.beforAllFn(checks); err != nil {
		logrus.Errorf("Unexpected error while running beforeAll function: %v", err)
		// Set first check's result as error and skip the remaining ones.
		return onFailure("beforeAll function unexpected error", err.Error(), group, firstCheck, checks)
	}

	return nil
}

func runAfterAllFn(group *ChecksGroup, checks []*Check) (err error) {
	logrus.Tracef("GROUP %s - Running afterAll", group.name)

	if group.afterAllFn == nil {
		return nil
	}

	lastCheck := checks[len(checks)-1]
	zeroRemainingChecks := []*Check{}
	defer func() {
		if r := recover(); r != nil {
			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())
			logrus.Errorf("Panic while running afterAll function:\n%v", stackTrace)
			// Set last check's result as error, no need to skip anyone.
			err = onFailure("afterAll function panicked", "\n: "+stackTrace, group, lastCheck, zeroRemainingChecks)
		}
	}()

	if err := group.afterAllFn(group.checks); err != nil {
		logrus.Errorf("Unexpected error while running afterAll function: %v", err.Error())
		// Set last check's result as error, no need to skip anyone.
		return onFailure("afterAll function unexpected error", err.Error(), group, lastCheck, zeroRemainingChecks)
	}

	return nil
}

func runBeforeEachFn(group *ChecksGroup, check *Check, remainingChecks []*Check) (err error) {
	logrus.Tracef("GROUP %s - Running beforeEach for check %s", group.name, check.ID)
	if group.beforeEachFn == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())
			logrus.Errorf("Panic while running beforeEach function:\n%v", stackTrace)
			// Set last check's result as error, no need to skip anyone.
			err = onFailure("beforeEach function panicked", "\n: "+stackTrace, group, check, remainingChecks)
		}
	}()

	if err := group.beforeEachFn(check); err != nil {
		logrus.Errorf("Unexpected error while running beforeEach function:\n%v", err.Error())
		// Set last check's result as error, no need to skip anyone.
		return onFailure("beforeEach function unexpected error", err.Error(), group, check, remainingChecks)
	}

	return nil
}

func runAfterEachFn(group *ChecksGroup, check *Check, remainingChecks []*Check) (err error) {
	logrus.Tracef("GROUP %s - Running afterEach for check %s", group.name, check.ID)

	if group.afterEachFn == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())
			logrus.Errorf("Panic while running afterEach function:\n%v", stackTrace)
			// Set last check's result as error, no need to skip anyone.
			err = onFailure("afterEach function panicked", "\n: "+stackTrace, group, check, remainingChecks)
		}
	}()

	if err := group.afterEachFn(check); err != nil {
		logrus.Errorf("Unexpected error while running afterEach function:\n%v", err.Error())
		// Set last check's result as error, no need to skip anyone.
		return onFailure("afterEach function unexpected error", err.Error(), group, check, remainingChecks)
	}

	return nil
}

func shouldSkipCheck(check *Check) (skip bool, reason string) {
	if check.SkipCheckFn == nil {
		return false, ""
	}

	logrus.Tracef("Running check %s skipCheck function.", check.ID)
	return check.SkipCheckFn()
}

func runCheck(check *Check, group *ChecksGroup, remainingChecks []*Check) (err error) {
	logrus.Infof("Running check %s", check.ID)
	defer func() {
		if r := recover(); r != nil {
			stackTrace := fmt.Sprint(r) + "\n" + string(debug.Stack())

			logrus.Errorf("Panic while running check %s function:\n%v", check.ID, stackTrace)
			err = onFailure("check "+check.ID+" function panic", stackTrace, group, check, remainingChecks)
		}
	}()

	if err := check.Run(); err != nil {
		logrus.Errorf("Unexpected error while running check %s function: %v", check.ID, err.Error())
		return onFailure("check "+check.ID+" function unexpected error", err.Error(), group, check, remainingChecks)
	}

	return nil
}

// Runs all the checks in the group whose labels match the label
// expression. Issues/errors/panics:
//   - BeforeAll panic/error: Set first check as error. Run AfterAll()
//   - BeforeEach panic/error: Set check as error and skip remaining. Skip check.Run(), run AfterEach + AfterAll.
//   - Check.Run() panic/error:  Set check as panicked. Run AfterEach + AfterAll
//   - AfterEach panic: Set check as error.
func (group *ChecksGroup) RunChecks(labelsExpr string, stopChan <-chan bool) (errs []error) {
	logrus.Infof("Running group %q checks.", group.name)

	labelsExprEvaluator, err := NewLabelsExprEvaluator(labelsExpr)
	if err != nil {
		return []error{fmt.Errorf("invalid labels expression: %v", err)}
	}

	// Get checks to run based on the label expr.
	checks := []*Check{}
	for _, check := range group.checks {
		if !labelsExprEvaluator.Eval(check.Labels) {
			skipCheck(check, "Not matching labels")
			continue
		}
		checks = append(checks, check)
	}

	if len(checks) == 0 {
		// No check matched the labels expression.
		// skipAll(checks, "Not matching labels")
		return nil
	}

	// Run afterAllFn always, no matter previous panics/crashes.
	defer func() {
		if err := runAfterAllFn(group, checks); err != nil {
			errs = append(errs, err)
		}
	}()

	if err := runBeforeAllFn(group, checks); err != nil {
		errs = append(errs, err)
		return errs
	}

	logrus.Infof("Checks to run: %d (group's total=%d)", len(checks), len(group.checks))
	group.currentRunningCheckIdx = 0
	for i, check := range checks {
		// Fast stop in case the stop (abort/timeout) signal received.
		select {
		case <-stopChan:
			return errs
		default:
		}

		// Create a remainingChecks list excluding the current check.
		remainingChecks := []*Check{}
		if i+1 < len(checks) {
			remainingChecks = checks[i+1:]
		}

		beforeEachFailed := false
		if err := runBeforeEachFn(group, check, remainingChecks); err != nil {
			beforeEachFailed = true
			errs = append(errs, err)
		}

		if !beforeEachFailed {
			// Should we skip this check?
			skip, reason := shouldSkipCheck(check)
			if skip {
				skipCheck(check, reason)
			} else {
				err := runCheck(check, group, remainingChecks)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
		// afterEach func must run even if the check was skipped or panicked/unexpected error.
		afterEachFailed := false
		if err := runAfterEachFn(group, check, remainingChecks); err != nil {
			errs = append(errs, err)
			afterEachFailed = true
		}

		if beforeEachFailed || afterEachFailed {
			break
		}

		group.currentRunningCheckIdx++
	}

	return errs
}

func (group *ChecksGroup) OnAbort(labelsExpr string, abortReason string) error {
	labelsExprEvaluator, err := NewLabelsExprEvaluator(labelsExpr)
	if err != nil {
		return fmt.Errorf("invalid labels expression: %v", err)
	}

	for i, check := range group.checks {
		if !labelsExprEvaluator.Eval(check.Labels) {
			continue
		}

		if i == group.currentRunningCheckIdx {
			check.SetResultAborted(abortReason)

		} else if i > group.currentRunningCheckIdx {
			check.SetResultSkipped(abortReason)
		}
	}

	return nil
}

func (group *ChecksGroup) RecordChecksResults() {
	logrus.Infof("Recording checks results of group %s", group.name)
	for _, check := range group.checks {
		recordCheckResult(check)
	}
}
