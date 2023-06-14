package qecoverage

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

var (
	// QeCoverageReportCmd is used to generate a QE coverage report.
	QeCoverageReportCmd = &cobra.Command{
		Use:   "qe-coverage-report",
		Short: "Generates the current QE coverage report.",
		RunE:  runGenerateQeCoverageReport,
	}
)

type QeCoverage struct {
	CoverageByTestSuite     map[string]TestSuiteQeCoverage
	TotalCoveragePercentage float32
	TestCasesTotal          int
	TestCasesWithQe         int
}

type TestSuiteQeCoverage struct {
	TestCases       int
	TestCasesWithQe int
	Coverage        float32
}

func GetQeCoverage(catalog map[claim.Identifier]claim.TestCaseDescription) QeCoverage {
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
		}

		// Update this test suite's coverage percentage
		tsQeCoverage.Coverage = 100.0 * (float32(tsQeCoverage.TestCasesWithQe) / float32(tsQeCoverage.TestCases))

		qeCoverageByTestSuite[tsName] = tsQeCoverage
	}

	totalCoverage := float32(0)
	if totalTcs > 0 {
		totalCoverage = 100.0 * (float32(totalTcsWithQe) / float32(totalTcs))
	}

	return QeCoverage{
		CoverageByTestSuite:     qeCoverageByTestSuite,
		TotalCoveragePercentage: totalCoverage,
		TestCasesTotal:          totalTcs,
		TestCasesWithQe:         totalTcsWithQe,
	}
}

func runGenerateQeCoverageReport(_ *cobra.Command, _ []string) error {
	qeCoverage := GetQeCoverage(identifiers.Catalog)

	// Order test suite names so the report is in ascending test suite name order.
	testSuites := []string{}
	for suite := range qeCoverage.CoverageByTestSuite {
		testSuites = append(testSuites, suite)
	}
	sort.Strings(testSuites)

	// Total QE coverage
	fmt.Printf("Total QE Coverage: %.f%%\n\n", qeCoverage.TotalCoveragePercentage)

	// Per test suite QE coverage
	fmt.Printf("%-30s\t%-20s\t%-20s\t%s\n", "Test Suite Name", "QE Coverage", "Total Test Cases", "Not Covered Test Count")
	for _, suite := range testSuites {
		tsCoverage := qeCoverage.CoverageByTestSuite[suite]
		fmt.Printf("%-30s\t%.0f%%\t%30d\t%10d\n", suite, tsCoverage.Coverage, tsCoverage.TestCases, tsCoverage.TestCases-tsCoverage.TestCasesWithQe)
	}

	fmt.Println()
	return nil
}
