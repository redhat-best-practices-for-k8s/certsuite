package qecoverage

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

const (
	multiplier = 100.0
)

type TestCoverageSummaryReport struct {
	CoverageByTestSuite     map[string]TestSuiteQeCoverage
	TotalCoveragePercentage float32
	TestCasesTotal          int
	TestCasesWithQe         int
}

type TestSuiteQeCoverage struct {
	TestCases               int
	TestCasesWithQe         int
	Coverage                float32
	NotImplementedTestCases []string
}

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
