// Copyright (C) 2020-2024 Red Hat, Inc.
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

package claimhelper

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/stretchr/testify/assert"
)

func TestPopulateXMLFromClaim(t *testing.T) {
	generateClaim := func(results map[string]claim.Result) claim.Claim {
		c := claim.Claim{}
		c.Results = make(map[string]claim.Result)

		// Set the results if any
		for k, v := range results {
			c.Results[k] = v
		}

		return c
	}

	generateFailureResult := func(testSuiteName, testCaseName, failureMessage string) claim.Result {
		return claim.Result{
			TestID: &claim.Identifier{
				Id:    testCaseName,
				Suite: testSuiteName,
			},
			State:              "failed",
			SkipReason:         failureMessage,
			StartTime:          "2023-12-20 14:51:33 -0600 MST",
			EndTime:            "2023-12-20 14:51:34 -0600 MST",
			CheckDetails:       "",
			CapturedTestOutput: "test output",
			CategoryClassification: &claim.CategoryClassification{
				Extended: "false",
				FarEdge:  "false",
				NonTelco: "false",
				Telco:    "true",
			},
		}
	}

	testCases := []struct {
		testResult        claim.Result
		expectedXMLResult TestSuitesXML
	}{
		{
			testResult: generateFailureResult(
				"test-suite1",
				"test-case1",
				"my custom failure message",
			),
			expectedXMLResult: TestSuitesXML{
				Failures: strconv.Itoa(1),
				Disabled: strconv.Itoa(0),
				Tests:    strconv.Itoa(1),
				Errors:   strconv.Itoa(0),
				Time:     strconv.Itoa(1),
				Testsuite: Testsuite{

					Name:     "test-suite1",
					Failures: strconv.Itoa(1),
					Skipped:  strconv.Itoa(0),
					Tests:    strconv.Itoa(1),
					Errors:   strconv.Itoa(0),
					Time:     strconv.Itoa(60),
					Testcase: []TestCase{
						{
							Name: "test-case1",
							Time: strconv.Itoa(1),
							Failure: &FailureMessage{
								Message: "my custom failure message",
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		// Build some 1 minute duration start and end time
		startTime, err := time.Parse(DateTimeFormatDirective, "2023-12-20 14:51:33 -0600 MST")
		assert.Nil(t, err)
		endTime, err := time.Parse(DateTimeFormatDirective, "2023-12-20 14:51:34 -0600 MST")
		assert.Nil(t, err)

		xmlResult := populateXMLFromClaim(generateClaim(map[string]claim.Result{"test-case1": tc.testResult}), startTime, endTime)

		// Compare the values in the XML
		assert.Equal(t, tc.expectedXMLResult.Failures, xmlResult.Failures)
		assert.Equal(t, tc.expectedXMLResult.Tests, xmlResult.Tests)
		assert.Equal(t, tc.expectedXMLResult.Errors, xmlResult.Errors)
		expectedTimeFloat, err := strconv.ParseFloat(tc.expectedXMLResult.Time, 32)
		assert.Nil(t, err)
		actualTimeFloat, err := strconv.ParseFloat(xmlResult.Time, 32)
		assert.Nil(t, err)
		assert.Equal(t, int(expectedTimeFloat), int(actualTimeFloat))
	}
}

func TestToJUnitXML(t *testing.T) {
	testCases := []struct {
		testResults       map[string]claim.Result
		expectedXMLResult string
	}{
		{
			testResults: map[string]claim.Result{
				"test-case1": {
					TestID: &claim.Identifier{
						Id:    "test-case1",
						Suite: "test-suite1",
					},
					State:              "failed",
					SkipReason:         "my custom failure message",
					StartTime:          "2023-12-20 14:51:33 -0600 MST",
					EndTime:            "2023-12-20 14:51:34 -0600 MST",
					CheckDetails:       "",
					CapturedTestOutput: "test output",
					CategoryClassification: &claim.CategoryClassification{
						Extended: "false",
						FarEdge:  "false",
						NonTelco: "false",
						Telco:    "true",
					},
				},
			},
			expectedXMLResult: "<testsuites tests=\"1\" disabled=\"0\" errors=\"0\" failures=\"1\" time=\"1.00000\">",
		},
		{
			testResults: map[string]claim.Result{
				"test-case1": {
					TestID: &claim.Identifier{
						Id:    "test-case1",
						Suite: "test-suite1",
					},
					State:              "passed",
					SkipReason:         "",
					StartTime:          "2023-12-20 14:51:33 -0600 MST",
					EndTime:            "2023-12-20 14:51:34 -0600 MST",
					CheckDetails:       "",
					CapturedTestOutput: "test output",
					CategoryClassification: &claim.CategoryClassification{
						Extended: "false",
						FarEdge:  "false",
						NonTelco: "false",
						Telco:    "true",
					},
				},
			},
			expectedXMLResult: "<testsuites tests=\"1\" disabled=\"0\" errors=\"0\" failures=\"0\" time=\"1.00000\">",
		},
	}

	t.Setenv("UNIT_TEST", "true")
	defer os.Remove("testfile.xml")
	defer os.Unsetenv("UNIT_TEST")

	for _, tc := range testCases {
		// Build some 1 minute duration start and end time
		startTime, err := time.Parse(DateTimeFormatDirective, "2023-12-20 14:51:33 -0600 MST")
		assert.Nil(t, err)
		endTime, err := time.Parse(DateTimeFormatDirective, "2023-12-20 14:51:34 -0600 MST")
		assert.Nil(t, err)

		testClaimBuilder, err := NewClaimBuilder()
		assert.Nil(t, err)

		testClaimBuilder.claimRoot.Claim.Results = make(map[string]claim.Result)
		testClaimBuilder.claimRoot.Claim.Results = tc.testResults

		testClaimBuilder.ToJUnitXML("testfile.xml", startTime, endTime)

		// read the file and compare the contents
		outputFile, err := os.ReadFile("testfile.xml")
		assert.Nil(t, err)

		xmlResult := string(outputFile)

		// Compare the values in the XML
		assert.Contains(t, xmlResult, tc.expectedXMLResult)
		os.Remove("testfile.xml")
	}
}

func TestMarshalClaimOutput(t *testing.T) {
	testClaimRoot := &claim.Root{
		Claim: &claim.Claim{
			Metadata: &claim.Metadata{
				StartTime: "2023-12-20 14:51:33 -0600 MST",
				EndTime:   "2023-12-20 14:51:34 -0600 MST",
			},
			Versions: &claim.Versions{
				CertSuite: "1.0.0",
			},
			Results: map[string]claim.Result{
				"test-case1": {
					TestID: &claim.Identifier{
						Id:    "test-case1",
						Suite: "test-suite1",
					},
					State: "failed",
				},
			},
		},
	}

	output := MarshalClaimOutput(testClaimRoot)
	assert.NotNil(t, output)

	// Check if the output is a valid JSON
	//nolint:lll
	assert.Contains(t, string(output), "{\n  \"claim\": {\n    \"configurations\": null,\n    \"metadata\": {\n      \"endTime\": \"2023-12-20 14:51:34 -0600 MST\",\n      \"startTime\": \"2023-12-20 14:51:33 -0600 MST\"\n    },\n    \"nodes\": null,\n    \"results\": {\n      \"test-case1\": {\n        \"capturedTestOutput\": \"\",\n        \"catalogInfo\": null,\n        \"categoryClassification\": null,\n        \"checkDetails\": \"\",\n        \"duration\": 0,\n        \"failureLineContent\": \"\",\n        \"failureLocation\": \"\",\n        \"skipReason\": \"\",\n        \"startTime\": \"\",\n        \"state\": \"failed\",\n        \"testID\": {\n          \"id\": \"test-case1\",\n          \"suite\": \"test-suite1\",\n          \"tags\": \"\"\n        }\n      }\n    },\n    \"versions\": {\n      \"claimFormat\": \"\",\n      \"k8s\": \"\",\n      \"ocClient\": \"\",\n      \"ocp\": \"\",\n      \"certsuite\": \"1.0.0\",\n      \"certsuiteGitCommit\": \"\"\n    }\n  }\n}")
}

func TestWriteClaimOutput(t *testing.T) {
	testClaimRoot := &claim.Root{
		Claim: &claim.Claim{
			Metadata: &claim.Metadata{
				StartTime: "2023-12-20 14:51:33 -0600 MST",
				EndTime:   "2023-12-20 14:51:34 -0600 MST",
			},
			Versions: &claim.Versions{
				CertSuite: "1.0.0",
			},
			Results: map[string]claim.Result{
				"test-case1": {
					TestID: &claim.Identifier{
						Id:    "test-case1",
						Suite: "test-suite1",
					},
					State: "failed",
				},
			},
		},
	}

	outputFile := "testfile_writeclaimoutput.json"
	claimOutput := MarshalClaimOutput(testClaimRoot)
	WriteClaimOutput(outputFile, claimOutput)
	defer os.Remove(outputFile)

	// read the file and compare the contents
	output, err := os.ReadFile(outputFile)
	assert.Nil(t, err)

	// Check if the output is a valid JSON
	//nolint:lll
	assert.Contains(t, string(output), "{\n  \"claim\": {\n    \"configurations\": null,\n    \"metadata\": {\n      \"endTime\": \"2023-12-20 14:51:34 -0600 MST\",\n      \"startTime\": \"2023-12-20 14:51:33 -0600 MST\"\n    },\n    \"nodes\": null,\n    \"results\": {\n      \"test-case1\": {\n        \"capturedTestOutput\": \"\",\n        \"catalogInfo\": null,\n        \"categoryClassification\": null,\n        \"checkDetails\": \"\",\n        \"duration\": 0,\n        \"failureLineContent\": \"\",\n        \"failureLocation\": \"\",\n        \"skipReason\": \"\",\n        \"startTime\": \"\",\n        \"state\": \"failed\",\n        \"testID\": {\n          \"id\": \"test-case1\",\n          \"suite\": \"test-suite1\",\n          \"tags\": \"\"\n        }\n      }\n    },\n    \"versions\": {\n      \"claimFormat\": \"\",\n      \"k8s\": \"\",\n      \"ocClient\": \"\",\n      \"ocp\": \"\",\n      \"certsuite\": \"1.0.0\",\n      \"certsuiteGitCommit\": \"\"\n    }\n  }\n}")

	// Assert the file permissions are 0644
	fileInfo, err := os.Stat(outputFile)
	assert.Nil(t, err)
	assert.Equal(t, "-rw-r--r--", fileInfo.Mode().String())
}

func TestReadClaimFile(t *testing.T) {
	testClaimRoot := &claim.Root{
		Claim: &claim.Claim{
			Metadata: &claim.Metadata{
				StartTime: "2023-12-20 14:51:33 -0600 MST",
				EndTime:   "2023-12-20 14:51:34 -0600 MST",
			},
			Versions: &claim.Versions{
				CertSuite: "1.0.0",
			},
			Results: map[string]claim.Result{
				"test-case1": {
					TestID: &claim.Identifier{
						Id:    "test-case1",
						Suite: "test-suite1",
					},
					State: "failed",
				},
			},
		},
	}

	outputFile := "testfile_readclaimfile.json"
	claimOutput := MarshalClaimOutput(testClaimRoot)
	WriteClaimOutput(outputFile, claimOutput)
	defer os.Remove(outputFile)

	// read the file and compare the contents
	output, err := ReadClaimFile(outputFile)
	assert.Nil(t, err)

	// Check if the output is a valid JSON
	assert.Contains(t, string(output), "test-case1")
}
