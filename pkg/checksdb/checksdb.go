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

// RunChecks Executes all check groups with timeout and signal handling
//
// The function locks the database, starts a timeout timer, and listens for
// SIGINT or SIGTERM signals. It iterates over each check group, launching a
// goroutine to run its checks while monitoring for aborts or timeouts. After
// execution it records results, prints a summary table, logs failures, and
// returns the count of failed checks or an error if any occurred.
//
//nolint:funlen
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

// recordCheckResult Stores the check result in the results database
//
// The function looks up a claim ID for a given test, logs debugging information
// if none is found, and otherwise records various fields such as state,
// timestamps, duration, skip reason, captured output, details, category
// classification, and catalog metadata into the global resultsDB map. It
// formats strings to uppercase for logging and calculates duration in seconds
// from start and end times.
func recordCheckResult(check *Check) {
	claimID, ok := identifiers.TestIDToClaimID[check.ID]
	if !ok {
		check.LogDebug("TestID %s has no corresponding Claim ID - skipping result recording", check.ID)
		return
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

// GetReconciledResults Aggregates all stored check results into a map
//
// The function collects entries from an internal database of test outcomes,
// mapping each key to its corresponding claim result object. It ensures every
// key is represented in the returned map, initializing missing entries before
// assigning the actual data. The resulting map is used by other components to
// populate the final claim report.
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

// getResultsSummary generates a table of check results per group
//
// This function builds a map where each key is the name of a check group and
// the value is a slice of three integers counting passed, failed, and skipped
// checks. It iterates over all groups in the database, tallies results for each
// check according to its status, and stores the counts. The resulting map is
// returned for use by the CLI output.
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

// printFailedChecksLog Displays logs for checks that failed
//
// This function iterates over all check groups and their individual checks,
// printing a formatted header and the log content only for those that did not
// succeed. For each failed check it calculates the appropriate number of dashes
// to align the header, prints separators, the colored header indicating the
// check ID, and then either the captured log or a message if no output was
// recorded. The function writes directly to standard output using fmt.Println.
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

// GetResults Retrieves the current mapping of check identifiers to their results
//
// The function returns a map where each key is a string identifier for a
// specific compliance check, and the corresponding value contains the result
// data for that check. It simply exposes an internal database that holds all
// recorded outcomes. No parameters are required or modified during its
// execution.
func GetResults() map[string]claim.Result {
	return resultsDB
}

// GetTestSuites Retrieves a list of unique test suite identifiers from the database
//
// This function iterates over all keys in an internal results map, collecting
// each distinct test suite name into a slice. It ensures no duplicates by
// checking membership before appending. The resulting slice of strings is
// returned for further processing.
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

// GetTotalTests Retrieves the number of tests stored in the database
//
// This function accesses an internal slice that holds test results and returns
// its length as an integer. It provides a quick way to determine how many tests
// are currently recorded without exposing the underlying data structure. The
// result is returned immediately after calculating the count.
func GetTotalTests() int {
	return len(resultsDB)
}

// GetTestsCountByState Counts tests that match a given state
//
// The function iterates over the global results database, incrementing a
// counter each time an entry’s state equals the provided string. It then
// returns the total number of matching entries as an integer. This is useful
// for summarizing how many tests are in a particular status.
func GetTestsCountByState(state string) int {
	count := 0
	for r := range resultsDB {
		if resultsDB[r].State == state {
			count++
		}
	}
	return count
}

// FilterCheckIDs Retrieves test case identifiers that satisfy the current label filter
//
// The function iterates through all check groups in the database, evaluating
// each check's labels against a global expression evaluator. If a check passes
// the evaluation, its identifier is appended to a result slice. After
// processing all checks, the slice of matching IDs is returned with no error.
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

// InitLabelsExprEvaluator Creates a label evaluator from a filter expression
//
// This function takes a string representing a label filter, expands the special
// keyword "all" into a comma‑separated list of known tags, then constructs a
// LabelsExprEvaluator using the helper in the labels package. If construction
// fails, it returns an error describing the problem; otherwise it stores the
// evaluator in a global variable for later use by other parts of the program.
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
