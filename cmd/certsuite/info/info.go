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

// showInfo Displays detailed information about selected test cases
//
// The function retrieves a list of test case identifiers based on a label
// expression, optionally listing them if the --list flag is set. If not
// listing, it fetches full descriptions for each matching test case and prints
// a formatted box containing identifier, description, remediation, exceptions,
// and best practice references. Errors are returned if no matches or retrieval
// fails.
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

// NewCommand Creates the info subcommand with a required test-label flag
//
// The function configures an information command for the CLI by adding
// persistent string and boolean flags that filter and display test case data.
// It marks the test-label flag as mandatory, printing an error to standard
// error if this fails, and then returns the configured command object.
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

// printTestCaseInfoBox Displays a formatted information box for a test case
//
// The function builds a bordered text block that shows the test case ID,
// description, remediation steps, exceptions, and best‑practice references.
// It uses helper functions to center or left‑align lines, color headers, and
// wrap long paragraphs to fit within the terminal width. Each section is
// separated by horizontal borders made of dashes.
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

// printTestList Displays a formatted list of test case identifiers
//
// The function receives a slice of strings representing test IDs, then prints a
// header, each ID within a bordered box, and a footer to visually separate the
// list. It uses fixed-width formatting so that all entries align consistently
// in the terminal output. No value is returned; the output is directed to
// standard output via fmt functions.
func printTestList(testIDs []string) {
	fmt.Println("------------------------------------------------------------")
	fmt.Println("|                   TEST CASE SELECTION                    |")
	fmt.Println("------------------------------------------------------------")
	for _, testID := range testIDs {
		fmt.Printf("| %-56s |\n", testID)
	}
	fmt.Println("------------------------------------------------------------")
}

// getMatchingTestIDs retrieves test case identifiers that match a label expression
//
// The function initializes a label evaluator with the provided expression,
// loads all internal check definitions, then filters those checks to return
// only IDs whose labels satisfy the evaluator. It returns a slice of matching
// IDs or an error if initialization or filtering fails.
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

// getTestDescriptionsFromTestIDs Retrieves test case descriptions for given IDs
//
// The function receives a slice of test ID strings, iterates over each ID, and
// searches a catalog map for matching entries by comparing the identifier
// field. When a match is found, the corresponding test case description is
// appended to a result slice. After processing all input IDs, it returns the
// slice containing all matched descriptions, which may be empty if no IDs were
// found.
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

// adjustLineMaxWidth Adjusts the maximum line width for output
//
// The function checks if standard input is a terminal, then retrieves the
// terminal's width. If the width is smaller than the current maximum plus
// padding, it reduces the maximum line width accordingly to fit the display. No
// value is returned; the global variable is updated in place.
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
