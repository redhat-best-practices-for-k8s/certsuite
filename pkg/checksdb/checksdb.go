package checksdb

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/internal/cli"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

var (
	dbLock    sync.Mutex
	db        []*Check
	dbByGroup map[string]*ChecksGroup

	resultsDB = map[string]claim.Result{}
)

func AddCheck(check *Check) {
	db = append(db, check)
}

//nolint:funlen
func RunChecks(labelsExpr string, timeout time.Duration) (failedCtr int, err error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	// Timeout channel
	timeOutChan := time.After(timeout)
	// SIGINT(ctrl+c)/SIGTERM capture channel.
	const SIGINTBufferLen = 10
	sigIntChan := make(chan os.Signal, SIGINTBufferLen)
	signal.Notify(sigIntChan, syscall.SIGINT, syscall.SIGTERM)
	// turn off ctrl-c capture on exit
	defer signal.Stop(sigIntChan)

	//  Labels expression parser not implemented yet. Assume labelsExpr is just a label.
	abort := false
	var abortReason string
	var errs []error
	for _, group := range dbByGroup {
		if abort {
			// ToDo: remove labelexpr checking.
			_ = group.OnAbort(labelsExpr, abortReason)
			continue
		}

		// Stop channel, so we can send a stop signal to group.RunChecks()
		stopChan := make(chan bool, 1)
		abortChan := make(chan bool, 1)

		// Done channel for the goroutine that runs group.RunChecks().
		groupDone := make(chan bool)
		go func() {
			checks, failedCheckCtr := group.RunChecks(labelsExpr, stopChan, abortChan)
			failedCtr += failedCheckCtr
			errs = append(errs, checks...)
			groupDone <- true
		}()

		select {
		case <-groupDone:
			log.Debug("Group %s finished running checks.", group.name)
		case <-abortChan:
			log.Warn("Group %s aborted.", group.name)
			stopChan <- true

			abort = true
			abortReason = "Test suite aborted due to error"
			_ = group.OnAbort(labelsExpr, abortReason)
		case <-timeOutChan:
			log.Warn("Running all checks timed-out.")
			stopChan <- true

			abort = true
			abortReason = "global time-out"
			_ = group.OnAbort(labelsExpr, abortReason)
		case <-sigIntChan:
			log.Warn("SIGINT/SIGTERM received.")
			stopChan <- true

			abort = true
			abortReason = "SIGINT/SIGTERM"
			_ = group.OnAbort(labelsExpr, abortReason)
		}

		group.RecordChecksResults()
	}

	// Print the results in the CLI
	cli.PrintResultsTable(getResultsSummary())
	printFailedChecksLog()

	if len(errs) > 0 {
		log.Error("RunChecks errors: %v", errs)
		return 0, fmt.Errorf("%d errors found in checks/groups", len(errs))
	}

	return failedCtr, nil
}

func recordCheckResult(check *Check) {
	claimID, ok := identifiers.TestIDToClaimID[check.ID]
	if !ok {
		check.LogError("TestID %s has no corresponding Claim ID", check.ID)
		os.Exit(1)
	}

	log.Info("Recording result %q of check %s, claimID: %+v", check.Result, check.ID, claimID)
	resultsDB[check.ID] = claim.Result{
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
	}
}

// GetReconciledResults is a function added to aggregate a Claim's results.  Due to the limitations of
// test-network-function-claim's Go Client, results are generalized to map[string]interface{}.
func GetReconciledResults() map[string]interface{} {
	resultMap := make(map[string]interface{})
	//nolint:gocritic
	for key, val := range resultsDB {
		// initializes the result map, if necessary
		if _, ok := resultMap[key]; !ok {
			resultMap[key] = make([]claim.Result, 0)
		}

		resultMap[key] = val
	}
	return resultMap
}

const (
	PASSED  = 0
	FAILED  = 1
	SKIPPED = 2
)

func getResultsSummary() map[string][]int {
	results := make(map[string][]int)
	for groupName, group := range dbByGroup {
		groupResults := []int{0, 0, 0}
		for _, check := range group.checks {
			switch check.Result {
			case CheckResultPassed:
				groupResults[PASSED]++
			case CheckResultFailed:
				groupResults[FAILED]++
			case CheckResultSkipped:
				groupResults[SKIPPED]++
			}
		}
		results[groupName] = groupResults
	}
	return results
}

const nbColorSymbols = 9

func printFailedChecksLog() {
	for _, group := range dbByGroup {
		for _, check := range group.checks {
			if check.Result != CheckResultFailed {
				continue
			}
			logHeader := fmt.Sprintf("| "+cli.Cyan+"LOG (%s)"+cli.Reset+" |", check.ID)
			nbSymbols := utf8.RuneCountInString(logHeader) - nbColorSymbols
			fmt.Println(strings.Repeat("-", nbSymbols))
			fmt.Println(logHeader)
			fmt.Println(strings.Repeat("-", nbSymbols))
			checkLogs := check.GetLogs()
			if checkLogs == "" {
				fmt.Println("Empty log output")
			} else {
				fmt.Println(checkLogs)
			}
		}
	}
}

func GetResults() map[string]claim.Result {
	return resultsDB
}

func GetTestSuites() []string {
	// Collect all of the unique test suites from the resultsDB
	var suites []string
	for key := range resultsDB {
		// Only append to the slice if it does not already exist
		if !stringhelper.StringInSlice(suites, key, false) {
			suites = append(suites, key)
		}
	}
	return suites
}

func GetTotalTests() int {
	return len(resultsDB)
}

func GetTestsCountByState(state string) int {
	count := 0
	//nolint:gocritic
	for _, results := range resultsDB {
		if results.State == state {
			count++
		}
	}
	return count
}
