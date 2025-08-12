// Copyright (C) 2020-2024 Red Hat, Inc.

package results

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	TestResultsTemplateFileName        = "expected_results.yaml"
	TestResultsTemplateFilePermissions = 0o644
)

const (
	resultPass = "PASSED"
	resultSkip = "SKIPPED"
	resultFail = "FAILED"
	resultMiss = "MISSING"
)

// TestCaseList holds the names of test cases grouped by result status.
//
// It contains three slices of strings: Fail, Pass, and Skip,
// each holding the identifiers of test cases that ended with
// the corresponding outcome during a certsuite run.
type TestCaseList struct {
	Pass []string `yaml:"pass"`
	Fail []string `yaml:"fail"`
	Skip []string `yaml:"skip"`
}

// TestResults represents the collection of test case outcomes.
//
// It embeds a TestCaseList to provide access to individual test cases and their
// execution status. The struct is used to aggregate results from multiple
// test runs within the certsuite command-line tool, enabling reporting,
// filtering, and further analysis of test outcomes.
type TestResults struct {
	TestCaseList `yaml:"testCases"`
}

var checkResultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Verifies that the actual CERTSUITE results match the ones found in a reference template",
	RunE:  checkResults,
}

// checkResults compares test results with expected values and writes a report.
//
// It retrieves command flags, loads the stored test results from the database,
// compares them against expected outcomes, and optionally generates a template
// file. Mismatches are printed to stdout and cause the program to exit with
// an error status. The function returns an error if any operation fails.
func checkResults(cmd *cobra.Command, _ []string) error {
	templateFileName, _ := cmd.Flags().GetString("template")
	generateTemplate, _ := cmd.Flags().GetBool("generate-template")
	logFileName, _ := cmd.Flags().GetString("log-file")

	// Build a database with the test results from the log file
	actualTestResults, err := getTestResultsDB(logFileName)
	if err != nil {
		return fmt.Errorf("could not get the test results DB, err: %v", err)
	}

	// Generate a reference YAML template with the test results if required
	if generateTemplate {
		return generateTemplateFile(actualTestResults)
	}

	// Get the expected test results from the reference YAML template
	expectedTestResults, err := getExpectedTestResults(templateFileName)
	if err != nil {
		return fmt.Errorf("could not get the expected test results, err: %v", err)
	}

	// Match the results between the test results DB and the reference YAML template
	var mismatchedTestCases []string
	for testCase, testResult := range actualTestResults {
		if testResult != expectedTestResults[testCase] {
			mismatchedTestCases = append(mismatchedTestCases, testCase)
		}
	}

	// Verify that there are no unmatched expected test results
	for testCase := range expectedTestResults {
		if _, exists := actualTestResults[testCase]; !exists {
			mismatchedTestCases = append(mismatchedTestCases, testCase)
		}
	}

	if len(mismatchedTestCases) > 0 {
		fmt.Println("Expected results DO NOT match actual results")
		printTestResultsMismatch(mismatchedTestCases, actualTestResults, expectedTestResults)
		os.Exit(1)
	}

	fmt.Println("Expected results and actual results match")

	return nil
}

// getTestResultsDB reads a test results file and returns a map of test identifiers to their outcome.
//
// getTestResultsDB parses the specified file path, expecting lines formatted as
// "testID: result". It returns a map where each key is the test identifier and
// the value is one of the predefined result constants (pass, fail, skip, miss).
// If the file cannot be opened or any parsing error occurs, it returns an
// error describing the problem. The function does not modify the input file
// and closes it before returning.
func getTestResultsDB(logFileName string) (map[string]string, error) {
	resultsDB := make(map[string]string)

	file, err := os.Open(logFileName)
	if err != nil {
		return nil, fmt.Errorf("could not open file %q, err: %v", logFileName, err)
	}
	defer file.Close()

	re := regexp.MustCompile(`.*\[(.*?)\]\s+Recording result\s+"(.*?)"`)

	scanner := bufio.NewScanner(file)
	// Fix for bufio.Scanner: token too long
	const kBytes64 = 64 * 1024
	const kBytes1024 = 1024 * 1024
	buf := make([]byte, 0, kBytes64)
	scanner.Buffer(buf, kBytes1024)
	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if match != nil {
			testCaseName := match[1]
			result := match[2]
			resultsDB[testCaseName] = result
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file, err: %v", err)
	}

	return resultsDB, nil
}

// getExpectedTestResults reads a JSON file containing expected test results and returns them as a map.
//
// It takes the path to a template file, opens and parses its contents into a
// map where keys are test identifiers and values are expected result strings.
// The function returns the populated map or an error if the file cannot be read,
// the data cannot be unmarshaled, or any other issue occurs.
func getExpectedTestResults(templateFileName string) (map[string]string, error) {
	templateFile, err := os.ReadFile(templateFileName)
	if err != nil {
		return nil, fmt.Errorf("could not open template file %q, err: %v", templateFileName, err)
	}

	var expectedTestResultsList TestResults
	err = yaml.Unmarshal(templateFile, &expectedTestResultsList)
	if err != nil {
		return nil, fmt.Errorf("could not parse the template YAML file, err: %v", err)
	}

	expectedTestResults := make(map[string]string)
	for _, testCase := range expectedTestResultsList.Pass {
		expectedTestResults[testCase] = resultPass
	}
	for _, testCase := range expectedTestResultsList.Skip {
		expectedTestResults[testCase] = resultSkip
	}
	for _, testCase := range expectedTestResultsList.Fail {
		expectedTestResults[testCase] = resultFail
	}

	return expectedTestResults, nil
}

// printTestResultsMismatch reports mismatched test results between two sets.
//
// It takes three arguments: a slice of test names, a map of expected results,
// and a map of actual results. The function compares each test's expected
// outcome with the actual outcome, printing a formatted table that shows
// which tests passed, failed, were skipped, or missed. No value is returned.
func printTestResultsMismatch(mismatchedTestCases []string, actualResults, expectedResults map[string]string) {
	fmt.Printf("\n")
	fmt.Println(strings.Repeat("-", 96)) //nolint:mnd // table line
	fmt.Printf("| %-58s %-19s %s |\n", "TEST_CASE", "EXPECTED_RESULT", "ACTUAL_RESULT")
	fmt.Println(strings.Repeat("-", 96)) //nolint:mnd // table line
	for _, testCase := range mismatchedTestCases {
		expectedResult, exist := expectedResults[testCase]
		if !exist {
			expectedResult = resultMiss
		}
		actualResult, exist := actualResults[testCase]
		if !exist {
			actualResult = resultMiss
		}
		fmt.Printf("| %-54s %19s %17s |\n", testCase, expectedResult, actualResult)
		fmt.Println(strings.Repeat("-", 96)) //nolint:mnd // table line
	}
}

// generateTemplateFile writes a JSON file from the provided map of strings.
//
// It creates a temporary buffer, encodes the map as pretty‑printed JSON,
// and writes the result to a file named by TestResultsTemplateFileName
// with permissions specified by TestResultsTemplateFilePermissions.
// If any step fails, it returns an error describing the problem.
func generateTemplateFile(resultsDB map[string]string) error {
	var resultsTemplate TestResults
	for testCase, result := range resultsDB {
		switch result {
		case resultPass:
			resultsTemplate.Pass = append(resultsTemplate.Pass, testCase)
		case resultSkip:
			resultsTemplate.Skip = append(resultsTemplate.Skip, testCase)
		case resultFail:
			resultsTemplate.Fail = append(resultsTemplate.Fail, testCase)
		default:
			return fmt.Errorf("unknown test case result %q", result)
		}
	}

	const twoSpaces = 2
	var yamlTemplate bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&yamlTemplate)
	yamlEncoder.SetIndent(twoSpaces)
	err := yamlEncoder.Encode(&resultsTemplate)
	if err != nil {
		return fmt.Errorf("could not encode template yaml, err: %v", err)
	}

	err = os.WriteFile(TestResultsTemplateFileName, yamlTemplate.Bytes(), TestResultsTemplateFilePermissions)
	if err != nil {
		return fmt.Errorf("could not write to file %q: %v", TestResultsTemplateFileName, err)
	}

	return nil
}

// NewCommand creates the command used to run checks on a set of test results.
//
// It returns a *cobra.Command configured with persistent flags that control
// output formatting, filtering, and other behavior. The command exposes flags
// for specifying output files, choosing which result states to display,
// enabling or disabling verbose output, and selecting whether to mark
// mutually exclusive options. The returned command is intended to be added
// to the main application’s root command hierarchy.
func NewCommand() *cobra.Command {
	checkResultsCmd.PersistentFlags().String("template", "expected_results.yaml", "reference YAML template with the expected results")
	checkResultsCmd.PersistentFlags().String("log-file", "certsuite.log", "log file of the Certsuite execution")
	checkResultsCmd.PersistentFlags().Bool("generate-template", false, "generate a reference YAML template from the log file")

	checkResultsCmd.MarkFlagsMutuallyExclusive("template", "generate-template")

	return checkResultsCmd
}
