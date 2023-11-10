// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later
// version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
// warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with this program; if not, write to the Free
// Software Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package claim

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	expectedMarshallJSONError bool
	expectedStartTime         string
	expectedEndTime           string
}

var testCases = map[string]*testCase{
	"claim-valid": {
		expectedMarshallJSONError: false,
		expectedStartTime:         "1970-01-01T10:05:08+01:00",
		expectedEndTime:           "1970-01-01T10:05:08+01:00",
	},
	"claim-invalid-junit-payload": {
		expectedMarshallJSONError: true,
	},
	"claim-invalid-additional-property": {
		expectedMarshallJSONError: true,
	},
	"claim-invalid-bool-results": {
		expectedMarshallJSONError: true,
	},
	// A little confusing;  since we remap the "results" field, this interface{} value is not actually checked.
	// This is a limitation of the JSON Schema go client generator, and is perfectly fine for this context.
	"claim-invalid-non-result-result": {
		expectedMarshallJSONError: false,
	},
	"invalid-json": {
		expectedMarshallJSONError: true,
	},
	"missing-claim": {
		expectedMarshallJSONError: true,
	},
}

func getTestFile(testCaseName string) string {
	return path.Join("testdata", testCaseName+".json")
}

func getTestFileContents(testCaseName string) ([]byte, error) {
	testFilePath := getTestFile(testCaseName)
	return os.ReadFile(testFilePath)
}

func TestRoot_MarshalJSON(t *testing.T) {
	for testCaseName, testCaseDefinition := range testCases {
		contents, err := getTestFileContents(testCaseName)

		// raw data read tests
		assert.Nil(t, err)
		assert.NotNil(t, contents)

		// try to UnmarshallJSON the input
		root := &Root{}
		err = json.Unmarshal(contents, root)
		fmt.Println(testCaseName)
		assert.Equal(t, testCaseDefinition.expectedMarshallJSONError, err != nil)

		if testCaseDefinition.expectedMarshallJSONError == false {
			// start time assertion
			assert.Equal(t, "1970-01-01T10:05:08+01:00", root.Claim.Metadata.StartTime)
			assert.Equal(t, "1970-01-01T10:05:08+01:00", root.Claim.Metadata.EndTime)

			generatedContents, err := json.Marshal(root)
			assert.Nil(t, err)
			assert.NotNil(t, generatedContents)
		}
	}
}

func TestResult_MarshalJSON(t *testing.T) {
	type fields struct {
		CapturedTestOutput     string
		CatalogInfo            *CatalogInfo
		CategoryClassification *CategoryClassification
		Duration               int
		EndTime                string
		FailureLineContent     string
		FailureLocation        string
		FailureReason          string
		StartTime              string
		State                  string
		TestID                 *Identifier
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strct := &Result{
				CapturedTestOutput:     tt.fields.CapturedTestOutput,
				CatalogInfo:            tt.fields.CatalogInfo,
				CategoryClassification: tt.fields.CategoryClassification,
				Duration:               tt.fields.Duration,
				EndTime:                tt.fields.EndTime,
				FailureLineContent:     tt.fields.FailureLineContent,
				FailureLocation:        tt.fields.FailureLocation,
				FailureReason:          tt.fields.FailureReason,
				StartTime:              tt.fields.StartTime,
				State:                  tt.fields.State,
				TestID:                 tt.fields.TestID,
			}
			got, err := strct.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Result.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Result.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_UnmarshalJSON(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		want    *Result
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			want: &Result{
				//nolint:lll
				CapturedTestOutput: "{\"CompliantObjectsOut\":[{\"ObjectType\":\"Container\",\"ObjectFieldsKeys\":[\"Reason For Compliance\",\"Namespace\",\"Pod Name\",\"Container Name\"],\"ObjectFieldsValues\":[\"Container is not modified\",\"tnf\",\"test-0\",\"test\"]},{\"ObjectType\":\"Container\",\"ObjectFieldsKeys\":[\"Reason For Compliance\",\"Namespace\",\"Pod Name\",\"Container Name\"],\"ObjectFieldsValues\":[\"Container is not modified\",\"tnf\",\"test-1\",\"test\"]},{\"ObjectType\":\"Container\",\"ObjectFieldsKeys\":[\"Reason For Compliance\",\"Namespace\",\"Pod Name\",\"Container Name\"],\"ObjectFieldsValues\":[\"Container is not modified\",\"tnf\",\"test-d78fbf8d6-jxgl2\",\"test\"]},{\"ObjectType\":\"Container\",\"ObjectFieldsKeys\":[\"Reason For Compliance\",\"Namespace\",\"Pod Name\",\"Container Name\"],\"ObjectFieldsValues\":[\"Container is not modified\",\"tnf\",\"test-d78fbf8d6-n4jlv\",\"test\"]}],\"NonCompliantObjectsOut\":null}\n",
				CatalogInfo: &CatalogInfo{
					BestPracticeReference: "https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-image-standards",
					//nolint:lll
					Description:      "Ensures that the Container Base Image is not altered post-startup. This test is a heuristic, and ensures that there are no changes to the following directories: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64",
					ExceptionProcess: "No exceptions",
					//nolint:lll
					Remediation: "Ensure that Container applications do not modify the Container Base Image. In particular, ensure that the following directories are not modified: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64 Ensure that all required binaries are built directly into the container image, and are not installed post startup.",
				},
				CategoryClassification: &CategoryClassification{
					Extended: "Mandatory",
					FarEdge:  "Mandatory",
					NonTelco: "Mandatory",
					Telco:    "Mandatory",
				},
				Duration:           7745320473,
				EndTime:            "2023-07-25 09:10:25.557493221 -0500 CDT m=+51.038323513",
				FailureLineContent: "",
				FailureLocation:    ":0",
				FailureReason:      "",
				StartTime:          "2023-07-25 09:10:17.812172748 -0500 CDT m=+43.293003040",
				State:              "passed",
				TestID: &Identifier{
					Id:    "platform-alteration-base-image",
					Suite: "platform-alteration",
					Tags:  "common",
				},
			},
			//nolint:lll
			args: args{b: []byte(`
				  {
					"capturedTestOutput": "{\"CompliantObjectsOut\":[{\"ObjectType\":\"Container\",\"ObjectFieldsKeys\":[\"Reason For Compliance\",\"Namespace\",\"Pod Name\",\"Container Name\"],\"ObjectFieldsValues\":[\"Container is not modified\",\"tnf\",\"test-0\",\"test\"]},{\"ObjectType\":\"Container\",\"ObjectFieldsKeys\":[\"Reason For Compliance\",\"Namespace\",\"Pod Name\",\"Container Name\"],\"ObjectFieldsValues\":[\"Container is not modified\",\"tnf\",\"test-1\",\"test\"]},{\"ObjectType\":\"Container\",\"ObjectFieldsKeys\":[\"Reason For Compliance\",\"Namespace\",\"Pod Name\",\"Container Name\"],\"ObjectFieldsValues\":[\"Container is not modified\",\"tnf\",\"test-d78fbf8d6-jxgl2\",\"test\"]},{\"ObjectType\":\"Container\",\"ObjectFieldsKeys\":[\"Reason For Compliance\",\"Namespace\",\"Pod Name\",\"Container Name\"],\"ObjectFieldsValues\":[\"Container is not modified\",\"tnf\",\"test-d78fbf8d6-n4jlv\",\"test\"]}],\"NonCompliantObjectsOut\":null}\n",
					"catalogInfo": {
					  "bestPracticeReference": "https://test-network-function.github.io/cnf-best-practices/#cnf-best-practices-image-standards",
					  "description": "Ensures that the Container Base Image is not altered post-startup. This test is a heuristic, and ensures that there are no changes to the following directories: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64",
					  "exceptionProcess": "No exceptions",
					  "remediation": "Ensure that Container applications do not modify the Container Base Image. In particular, ensure that the following directories are not modified: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64 Ensure that all required binaries are built directly into the container image, and are not installed post startup."
					},
					"categoryClassification": {
					  "Extended": "Mandatory",
					  "FarEdge": "Mandatory",
					  "NonTelco": "Mandatory",
					  "Telco": "Mandatory"
					},
					"duration": 7745320473,
					"endTime": "2023-07-25 09:10:25.557493221 -0500 CDT m=+51.038323513",
					"failureLineContent": "",
					"failureLocation": ":0",
					"failureReason": "",
					"startTime": "2023-07-25 09:10:17.812172748 -0500 CDT m=+43.293003040",
					"state": "passed",
					"testID": {
					  "id": "platform-alteration-base-image",
					  "suite": "platform-alteration",
					  "tags": "common"
					}
				  }  
			`)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &Result{}
			if err := got.UnmarshalJSON(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Result.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			deepEqual(got, tt.want)
		})
	}
}

func deepEqual(r1, r2 *Result) bool {
	r1Nil, r2Nil := r1, r2
	r1Nil.TestID = nil
	r1Nil.CategoryClassification = nil
	r1Nil.CatalogInfo = nil
	r2Nil.TestID = nil
	r2Nil.CategoryClassification = nil
	r2Nil.CatalogInfo = nil

	equal := reflect.DeepEqual(r1, r2)
	if !equal {
		return false
	}
	equal = reflect.DeepEqual(r1.CategoryClassification, r2.CategoryClassification)
	if !equal {
		return false
	}
	reflect.DeepEqual(r1.TestID, r2.TestID)
	if !equal {
		return false
	}
	reflect.DeepEqual(r1.CatalogInfo, r2.CatalogInfo)

	return equal
}
