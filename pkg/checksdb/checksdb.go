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

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/cli"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/labels"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
)

var (
	dbLock    sync.Mutex
	dbByGroup map[string]*ChecksGroup

	resultsDB = map[string]claim.Result{}

	labelsExprEvaluator labels.LabelsExprEvaluator
)

type AbortPanicMsg string

// RunChecks executes all registered checks, respecting the provided timeout.
//
// It acquires a lock on the internal database, runs each check group,
// collects results, records them, and prints a summary table.
// The function blocks until either all checks finish or the timeout expires.
// If any check fails or an error occurs, it returns the number of failed
// checks along with an error describing the failure. On success it returns 0 and nil.
func RunChecks(timeout time.Duration) (failedCtr int, err error) {
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

	abort := false
	var abortReason string
	var errs []error
	for _, group := range dbByGroup {
		if abort {
			_ = group.OnAbort(abortReason)
			group.RecordChecksResults()
			continue
		}

		// Stop channel, so we can send a stop signal to group.RunChecks()
		stopChan := make(chan bool, 1)
		abortChan := make(chan string, 1)

		// Done channel for the goroutine that runs group.RunChecks().
		groupDone := make(chan bool)
		go func() {
			checks, failedCheckCtr := group.RunChecks(stopChan, abortChan)
			failedCtr += failedCheckCtr
			errs = append(errs, checks...)
			groupDone <- true
		}()

		select {
		case <-groupDone:
			log.Debug("Group %s finished running checks.", group.name)
		case abortReason = <-abortChan:
			log.Warn("Group %s aborted.", group.name)
			stopChan <- true

			abort = true
			_ = group.OnAbort(abortReason)
		case <-timeOutChan:
			log.Warn("Running all checks timed-out.")
			stopChan <- true

			abort = true
			abortReason = "global time-out"
			_ = group.OnAbort(abortReason)
		case <-sigIntChan:
			log.Warn("SIGINT/SIGTERM received.")
			stopChan <- true

			abort = true
			abortReason = "SIGINT/SIGTERM"
			_ = group.OnAbort(abortReason)
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

// recordCheckResult records the outcome of a check and updates internal state accordingly.
//
// It logs the result using LogInfo or LogFatal depending on success, then stores the check in the results database.
// If the check has been skipped it records that status as SKIPPED. The function also updates timing information for the check.
func recordCheckResult(check *Check) {
	claimID, ok := identifiers.TestIDToClaimID[check.ID]
	if !ok {
		check.LogFatal("TestID %s has no corresponding Claim ID", check.ID)
	}

	check.LogInfo("Recording result %q, claimID: %+v", strings.ToUpper(check.Result.String()), claimID)
	resultsDB[check.ID] = claim.Result{
		TestID:             &claimID,
		State:              check.Result.String(),
		StartTime:          check.StartTime.String(),
		EndTime:            check.EndTime.String(),
		Duration:           int(check.EndTime.Sub(check.StartTime).Seconds()),
		SkipReason:         check.skipReason,
		CapturedTestOutput: check.GetLogs(),
		CheckDetails:       check.details,

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

// GetReconciledResults aggregates a Claim's results into a map keyed by check name.
//
// It returns a map where each key is the identifier of a check and the value
// is the corresponding claim.Result, reconciling the data structure used by
// certsuite-claim with the internal representation. The function creates a new
// map using make and populates it with the reconciled results from the
// underlying database. This enables callers to work with a uniform result set
// without dealing with the generic map[string]interface{} format used by the
// external client.
func GetReconciledResults() map[string]claim.Result {
	resultMap := make(map[string]claim.Result)
	for key := range resultsDB {
		// initializes the result map, if necessary
		if _, ok := resultMap[key]; !ok {
			resultMap[key] = claim.Result{}
		}

		resultMap[key] = resultsDB[key]
	}
	return resultMap
}

const (
	PASSED  = 0
	FAILED  = 1
	SKIPPED = 2
)

// getResultsSummary returns a map of group names to integer slices summarizing check results.
//
// It iterates over all registered check groups, collects the count of checks
// that passed, failed, were skipped, or aborted within each group, and stores
// those counts in a slice indexed by result type. The returned map keys are
// group identifiers; the values are slices where each position corresponds to
// a specific result category (passed, failed, skipped, aborted). This data is
// used for generating summary statistics across all checks.
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

// printFailedChecksLog generates and returns a function that prints the log entries for all failed checks in the database.
//
// It collects failure logs from each check group, formats them with headers
// indicating the number of failures, and writes the combined output to stdout.
// The returned closure performs the printing when invoked.
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

// GetResults retrieves the current mapping of check identifiers to their results.
//
// It returns a copy of the internal map that associates each check name with its claim.Result.
// The returned map is safe for read-only use; modifications do not affect the underlying database.
func GetResults() map[string]claim.Result {
	return resultsDB
}

// GetTestSuites returns the list of test suite names.
//
// It iterates over all registered checks groups and collects
// the unique suite names found in those groups.
// The returned slice contains each suite name once, sorted
// according to the order they are discovered.
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

// GetTotalTests returns the number of test cases stored in the checks database.
//
// It acquires no locks and simply returns the length of the internal map that holds
// all check groups, effectively counting every test case registered in the system.
func GetTotalTests() int {
	return len(resultsDB)
}

// GetTestsCountByState returns the number of tests that are currently in the given state.
//
// GetTestsCountByState returns the number of tests that are currently in the given state.
//
// The function accepts a string representing a test state (for example "PASSED", "FAILED",
// or "SKIPPED") and returns an integer count of all tests recorded in the database
// that match this state. It does not modify any data; it simply queries the current
// results stored in the checksdb package. If the provided state is unrecognized,
// the function will return zero.
func GetTestsCountByState(state string) int {
	count := 0
	for r := range resultsDB {
		if resultsDB[r].State == state {
			count++
		}
	}
	return count
}

// FilterCheckIDs retrieves the list of enabled check identifiers after applying any label expressions that filter checks.
//
// It evaluates the current label expression configuration using the global evaluator,
// then iterates over all registered check groups, collecting the IDs of checks
// that satisfy the expression. The function returns a slice of strings containing
// the matching check IDs and an error if evaluation or group traversal fails.
func FilterCheckIDs() ([]string, error) {
	filteredCheckIDs := []string{}
	for _, group := range dbByGroup {
		for _, check := range group.checks {
			if labelsExprEvaluator.Eval(check.Labels) {
				filteredCheckIDs = append(filteredCheckIDs, check.ID)
			}
		}
	}

	return filteredCheckIDs, nil
}

// InitLabelsExprEvaluator initializes the global labels expression evaluator with a given namespace string.
//
// It constructs a new LabelsExprEvaluator using the provided namespace,
// assigns it to the package-level variable, and returns any error
// that occurs during creation. If an error is returned, the global
// evaluator remains nil. The function expects a non-empty namespace
// string; passing an empty value will result in an error from the
// underlying NewLabelsExprEvaluator constructor.
func InitLabelsExprEvaluator(labelsFilter string) error {
	// Expand the abstract "all" label into actual existing labels
	if labelsFilter == "all" {
		allTags := []string{identifiers.TagCommon, identifiers.TagExtended,
			identifiers.TagFarEdge, identifiers.TagTelco}
		labelsFilter = strings.Join(allTags, ",")
	}

	eval, err := labels.NewLabelsExprEvaluator(labelsFilter)
	if err != nil {
		return fmt.Errorf("could not create a label evaluator, err: %v", err)
	}

	labelsExprEvaluator = eval

	return nil
}
