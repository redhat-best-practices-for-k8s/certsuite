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

// showInfo displays information about selected test cases.
//
// It retrieves flags from the provided cobra command, filters test IDs,
// and prints detailed descriptions or a list of tests depending on
// user options. The function returns an error if any step fails.
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

// NewCommand creates and configures the top-level info command.
//
// It returns a *cobra.Command that prints usage information and
// registers persistent flags such as --verbose and --version.
// The command is set up with short help text, a longer long description,
// and marks required flags where appropriate. No arguments are accepted.
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

// printTestCaseInfoBox displays a formatted information box for a test case.
//
// It prints the title, description, and optional fields of the provided TestCaseDescription
// using styled lines and wrapped text. The output is aligned center or left as appropriate,
// and includes padding based on the lineMaxWidth constant. No value is returned.
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

// printTestList displays a list of test names in a formatted table.
//
// It receives a slice of strings, each representing a test name,
// and prints them to standard output with headers, separators, and
// aligned columns. The function does not return any value.
func printTestList(testIDs []string) {
	fmt.Println("------------------------------------------------------------")
	fmt.Println("|                   TEST CASE SELECTION                    |")
	fmt.Println("------------------------------------------------------------")
	for _, testID := range testIDs {
		fmt.Printf("| %-56s |\n", testID)
	}
	fmt.Println("------------------------------------------------------------")
}

// getMatchingTestIDs returns a slice of test identifiers that satisfy the provided label expression.
//
// It loads the internal checks database, initializes a label expression evaluator with the
// given expression string, and then filters the check IDs based on that evaluator.
// The function may return an error if the database cannot be loaded or if the
// expression evaluation fails.
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

// getTestDescriptionsFromTestIDs retrieves test case descriptions for the provided IDs.
//
// It accepts a slice of strings containing test identifiers and returns
// a slice of claim.TestCaseDescription structs that correspond to those IDs.
// The function collects the descriptions by appending them to a new slice,
// preserving the order of the input IDs. If an ID does not match any known
// description, it is silently skipped. The returned slice may be empty if
// none of the provided IDs are found.
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

// Adjust the maximum width of a line in the output to fit the current terminal.
//
// Adjusts the global variable that controls how wide each line can be.
// If the program is running in a terminal, it queries the terminal size and
// sets the line width to the terminal's column count minus padding. If the
// terminal cannot be detected or an error occurs, it leaves the default
// value unchanged. The function does not return any values.
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
