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

// TestCaseList Stores the names of test cases categorized by outcome
//
// This structure keeps three slices, each holding strings that represent test
// case identifiers. The Pass slice lists all tests that succeeded, Fail
// contains those that failed, and Skip holds tests that were not executed. It
// is used to report results in a concise format.
type TestCaseList struct {
	Pass []string `yaml:"pass"`
	Fail []string `yaml:"fail"`
	Skip []string `yaml:"skip"`
}

// TestResults Holds a collection of test case results
//
// This structure contains a slice of individual test case outcomes, allowing
// the program to group related results together. The embedded field
// automatically inherits all fields and methods from the underlying test case
// list type, enabling direct access to the collection’s elements. It serves
// as a container for serializing or reporting aggregated test data.
type TestResults struct {
	TestCaseList `yaml:"testCases"`
}

var checkResultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Verifies that the actual CERTSUITE results match the ones found in a reference template",
	RunE:  checkResults,
}

// checkResults compares recorded test outcomes against a reference template
//
// The function reads actual test results from a log file, optionally generates
// a YAML template of those results, or loads expected results from an existing
// template. It then checks each test case for mismatches between actual and
// expected values, reporting any discrepancies in a formatted table and
// terminating the program if differences are found. If all results match, it
// prints a success message.
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

// getTestResultsDB Parses a log file to build a test result map
//
// The function opens the specified log file, reads it line by line, and
// extracts test case names and their recorded results using a regular
// expression. Each matched pair is stored in a map where the key is the test
// case name and the value is its result string. It returns this map along with
// an error if any step fails.
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

// getExpectedTestResults loads expected test outcomes from a YAML template
//
// The function reads a specified file, decodes its YAML content into a
// structured list of test cases classified as pass, skip, or fail, then builds
// a map associating each case with the corresponding result string. It returns
// this map along with any error that occurs during file reading or
// unmarshalling.
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

// printTestResultsMismatch Displays a formatted table of test cases that did not match the expected results
//
// The function receives a list of mismatched test case identifiers along with
// maps of actual and expected outcomes. It prints a header, then iterates over
// each mismatched case, retrieving the corresponding expected and actual
// values—using a placeholder when either is missing—and outputs them in
// aligned columns. Finally, it draws separators to delineate each row for
// readability.
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

// generateTemplateFile Creates a YAML template file summarizing test case outcomes
//
// This function takes a map of test cases to result strings and builds a
// structured template containing lists for passed, skipped, and failed tests.
// It encodes the structure into YAML with two-space indentation and writes it
// to a predefined file path with specific permissions. If an unknown result
// value is encountered or any I/O operation fails, it returns an error
// detailing the issue.
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

// NewCommand Creates a command for checking test results against expected templates
//
// It defines persistent flags for specifying the template file, log file, and
// an option to generate a new template from logs. The flags are mutually
// exclusive to avoid conflicting inputs. Finally, it returns the configured
// command instance.
func NewCommand() *cobra.Command {
	checkResultsCmd.PersistentFlags().String("template", "expected_results.yaml", "reference YAML template with the expected results")
	checkResultsCmd.PersistentFlags().String("log-file", "certsuite.log", "log file of the Certsuite execution")
	checkResultsCmd.PersistentFlags().Bool("generate-template", false, "generate a reference YAML template from the log file")

	checkResultsCmd.MarkFlagsMutuallyExclusive("template", "generate-template")

	return checkResultsCmd
}
