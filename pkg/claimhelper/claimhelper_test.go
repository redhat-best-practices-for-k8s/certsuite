// Copyright (C) 2020-2026 Red Hat, Inc.
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
	j "encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				Telco:    unitTestEnvTrue,
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
						Telco:    unitTestEnvTrue,
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
						Telco:    unitTestEnvTrue,
					},
				},
			},
			expectedXMLResult: "<testsuites tests=\"1\" disabled=\"0\" errors=\"0\" failures=\"0\" time=\"1.00000\">",
		},
	}

	t.Setenv("UNIT_TEST", unitTestEnvTrue)
	defer os.Remove("testfile.xml")
	defer os.Unsetenv("UNIT_TEST")

	for _, tc := range testCases {
		// Build some 1 minute duration start and end time
		startTime, err := time.Parse(DateTimeFormatDirective, "2023-12-20 14:51:33 -0600 MST")
		assert.Nil(t, err)
		endTime, err := time.Parse(DateTimeFormatDirective, "2023-12-20 14:51:34 -0600 MST")
		assert.Nil(t, err)

		testClaimBuilder, err := NewClaimBuilder(&provider.TestEnvironment{})
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
	assert.Contains(t, string(output), "{\n  \"claim\": {\n    \"configurations\": null,\n    \"metadata\": {\n      \"endTime\": \"2023-12-20 14:51:34 -0600 MST\",\n      \"startTime\": \"2023-12-20 14:51:33 -0600 MST\"\n    },\n    \"nodes\": null,\n    \"results\": {\n      \"test-case1\": {\n        \"capturedTestOutput\": \"\",\n        \"catalogInfo\": null,\n        \"categoryClassification\": null,\n        \"checkDetails\": \"\",\n        \"duration\": 0,\n        \"failureLineContent\": \"\",\n        \"failureLocation\": \"\",\n        \"skipReason\": \"\",\n        \"startTime\": \"\",\n        \"state\": \"failed\",\n        \"testID\": {\n          \"id\": \"test-case1\",\n          \"suite\": \"test-suite1\",\n          \"tags\": \"\"\n        }\n      }\n    },\n    \"versions\": {\n      \"certSuite\": \"1.0.0\",\n      \"certSuiteGitCommit\": \"\",\n      \"claimFormat\": \"\",\n      \"k8s\": \"\",\n      \"ocClient\": \"\",\n      \"ocp\": \"\"\n    }\n  }\n}")
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
	assert.Contains(t, string(output), "{\n  \"claim\": {\n    \"configurations\": null,\n    \"metadata\": {\n      \"endTime\": \"2023-12-20 14:51:34 -0600 MST\",\n      \"startTime\": \"2023-12-20 14:51:33 -0600 MST\"\n    },\n    \"nodes\": null,\n    \"results\": {\n      \"test-case1\": {\n        \"capturedTestOutput\": \"\",\n        \"catalogInfo\": null,\n        \"categoryClassification\": null,\n        \"checkDetails\": \"\",\n        \"duration\": 0,\n        \"failureLineContent\": \"\",\n        \"failureLocation\": \"\",\n        \"skipReason\": \"\",\n        \"startTime\": \"\",\n        \"state\": \"failed\",\n        \"testID\": {\n          \"id\": \"test-case1\",\n          \"suite\": \"test-suite1\",\n          \"tags\": \"\"\n        }\n      }\n    },\n    \"versions\": {\n      \"certSuite\": \"1.0.0\",\n      \"certSuiteGitCommit\": \"\",\n      \"claimFormat\": \"\",\n      \"k8s\": \"\",\n      \"ocClient\": \"\",\n      \"ocp\": \"\"\n    }\n  }\n}")

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

func TestCreateClaimRoot(t *testing.T) {
	t.Parallel()

	before := time.Now().UTC().Truncate(time.Second)
	root := CreateClaimRoot()

	require.NotNil(t, root)
	require.NotNil(t, root.Claim)
	require.NotNil(t, root.Claim.Metadata)
	assert.NotEmpty(t, root.Claim.Metadata.StartTime)

	startTime, err := time.Parse(DateTimeFormatDirective, root.Claim.Metadata.StartTime)
	require.NoError(t, err)
	assert.False(t, startTime.Before(before), "startTime %v should not be before %v", startTime, before)
}

func TestClaimBuilderReset(t *testing.T) {
	t.Parallel()

	root := &claim.Root{
		Claim: &claim.Claim{
			Metadata: &claim.Metadata{
				StartTime: "2020-01-01 00:00:00 +0000 UTC",
			},
		},
	}
	builder := &ClaimBuilder{claimRoot: root}

	builder.Reset()

	resetTime, err := time.Parse(DateTimeFormatDirective, root.Claim.Metadata.StartTime)
	require.NoError(t, err)

	pastTime, err := time.Parse(DateTimeFormatDirective, "2020-01-01 00:00:00 +0000 UTC")
	require.NoError(t, err)

	assert.True(t, resetTime.After(pastTime))
}

func TestMarshalConfigurations(t *testing.T) {
	t.Parallel()

	env := &provider.TestEnvironment{}
	data, err := MarshalConfigurations(env)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var parsed map[string]interface{}
	err = j.Unmarshal(data, &parsed)
	assert.NoError(t, err)
}

func TestUnmarshalConfigurations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		key   string
		value interface{}
	}{
		{
			name:  "simple key-value",
			input: `{"foo": "bar"}`,
			key:   "foo",
			value: "bar",
		},
		{
			name:  "nested object",
			input: `{"outer": {"inner": 42}}`,
			key:   "outer",
		},
		{
			name:  "empty object",
			input: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := map[string]interface{}{}
			UnmarshalConfigurations([]byte(tt.input), result)
			if tt.key != "" {
				assert.Contains(t, result, tt.key)
				if tt.value != nil {
					assert.Equal(t, tt.value, result[tt.key])
				}
			}
		})
	}
}

func TestUnmarshalClaim(t *testing.T) {
	t.Parallel()

	claimJSON := `{
		"claim": {
			"metadata": {
				"startTime": "2024-01-15 10:00:00 +0000 UTC",
				"endTime": "2024-01-15 10:05:00 +0000 UTC"
			},
			"configurations": {},
			"nodes": null,
			"versions": {
				"certSuite": "test",
				"certSuiteGitCommit": "",
				"claimFormat": "",
				"k8s": "",
				"ocClient": "",
				"ocp": ""
			},
			"results": {
				"test-1": {
					"capturedTestOutput": "",
					"catalogInfo": null,
					"categoryClassification": null,
					"checkDetails": "",
					"duration": 0,
					"failureLineContent": "",
					"failureLocation": "",
					"skipReason": "",
					"startTime": "",
					"state": "passed",
					"testID": {"id": "test-1", "suite": "suite-1", "tags": ""}
				}
			}
		}
	}`

	var root claim.Root
	UnmarshalClaim([]byte(claimJSON), &root)

	require.NotNil(t, root.Claim)
	require.NotNil(t, root.Claim.Metadata)
	assert.Equal(t, "2024-01-15 10:00:00 +0000 UTC", root.Claim.Metadata.StartTime)
	require.Contains(t, root.Claim.Results, "test-1")
	assert.Equal(t, "passed", root.Claim.Results["test-1"].State)
}

func TestReadClaimFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := ReadClaimFile("nonexistent_claim_file.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent_claim_file.json")
}

func TestGetConfigurationFromClaimFile(t *testing.T) {
	t.Parallel()

	claimRoot := &claim.Root{
		Claim: &claim.Claim{
			Metadata: &claim.Metadata{
				StartTime: "2024-01-15 10:00:00 +0000 UTC",
				EndTime:   "2024-01-15 10:05:00 +0000 UTC",
			},
			Versions:       &claim.Versions{CertSuite: "test"},
			Configurations: map[string]interface{}{},
			Results:        map[string]claim.Result{},
		},
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), "claim-*.json")
	require.NoError(t, err)

	payload := MarshalClaimOutput(claimRoot)
	_, err = tmpFile.Write(payload)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	env, err := GetConfigurationFromClaimFile(tmpFile.Name())
	require.NoError(t, err)
	assert.NotNil(t, env)
}

func TestGetConfigurationFromClaimFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := GetConfigurationFromClaimFile("nonexistent.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent.json")
}

func TestSanitizeClaimFile(t *testing.T) {
	// Not parallel: mutates identifiers.TestIDToClaimID
	origTestIDToClaimID := identifiers.TestIDToClaimID
	t.Cleanup(func() { identifiers.TestIDToClaimID = origTestIDToClaimID })
	identifiers.TestIDToClaimID = map[string]claim.Identifier{}

	commonID := claim.Identifier{Id: "common-check", Suite: "test-suite", Tags: "common"}
	extendedID := claim.Identifier{Id: "extended-check", Suite: "test-suite", Tags: "extended"}

	claimRoot := &claim.Root{
		Claim: &claim.Claim{
			Metadata: &claim.Metadata{
				StartTime: "2024-01-15 10:00:00 +0000 UTC",
				EndTime:   "2024-01-15 10:05:00 +0000 UTC",
			},
			Versions:       &claim.Versions{CertSuite: "test"},
			Configurations: map[string]interface{}{},
			Results: map[string]claim.Result{
				"common-check": {
					TestID: &commonID,
					State:  "passed",
				},
				"extended-check": {
					TestID: &extendedID,
					State:  "passed",
				},
			},
		},
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), "sanitize-*.json")
	require.NoError(t, err)

	payload := MarshalClaimOutput(claimRoot)
	_, err = tmpFile.Write(payload)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	resultFile, err := SanitizeClaimFile(tmpFile.Name(), "common")
	require.NoError(t, err)
	assert.Equal(t, tmpFile.Name(), resultFile)

	// Read back and verify only the common check remains
	data, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	var sanitized claim.Root
	err = j.Unmarshal(data, &sanitized)
	require.NoError(t, err)

	assert.Contains(t, sanitized.Claim.Results, "common-check")
	assert.NotContains(t, sanitized.Claim.Results, "extended-check")
}

func TestSanitizeClaimFileInvalidFilter(t *testing.T) {
	// Not parallel: mutates identifiers.TestIDToClaimID
	origTestIDToClaimID := identifiers.TestIDToClaimID
	t.Cleanup(func() { identifiers.TestIDToClaimID = origTestIDToClaimID })
	identifiers.TestIDToClaimID = map[string]claim.Identifier{}

	testID := claim.Identifier{Id: "check-1", Suite: "suite", Tags: "common"}

	claimRoot := &claim.Root{
		Claim: &claim.Claim{
			Metadata: &claim.Metadata{
				StartTime: "2024-01-15 10:00:00 +0000 UTC",
				EndTime:   "2024-01-15 10:05:00 +0000 UTC",
			},
			Versions:       &claim.Versions{CertSuite: "test"},
			Configurations: map[string]interface{}{},
			Results: map[string]claim.Result{
				"check-1": {
					TestID: &testID,
					State:  "passed",
				},
			},
		},
	}

	tmpFile, err := os.CreateTemp(t.TempDir(), "sanitize-invalid-*.json")
	require.NoError(t, err)

	payload := MarshalClaimOutput(claimRoot)
	_, err = tmpFile.Write(payload)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	_, err = SanitizeClaimFile(tmpFile.Name(), "&&&&")
	require.Error(t, err)
}

func TestSanitizeClaimFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := SanitizeClaimFile("nonexistent.json", "common")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent.json")
}

func TestNewClaimBuilderUnitTest(t *testing.T) {
	t.Setenv("UNIT_TEST", unitTestEnvTrue)

	builder, err := NewClaimBuilder(&provider.TestEnvironment{})
	require.NoError(t, err)
	require.NotNil(t, builder)
	require.NotNil(t, builder.claimRoot)
	require.NotNil(t, builder.claimRoot.Claim)
	require.NotNil(t, builder.claimRoot.Claim.Metadata)
	assert.NotEmpty(t, builder.claimRoot.Claim.Metadata.StartTime)
}

func TestPopulateXMLFromClaimSkipped(t *testing.T) {
	t.Parallel()

	c := claim.Claim{
		Results: map[string]claim.Result{
			"skipped-test": {
				TestID:     &claim.Identifier{Id: "skipped-test", Suite: "suite"},
				State:      TestStateSkipped,
				SkipReason: "not applicable",
				StartTime:  "2024-01-15 10:00:00 +0000 UTC",
				EndTime:    "2024-01-15 10:00:01 +0000 UTC",
			},
		},
	}

	start, err := time.Parse(DateTimeFormatDirective, "2024-01-15 10:00:00 +0000 UTC")
	require.NoError(t, err)
	end, err := time.Parse(DateTimeFormatDirective, "2024-01-15 10:00:01 +0000 UTC")
	require.NoError(t, err)

	xml := populateXMLFromClaim(c, start, end)

	assert.Equal(t, "0", xml.Failures)
	assert.Equal(t, "1", xml.Disabled)
	assert.Equal(t, "1", xml.Tests)
	require.Len(t, xml.Testsuite.Testcase, 1)

	tc := xml.Testsuite.Testcase[0]
	assert.Equal(t, TestStateSkipped, tc.Status)
	require.NotNil(t, tc.Skipped)
	assert.Equal(t, "not applicable", tc.Skipped.Text)
	assert.Nil(t, tc.Failure)
}

func TestPopulateXMLFromClaimPassed(t *testing.T) {
	t.Parallel()

	c := claim.Claim{
		Results: map[string]claim.Result{
			"passed-test": {
				TestID:    &claim.Identifier{Id: "passed-test", Suite: "suite"},
				State:     "passed",
				StartTime: "2024-01-15 10:00:00 +0000 UTC",
				EndTime:   "2024-01-15 10:00:02 +0000 UTC",
			},
		},
	}

	start, err := time.Parse(DateTimeFormatDirective, "2024-01-15 10:00:00 +0000 UTC")
	require.NoError(t, err)
	end, err := time.Parse(DateTimeFormatDirective, "2024-01-15 10:00:02 +0000 UTC")
	require.NoError(t, err)

	xml := populateXMLFromClaim(c, start, end)

	assert.Equal(t, "0", xml.Failures)
	assert.Equal(t, "0", xml.Disabled)
	require.Len(t, xml.Testsuite.Testcase, 1)

	tc := xml.Testsuite.Testcase[0]
	assert.Equal(t, "passed", tc.Status)
	assert.Nil(t, tc.Skipped)
	assert.Nil(t, tc.Failure)
}

func TestPopulateXMLFromClaimMultipleResults(t *testing.T) {
	t.Parallel()

	c := claim.Claim{
		Results: map[string]claim.Result{
			"pass-1": {
				TestID:    &claim.Identifier{Id: "pass-1", Suite: "s"},
				State:     "passed",
				StartTime: "2024-01-15 10:00:00 +0000 UTC",
				EndTime:   "2024-01-15 10:00:01 +0000 UTC",
			},
			"fail-1": {
				TestID:       &claim.Identifier{Id: "fail-1", Suite: "s"},
				State:        TestStateFailed,
				CheckDetails: "assertion failed",
				StartTime:    "2024-01-15 10:00:00 +0000 UTC",
				EndTime:      "2024-01-15 10:00:01 +0000 UTC",
			},
			"skip-1": {
				TestID:     &claim.Identifier{Id: "skip-1", Suite: "s"},
				State:      TestStateSkipped,
				SkipReason: "n/a",
				StartTime:  "2024-01-15 10:00:00 +0000 UTC",
				EndTime:    "2024-01-15 10:00:01 +0000 UTC",
			},
		},
	}

	start, err := time.Parse(DateTimeFormatDirective, "2024-01-15 10:00:00 +0000 UTC")
	require.NoError(t, err)
	end, err := time.Parse(DateTimeFormatDirective, "2024-01-15 10:00:05 +0000 UTC")
	require.NoError(t, err)

	xml := populateXMLFromClaim(c, start, end)

	assert.Equal(t, "3", xml.Tests)
	assert.Equal(t, "1", xml.Failures)
	assert.Equal(t, "1", xml.Disabled)
	assert.Equal(t, "0", xml.Errors)
	assert.Len(t, xml.Testsuite.Testcase, 3)

	// Verify test cases are sorted by ID
	assert.Equal(t, "fail-1", xml.Testsuite.Testcase[0].Name)
	assert.Equal(t, "pass-1", xml.Testsuite.Testcase[1].Name)
	assert.Equal(t, "skip-1", xml.Testsuite.Testcase[2].Name)
}
