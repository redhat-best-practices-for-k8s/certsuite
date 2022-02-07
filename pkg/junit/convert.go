// Copyright (C) 2020-2021 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package junit

import (
	"bytes"
	j "encoding/json"
	"fmt"
	"os"

	xj "github.com/basgys/goxml2json"
)

const (
	// junitContentKey is the "#content" key in the JSON serialized JUnit file.
	junitContentKey = "#content"

	// junitMessageKey is the "-message" key in the JSON serialized JUnit file.
	junitMessageKey = "-message"

	// junitFailureKey is the "failure" key in the JSON serialized Junit file.
	junitFailureKey = "failure"

	// junitTestCaseKey is the "testcase" key in the JSON serialized Junit file.
	junitTestCaseKey = "testcase"

	// junitTestNameKey is the "-name" key in the JSON serialized JSON file.
	junitTestNameKey = "-name"

	// junitTestSuiteKey is the "testsuite" key in the JSON serialized JSON file.
	junitTestSuiteKey = "testsuite"

	// junitTestSuitesKey is the "testsuites" key in the JSON serialized JSON file.
	junitTestSuitesKey = "testsuites"

	// CouldNotDeriveFailureReason is the sentinel message emitted when JUnit failure reason cannot be determined.
	CouldNotDeriveFailureReason = "could not derive a reason for the failure from the output JSON"
)

// ExportJUnitAsMap attempts to read a JUnit XML file and converts it to a generic map.
func ExportJUnitAsMap(junitFilename string) (map[string]interface{}, error) {
	xmlReader, err := os.Open(junitFilename)
	// An error is encountered reading the file.
	if err != nil {
		return nil, err
	}

	junitJSONBuffer, err := xj.Convert(xmlReader)
	// An error is encountered translating from XML to JSON.
	if err != nil {
		return nil, err
	}

	jsonMap, err := convertJSONBytesToMap(junitJSONBuffer)
	// An error is encountered unmarshalling the data.
	if err != nil {
		return nil, err
	}

	return jsonMap, err
}

// convertJSONBytesToMap is a utility function to convert a bytes.Buffer to a generic JSON map.
func convertJSONBytesToMap(junitJSONBuffer *bytes.Buffer) (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	err := j.Unmarshal(junitJSONBuffer.Bytes(), &jsonMap)
	return jsonMap, err
}

// TestResult stores whether the test Passed, and an optional FailureReason.
type TestResult struct {
	Passed        bool
	FailureReason string
}

// determineFailureReason attempts to determine a JUnit failure reason.  First, the method attempts to read the failure
// as an object.  Next, the method attempts to extract the "#content" key.  If either of these fails,
// CouldNotDeriveFailureReason is returned to the caller.
func determineFailureReason(failure interface{}) string {
	var failureReason string
	if failureReasonObject, ok := failure.(map[string]interface{}); ok {
		if derivedFailureReason, ok := failureReasonObject[junitContentKey]; ok {
			failureReason = "Failed due to line: " + derivedFailureReason.(string)
		} else {
			failureReason = "Failed due to line: No error line found in JUnit" //nolint:goconst // only instance
		}
		if derivedFailureReason, ok := failureReasonObject[junitMessageKey]; ok {
			failureReason = failureReason + "\n" + "Error message: " + derivedFailureReason.(string)
		} else {
			failureReason = failureReason + "\n" + "Error message: No JUinit message found"
		}
	}
	return failureReason
}

// addResultToMap adds an individual testcase to the aggregated resultsMap.
func addResultToMap(resultsMap map[string]TestResult, resultKey string, passed bool, failureReason string) {
	resultsMap[resultKey] = TestResult{
		Passed:        passed,
		FailureReason: failureReason,
	}
}

// parseResult parses a testcase result and adds it to the resultsMap.
func parseResult(individualTestResultMap map[string]interface{}, resultsMap map[string]TestResult) {
	if key, ok := individualTestResultMap[junitTestNameKey]; ok {
		resultKey := key.(string)
		if failure, ok := individualTestResultMap[junitFailureKey]; ok {
			addResultToMap(resultsMap, resultKey, false, determineFailureReason(failure))
		} else {
			addResultToMap(resultsMap, resultKey, true, "")
		}
	}
}

// parseResult parses all junit testcase instances and adds the results to resultsMap.
func parseResults(junitResultsObjects []interface{}) map[string]TestResult {
	resultsMap := make(map[string]TestResult)
	for _, resultObject := range junitResultsObjects {
		if individualTestResultMap, ok := resultObject.(map[string]interface{}); ok {
			parseResult(individualTestResultMap, resultsMap)
		}
	}
	return resultsMap
}

// toInterfaceArray is a convenience method.  When JUnit contains a single testcase, the results are not stored in an
// array.  To utilize a consistent parseResults() call, the object is embedded into an array.
func toInterfaceArray(object interface{}) []interface{} {
	return []interface{}{object}
}

// ExtractTestSuiteResults takes the JUnit results serialized as JSON, and parses out pass/fail for each JUnit
// "testcase" result.  Note:  This is needed as ginkgo does not assign failure state until after exiting the function
// provided via Ginkgo.It(string, func()).  This allows post-RunSpecs(...) result correlation.
func ExtractTestSuiteResults(junitMap map[string]interface{}, reportKeyName string) (map[string]TestResult, error) {
	// Note:  All of the follow checks are paranoia;  assuming a well formed JUnit output file, most of these checks
	// will never fail.  As such, individual error reporting per case is ignored in favor of a blanket error statement.
	if suites, ok := junitMap[reportKeyName].(map[string]interface{}); ok {
		if testSuitesResults, ok := suites[junitTestSuitesKey]; ok {
			if testSuitesResultsMap, ok := testSuitesResults.(map[string]interface{}); ok {
				if testSuiteResults, ok := testSuitesResultsMap[junitTestSuiteKey]; ok {
					if testCaseResultsMap, ok := testSuiteResults.(map[string]interface{}); ok {
						if testCaseResults, ok := testCaseResultsMap[junitTestCaseKey]; ok {
							// Note:  order is important since an []interface{} can still cast as interface{}, but an
							// interface{} cannot be cast as []interface{}
							if resultsObjects, ok := testCaseResults.([]interface{}); ok {
								resultsMap := parseResults(resultsObjects)
								parseResults(resultsObjects)
								return resultsMap, nil
							}
							resultsObjects := toInterfaceArray(testCaseResults)
							resultsMap := parseResults(resultsObjects)
							parseResults(resultsObjects)
							return resultsMap, nil
						}
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("couldn't parse the JUnit for results")
}
