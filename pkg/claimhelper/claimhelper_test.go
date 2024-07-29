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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
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
