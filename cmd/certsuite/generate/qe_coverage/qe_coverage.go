package qecoverage

import (
	"fmt"
	"sort"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/spf13/cobra"
)

const (
	multiplier = 100.0
)

// TestCoverageSummaryReport Provides a snapshot of QE coverage across test suites
//
// This struct holds overall statistics such as total test cases, those covered
// by QE, and the percentage of coverage. It also maps each suite name to its
// own TestSuiteQeCoverage record for detailed per-suite information. The data
// is used by reporting functions to display coverage metrics.
type TestCoverageSummaryReport struct {
	CoverageByTestSuite     map[string]TestSuiteQeCoverage
	TotalCoveragePercentage float32
	TestCasesTotal          int
	TestCasesWithQe         int
}

// TestSuiteQeCoverage Represents coverage statistics for a test suite
//
// This structure holds counts of total test cases, how many include QE-specific
// tests, and the calculated percentage coverage. It also tracks any test cases
// that are not yet implemented. The data can be used to assess overall quality
// and identify gaps in QE integration.
type TestSuiteQeCoverage struct {
	TestCases               int
	TestCasesWithQe         int
	Coverage                float32
	NotImplementedTestCases []string
}

// NewCommand Creates a command to report QE test coverage
//
// The function builds a new command instance that includes a persistent string
// flag named "suitename" for filtering coverage output by suite name. It
// returns this configured command so it can be added to the parent generate
// command hierarchy.
func NewCommand() *cobra.Command {
	qeCoverageReportCmd.PersistentFlags().String("suitename", "", "Displays the remaining tests not covered by QE for the specified suite name.")

	return qeCoverageReportCmd
}

var (
	// QeCoverageReportCmd is used to generate a QE coverage report.
	qeCoverageReportCmd = &cobra.Command{
		Use:   "qe-coverage-report",
		Short: "Generates the current QE coverage report.",
		Run: func(cmd *cobra.Command, args []string) {
			testSuiteName, _ := cmd.Flags().GetString("suitename")

			qeCoverage := GetQeCoverage(identifiers.Catalog)

			if testSuiteName != "" {
				_, exists := qeCoverage.CoverageByTestSuite[testSuiteName]
				if exists {
					showQeCoverageForTestCaseName(testSuiteName, qeCoverage)
				} else {
					fmt.Println("Invalid test suite name")
				}
			} else {
				showQeCoverageSummaryReport()
			}
		},
	}
)

// showQeCoverageForTestCaseName Displays QE coverage statistics for a specified test suite
//
// The function prints the name of the test suite, total number of test cases,
// overall coverage percentage, and how many are not covered by QE. It then
// reports whether all tests have QE coverage or lists any unimplemented test
// cases in detail.
func showQeCoverageForTestCaseName(suiteName string, qeCoverage TestCoverageSummaryReport) {
	tsCoverage := qeCoverage.CoverageByTestSuite[suiteName]

	fmt.Println("Suite Name : ", suiteName)
	fmt.Printf("Total Test Cases : %d, QE Coverage:  %.f%%, Unimplemented Test Cases : %d\n",
		tsCoverage.TestCases, tsCoverage.Coverage, tsCoverage.TestCases-tsCoverage.TestCasesWithQe)

	if len(tsCoverage.NotImplementedTestCases) == 0 {
		fmt.Println("Congrats! All tests are QE test covered")
	} else {
		var testCases = strings.Join(tsCoverage.NotImplementedTestCases, "\n")
		fmt.Printf("\nUnimplemented Test Cases are the following: \n\n%s", testCases)
	}

	fmt.Println()
}

// GetQeCoverage Calculates overall and per-suite QE coverage statistics
//
// The function iterates over a catalog of test case descriptions, counting
// total cases, those marked for QE, and noting which are not implemented. It
// aggregates these counts by test suite, computing a percentage coverage for
// each suite using a multiplier factor. Finally, it returns a summary report
// containing per-suite data, overall coverage, and total counts.
func GetQeCoverage(catalog map[claim.Identifier]claim.TestCaseDescription) TestCoverageSummaryReport {
	totalTcs := 0
	totalTcsWithQe := 0

	qeCoverageByTestSuite := map[string]TestSuiteQeCoverage{}

	for claimID := range catalog {
		totalTcs++

		tcDescription := catalog[claimID]

		tsName := tcDescription.Identifier.Suite

		tsQeCoverage, exists := qeCoverageByTestSuite[tsName]
		if !exists {
			tsQeCoverage = TestSuiteQeCoverage{}
		}

		tsQeCoverage.TestCases++
		if tcDescription.Qe {
			tsQeCoverage.TestCasesWithQe++
			totalTcsWithQe++
		} else {
			tsQeCoverage.NotImplementedTestCases = append(tsQeCoverage.NotImplementedTestCases, tcDescription.Identifier.Id)
		}

		// Update this test suite's coverage percentage
		tsQeCoverage.Coverage = multiplier * (float32(tsQeCoverage.TestCasesWithQe) / float32(tsQeCoverage.TestCases))

		qeCoverageByTestSuite[tsName] = tsQeCoverage
	}

	totalCoverage := float32(0)
	if totalTcs > 0 {
		totalCoverage = multiplier * (float32(totalTcsWithQe) / float32(totalTcs))
	}

	return TestCoverageSummaryReport{
		CoverageByTestSuite:     qeCoverageByTestSuite,
		TotalCoveragePercentage: totalCoverage,
		TestCasesTotal:          totalTcs,
		TestCasesWithQe:         totalTcsWithQe,
	}
}

// showQeCoverageSummaryReport Displays a formatted report of QE coverage statistics
//
// This routine calculates overall and per-test-suite coverage by calling
// GetQeCoverage, then sorts the suite names alphabetically. It prints total
// percentages and counts, followed by a table showing each suite’s name, its
// coverage percentage, total test cases, and how many are not covered. The
// output is formatted for console readability.
func showQeCoverageSummaryReport() {
	qeCoverage := GetQeCoverage(identifiers.Catalog)

	// Order test suite names so the report is in ascending test suite name order.
	testSuites := []string{}
	for suite := range qeCoverage.CoverageByTestSuite {
		testSuites = append(testSuites, suite)
	}
	sort.Strings(testSuites)

	// QE Coverage details
	fmt.Printf("Total QE Coverage: %.f%%\n\n", qeCoverage.TotalCoveragePercentage)
	fmt.Printf("Total Test Cases: %d, Total QE Test Cases: %d\n\n", qeCoverage.TestCasesTotal, qeCoverage.TestCasesWithQe)

	// Per test suite QE coverage
	fmt.Printf("%-30s\t%-20s\t%-20s\t%s\n", "Test Suite Name", "QE Coverage", "Total Test Cases", "Not Covered Test Count")
	for _, suite := range testSuites {
		tsCoverage := qeCoverage.CoverageByTestSuite[suite]
		fmt.Printf("%-30s\t%.0f%%\t%30d\t%10d\n", suite, tsCoverage.Coverage, tsCoverage.TestCases, tsCoverage.TestCases-tsCoverage.TestCasesWithQe)
	}

	fmt.Println()
}
