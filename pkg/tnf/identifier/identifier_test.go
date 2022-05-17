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

package identifier_test

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf/identifier"
)

type testIdentifierUnmarshalJSONTestCase struct {
	expectedUnmarshalErr     bool
	expectedUnmarshalErrText string
	expectedIdentifier       identifier.Identifier
}

var testIdentifierUnmarshalJSONTestCases = map[string]testIdentifierUnmarshalJSONTestCase{
	"valid_identifier": {
		expectedUnmarshalErr: false,
		expectedIdentifier: identifier.Identifier{
			URL:             "http://test-network-function.com/tests/generic/ping",
			SemanticVersion: "v0.0.1",
		},
	},
	"bad_url": {
		expectedUnmarshalErr:     true,
		expectedUnmarshalErrText: "parse \"http://[::1]:namedport\": invalid port \":namedport\" after host",
	},
	"bad_version": {
		expectedUnmarshalErr:     true,
		expectedUnmarshalErrText: "Invalid Semantic Version",
	},
	"missing_url": {
		expectedUnmarshalErr:     true,
		expectedUnmarshalErrText: "missing required field: \"url\"",
	},
	"missing_version": {
		expectedUnmarshalErr:     true,
		expectedUnmarshalErrText: "missing required field: \"version\"",
	},
	"not_json": {
		expectedUnmarshalErr:     true,
		expectedUnmarshalErrText: "invalid character 'T' looking for beginning of value",
	},
}

func getTestFile(testName string) string {
	return path.Join("testdata", testName+".json")
}

func TestIdentifier_UnmarshalJSON(t *testing.T) {
	for testName, testCase := range testIdentifierUnmarshalJSONTestCases {
		testFile := getTestFile(testName)
		contents, err := os.ReadFile(testFile)
		assert.Nil(t, err)
		assert.NotNil(t, contents)

		var actualIdentifier identifier.Identifier
		err = json.Unmarshal(contents, &actualIdentifier)
		assert.Equal(t, testCase.expectedUnmarshalErr, err != nil)
		if !testCase.expectedUnmarshalErr {
			assert.Equal(t, testCase.expectedIdentifier, actualIdentifier)
		} else {
			assert.Equal(t, testCase.expectedUnmarshalErrText, err.Error())
		}
	}
}

func TestGetShortNameFromIdentifier(t *testing.T) {
	type testURLTestName struct {
		URL            string
		testName       string
		expectedResult string
	}
	testsURLs := []testURLTestName{
		{
			URL:            identifier.GetIdentifierURLBaseDomain() + "/command",
			testName:       "command",
			expectedResult: "command",
		},
		{
			URL:            identifier.GetIdentifierURLBaseDomain() + "/whatever",
			testName:       "whatever",
			expectedResult: "whatever",
		},
		{
			URL:            "http://test-network-function.org/tests" + "/command",
			testName:       "command",
			expectedResult: "",
		},
		{
			URL:            "http://test-network-function.es/functional-tests" + "/whatever",
			testName:       "whatever",
			expectedResult: "",
		},
	}

	for _, test := range testsURLs {
		id := identifier.Identifier{URL: test.URL, SemanticVersion: ""}
		assert.Equal(t, test.expectedResult, identifier.GetShortNameFromIdentifier(id))
	}
}
