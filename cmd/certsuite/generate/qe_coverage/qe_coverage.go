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

// TestCoverageSummaryReport represents a summary of quality engineering coverage across test suites.
//
// It contains the total number of test cases, how many have QE data,
// and an overall coverage percentage. The CoverageByTestSuite map holds
// detailed coverage metrics per individual test suite. This struct is used
// by GetQeCoverage to aggregate results from multiple test case descriptions.
type TestCoverageSummaryReport struct {
	CoverageByTestSuite     map[string]TestSuiteQeCoverage
	TotalCoveragePercentage float32
	TestCasesTotal          int
	TestCasesWithQe         int
}

// TestSuiteQeCoverage holds coverage statistics for a test suite.
//
// It records the overall coverage percentage, the total number of test cases,
// how many have QE tests, and a list of test case names that are not yet
// implemented for QE testing. The Coverage field is expressed as a float32
// between 0 and 100. TestCasesWithQe indicates how many test cases include
// QE checks, while NotImplementedTestCases lists the remaining ones.
type TestSuiteQeCoverage struct {
	TestCases               int
	TestCasesWithQe         int
	Coverage                float32
	NotImplementedTestCases []string
}

// NewCommand creates the QE coverage command for certsuite.
//
// It constructs a new cobra.Command with appropriate usage information
// and registers persistent flags required by the QE coverage subcommand.
// The function returns a pointer to this configured command, ready to be added
// to the main command hierarchy.
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

// showQeCoverageForTestCaseName displays the QE coverage summary for a given test case.
//
// It takes a string representing the test case name and a TestCoverageSummaryReport containing
// coverage data. The function prints formatted output to standard out, including the number of
// total tests, passed tests, failed tests, and any other relevant metrics contained in the report.
// If the report has no entries, it simply indicates that no coverage information is available.
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

// GetQeCoverage generates a summary report of test case coverage from QE data.
//
// It accepts a map that associates claim identifiers with their corresponding
// test case descriptions. The function iterates over the provided data,
// calculates various coverage metrics, and returns a TestCoverageSummaryReport
// struct containing aggregated statistics such as total tests, passed tests,
// failed tests, and coverage percentages.
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

// showQeCoverageSummaryReport displays a summary report of QE coverage metrics.
//
// It retrieves the current QE coverage data, formats key statistics such as total tests,
// passed tests, and coverage percentages, then prints the information to standard output.
// The function does not return any values.
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
