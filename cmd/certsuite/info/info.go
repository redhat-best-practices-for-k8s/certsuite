package info

import (
	"fmt"
	"os"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/cli"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/certsuite"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const linePadding = 4

var (
	infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Displays information from the test catalog",
		RunE:  showInfo,
	}
	lineMaxWidth = 120
)

func showInfo(cmd *cobra.Command, _ []string) error {
	testCaseFlag, _ := cmd.Flags().GetString("test-label")
	listFlag, _ := cmd.Flags().GetBool("list")

	// Get a list of matching test cases names
	testIDs, err := getMatchingTestIDs(testCaseFlag)
	if err != nil {
		return fmt.Errorf("could not get the matching test case list, err: %v", err)
	}

	// Print the list and leave if only listing is required
	if listFlag {
		printTestList(testIDs)
		return nil
	}

	// Get a list of test descriptions with detail info per test case
	testCases := getTestDescriptionsFromTestIDs(testIDs)
	if len(testCases) == 0 {
		return fmt.Errorf("no test case found matching name %q", testCaseFlag)
	}

	// Adjust text box line width
	adjustLineMaxWidth()

	// Print test case info box
	for i := range testCases {
		printTestCaseInfoBox(&testCases[i])
	}

	return nil
}

func NewCommand() *cobra.Command {
	infoCmd.PersistentFlags().StringP("test-label", "t", "", "The test label filter to select the test cases to show information about")
	infoCmd.PersistentFlags().BoolP("list", "l", false, "Show only the names of the test cases for a given test label")
	err := infoCmd.MarkPersistentFlagRequired("test-label")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not mark persistent flag \"test-case\" as required, err: %v", err)
		return nil
	}
	return infoCmd
}

func printTestCaseInfoBox(testCase *claim.TestCaseDescription) {
	// Test case identifier
	border := strings.Repeat("-", lineMaxWidth+linePadding)
	fmt.Println(border)
	fmt.Printf("| %s |\n", cli.LineColor(cli.LineAlignCenter(testCase.Identifier.Id, lineMaxWidth), cli.Cyan))

	// Description
	border = strings.Repeat("-", lineMaxWidth+linePadding)
	fmt.Println(border)
	fmt.Printf("| %s |\n", cli.LineColor(cli.LineAlignCenter("DESCRIPTION", lineMaxWidth), cli.Green))
	fmt.Println(border)
	for _, line := range cli.WrapLines(testCase.Description, lineMaxWidth) {
		fmt.Printf("| %s |\n", cli.LineAlignLeft(line, lineMaxWidth))
	}

	// Remediation
	fmt.Println(border)
	fmt.Printf("| %s |\n", cli.LineColor(cli.LineAlignCenter("REMEDIATION", lineMaxWidth), cli.Green))
	fmt.Println(border)
	for _, line := range cli.WrapLines(testCase.Remediation, lineMaxWidth) {
		fmt.Printf("| %s |\n", cli.LineAlignLeft(line, lineMaxWidth))
	}

	// Exceptions
	fmt.Println(border)
	fmt.Printf("| %s |\n", cli.LineColor(cli.LineAlignCenter("EXCEPTIONS", lineMaxWidth), cli.Green))
	fmt.Println(border)
	for _, line := range cli.WrapLines(testCase.ExceptionProcess, lineMaxWidth) {
		fmt.Printf("| %s |\n", cli.LineAlignLeft(line, lineMaxWidth))
	}

	// Best Practices reference
	fmt.Println(border)
	fmt.Printf("| %s |\n", cli.LineColor(cli.LineAlignCenter("BEST PRACTICES REFERENCE", lineMaxWidth), cli.Green))
	fmt.Println(border)
	for _, line := range cli.WrapLines(testCase.BestPracticeReference, lineMaxWidth) {
		fmt.Printf("| %s |\n", cli.LineAlignLeft(line, lineMaxWidth))
	}
	fmt.Println(border)
	fmt.Printf("\n\n")
}

func printTestList(testIDs []string) {
	fmt.Println("------------------------------------------------------------")
	fmt.Println("|                   TEST CASE SELECTION                    |")
	fmt.Println("------------------------------------------------------------")
	for _, testID := range testIDs {
		fmt.Printf("| %-56s |\n", testID)
	}
	fmt.Println("------------------------------------------------------------")
}

func getMatchingTestIDs(labelExpr string) ([]string, error) {
	if err := checksdb.InitLabelsExprEvaluator(labelExpr); err != nil {
		return nil, fmt.Errorf("failed to initialize a test case label evaluator, err: %v", err)
	}
	certsuite.LoadInternalChecksDB()
	testIDs, err := checksdb.FilterCheckIDs()
	if err != nil {
		return nil, fmt.Errorf("could not list test cases, err: %v", err)
	}

	return testIDs, nil
}

func getTestDescriptionsFromTestIDs(testIDs []string) []claim.TestCaseDescription {
	var testCases []claim.TestCaseDescription
	for _, test := range testIDs {
		for id := range identifiers.Catalog {
			if id.Id == test {
				testCases = append(testCases, identifiers.Catalog[id])
				break
			}
		}
	}
	return testCases
}

func adjustLineMaxWidth() {
	if term.IsTerminal(0) {
		width, _, err := term.GetSize(0)
		if err != nil {
			return
		}
		if width < lineMaxWidth+linePadding {
			lineMaxWidth = width - linePadding
		}
	}
}
