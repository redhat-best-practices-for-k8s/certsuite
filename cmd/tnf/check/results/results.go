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

type TestCaseList struct {
	Pass []string `yaml:"pass"`
	Fail []string `yaml:"fail"`
	Skip []string `yaml:"skip"`
}

type TestResults struct {
	TestCaseList `yaml:"testCases"`
}

var checkResultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Verifies that the actual CNFCERT results match the ones found in a reference template",
	RunE:  checkResults,
}

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

func printTestResultsMismatch(mismatchedTestCases []string, actualResults, expectedResults map[string]string) {
	fmt.Printf("\n")
	fmt.Println(strings.Repeat("-", 96)) //nolint:gomnd // table line
	fmt.Printf("| %-58s %-19s %s |\n", "TEST_CASE", "EXPECTED_RESULT", "ACTUAL_RESULT")
	fmt.Println(strings.Repeat("-", 96)) //nolint:gomnd // table line
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
		fmt.Println(strings.Repeat("-", 96)) //nolint:gomnd // table line
	}
}

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

func NewCommand() *cobra.Command {
	checkResultsCmd.PersistentFlags().String("template", "expected_results.yaml", "reference YAML template with the expected results")
	checkResultsCmd.PersistentFlags().String("log-file", "cnf-certification-test/cnf-certsuite.log", "log file of the CNFCERT execution")
	checkResultsCmd.PersistentFlags().Bool("generate-template", false, "generate a reference YAML template from the log file")

	checkResultsCmd.MarkFlagsMutuallyExclusive("template", "generate-template")

	return checkResultsCmd
}
