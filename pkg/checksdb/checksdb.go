package checksdb

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

// Checks DB
//
// For each CheckGroup:
// 1. Calls group.BeforeAll()
//
//   For each Check in a CheckGroup
//   1. Calls group.BeforeEach(). -> get/refresh environment.
//     2. Calls check.SkipCheckFn() -> if true, go to next check.
//     3. Calls check.BeforeCheckFn()
//       4. Calls check.Run() -> Actual CNF Cert requirement check function.
//     5. Calls check.AfterCheckFn() -> Set env to be refreshed/cleanup.
//   6. Calls group.AfterEach()
//
// 2. Calls group.AfterAll()

var (
	dbLock    sync.Mutex
	db        []*Check
	dbByGroup map[string]*ChecksGroup

	resultsDb = map[string][]claim.Result{}
)

func AddCheck(check *Check) {
	db = append(db, check)
}

func RunChecks(labelsExpr string, timeout time.Duration) error {
	dbLock.Lock()
	defer dbLock.Unlock()

	// Timeout channel
	timeOutChan := time.After(timeout)
	// SIGINT(ctrl+c)/SIGTERM capture channel.
	sigIntChan := make(chan os.Signal, 10)
	signal.Notify(sigIntChan, syscall.SIGINT, syscall.SIGTERM)

	//  Labels expression parser not implemented yet. Assume labelsExpr is just a label.
	abort := false
	var abortReason string
	var errs []error
	for _, group := range dbByGroup {
		if abort {
			group.OnAbort(labelsExpr, abortReason)
			continue
		}

		// Stop channel, so we can send a stop signal to group.RunChecks()
		stopChan := make(chan bool, 1)

		// Done channel for the goroutine that runs group.RunChecks().
		groupDone := make(chan bool)
		go func() {
			errs = append(errs, group.RunChecks(labelsExpr, stopChan)...)
			groupDone <- true
		}()

		select {
		case <-groupDone:
			logrus.Tracef("Group %s finished running checks.", group.name)
		case <-timeOutChan:
			logrus.Warnf("Running all checks timed-out.")
			stopChan <- true

			abort = true
			abortReason = "global time-out"
			group.OnAbort(labelsExpr, abortReason)
		case <-sigIntChan:
			logrus.Warnf("SIGINT/SIGTERM received.")
			stopChan <- true

			abort = true
			abortReason = "SIGINT/SIGTERM"
			group.OnAbort(labelsExpr, abortReason)
		}

		group.RecordChecksResults()
	}

	if len(errs) > 0 {
		logrus.Errorf("RunChecks errors: %v", errs)
		return fmt.Errorf("%d errors found in checks/groups", len(errs))
	}

	return nil
}

func recordCheckResult(check *Check) {
	claimID, ok := identifiers.TestIDToClaimID[check.ID]
	if !ok {
		logrus.Fatalf("TestID %s has no corresponding Claim ID", check.ID)
	}

	logrus.Infof("Recording result %q fo check %s, claimID: %+v", check.Result, check.ID, claimID)
	resultsDb[check.ID] = append(resultsDb[check.ID], claim.Result{
		TestID:             &claimID,
		State:              check.Result.String(),
		StartTime:          check.StartTime.String(),
		EndTime:            check.EndTime.String(),
		Duration:           int(check.EndTime.Sub(check.StartTime).Seconds()),
		FailureReason:      check.FailureReason,
		CapturedTestOutput: check.CapturedOutput,

		CategoryClassification: &claim.CategoryClassification{
			Extended: identifiers.Catalog[claimID].CategoryClassification[identifiers.Extended],
			FarEdge:  identifiers.Catalog[claimID].CategoryClassification[identifiers.FarEdge],
			NonTelco: identifiers.Catalog[claimID].CategoryClassification[identifiers.NonTelco],
			Telco:    identifiers.Catalog[claimID].CategoryClassification[identifiers.Telco]},
		CatalogInfo: &claim.CatalogInfo{
			Description:           identifiers.Catalog[claimID].Description,
			Remediation:           identifiers.Catalog[claimID].Remediation,
			BestPracticeReference: identifiers.Catalog[claimID].BestPracticeReference,
			ExceptionProcess:      identifiers.Catalog[claimID].ExceptionProcess,
		},
	})
}

// GetReconciledResults is a function added to aggregate a Claim's results.  Due to the limitations of
// test-network-function-claim's Go Client, results are generalized to map[string]interface{}.
func GetReconciledResults() map[string]interface{} {
	resultMap := make(map[string]interface{})
	for key, vals := range resultsDb {
		// initializes the result map, if necessary
		if _, ok := resultMap[key]; !ok {
			resultMap[key] = make([]claim.Result, 0)
		}
		for _, val := range vals { //nolint:gocritic // Only done once at the end
			resultMap[key] = append(resultMap[key].([]claim.Result), val)
		}
	}
	return resultMap
}
