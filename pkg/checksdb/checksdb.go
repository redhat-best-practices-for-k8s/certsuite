// Copyright (C) 2023-2026 Red Hat, Inc.
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
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/redhat-best-practices-for-k8s/checks"
)

var (
	dbLock    sync.Mutex
	dbByGroup map[string]*ChecksGroup

	resultsDB = map[string]claim.Result{}

	labelsExprEvaluator labels.LabelsExprEvaluator
)

type AbortPanicMsg string

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
			groupErrs, failedCheckCtr := group.RunChecks(stopChan, abortChan)
			failedCtr += failedCheckCtr
			errs = append(errs, groupErrs...)
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
	printDaemonsetSkippedChecks()

	if len(errs) > 0 {
		log.Error("RunChecks errors: %v", errs)
		return 0, fmt.Errorf("%d errors found in checks/groups", len(errs))
	}

	return failedCtr, nil
}

func recordCheckResult(check *Check) {
	claimID, catClass, catInfo := resolveCheckMetadata(check)
	if claimID == nil {
		return
	}

	check.LogInfo("Recording result %q, claimID: %+v", strings.ToUpper(check.Result.String()), *claimID)
	resultsDB[check.ID] = claim.Result{
		TestID:                 claimID,
		State:                  check.Result.String(),
		StartTime:              check.StartTime.String(),
		EndTime:                check.EndTime.String(),
		Duration:               int(check.EndTime.Sub(check.StartTime).Seconds()),
		SkipReason:             check.skipReason,
		CapturedTestOutput:     check.GetLogs(),
		CheckDetails:           check.details,
		CategoryClassification: catClass,
		CatalogInfo:            catInfo,
	}
}

func resolveCheckMetadata(check *Check) (*claim.Identifier, *claim.CategoryClassification, *claim.CatalogInfo) {
	// Try the checks library first (single source of truth)
	if info, ok := checks.ByName(check.ID); ok {
		id := claim.Identifier{Id: info.Name, Suite: info.Category, Tags: strings.Join(info.Tags, ",")}
		return &id,
			&claim.CategoryClassification{
				Extended: info.CategoryClassification[checks.Extended],
				FarEdge:  info.CategoryClassification[checks.FarEdge],
				NonTelco: info.CategoryClassification[checks.NonTelco],
				Telco:    info.CategoryClassification[checks.Telco],
			},
			&claim.CatalogInfo{
				Description:           info.Description,
				Remediation:           info.Remediation,
				BestPracticeReference: info.BestPracticeReference,
				ExceptionProcess:      info.ExceptionProcess,
			}
	}

	// Fall back to legacy identifiers for checks not yet in the library
	claimID, ok := identifiers.TestIDToClaimID[check.ID]
	if !ok {
		check.LogDebug("TestID %s has no corresponding Claim ID - skipping result recording", check.ID)
		return nil, nil, nil
	}

	desc := identifiers.Catalog[claimID]
	return &claimID,
		&claim.CategoryClassification{
			Extended: desc.CategoryClassification[identifiers.Extended],
			FarEdge:  desc.CategoryClassification[identifiers.FarEdge],
			NonTelco: desc.CategoryClassification[identifiers.NonTelco],
			Telco:    desc.CategoryClassification[identifiers.Telco],
		},
		&claim.CatalogInfo{
			Description:           desc.Description,
			Remediation:           desc.Remediation,
			BestPracticeReference: desc.BestPracticeReference,
			ExceptionProcess:      desc.ExceptionProcess,
		}
}

// GetReconciledResults is a function added to aggregate a Claim's results.  Due to the limitations of
// certsuite-claim's Go Client, results are generalized to map[string]interface{}.
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

func printDaemonsetSkippedChecks() {
	var skippedIDs []string
	for _, group := range dbByGroup {
		for _, check := range group.checks {
			if check.Result == CheckResultSkipped && check.skipReason == testhelper.DaemonsetFailedToSpawnSkipReason {
				skippedIDs = append(skippedIDs, check.ID)
			}
		}
	}

	if len(skippedIDs) == 0 {
		return
	}

	header := "| " + cli.Yellow + "SKIPPED DUE TO PROBE DAEMONSET FAILURE" + cli.Reset + " |"
	nbSymbols := utf8.RuneCountInString(header) - nbColorSymbols
	fmt.Println(strings.Repeat("=", nbSymbols))
	fmt.Println(header)
	fmt.Println(strings.Repeat("=", nbSymbols))
	fmt.Printf("The probe daemonset failed to deploy. %d test(s) were skipped:\n", len(skippedIDs))
	for _, id := range skippedIDs {
		fmt.Printf("  - %s\n", id)
	}
	fmt.Println()
	fmt.Println("To abort on probe failure instead of skipping, use --require-probe")
	fmt.Println(strings.Repeat("=", nbSymbols))
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
	for r := range resultsDB {
		if resultsDB[r].State == state {
			count++
		}
	}
	return count
}

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

func InitLabelsExprEvaluator(labelsFilter string) error {
	// Expand the abstract "all" label into actual existing labels
	if labelsFilter == "all" {
		allTags := []string{checks.TagCommon, checks.TagExtended,
			checks.TagFarEdge, checks.TagTelco}
		labelsFilter = strings.Join(allTags, ",")
	}

	eval, err := labels.NewLabelsExprEvaluator(labelsFilter)
	if err != nil {
		return fmt.Errorf("could not create a label evaluator, err: %v", err)
	}

	labelsExprEvaluator = eval

	return nil
}
