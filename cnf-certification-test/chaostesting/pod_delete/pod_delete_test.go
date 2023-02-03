// Copyright (C) 2020-2023 Red Hat, Inc.
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
package poddelete

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetLabelDeploymentValue(t *testing.T) {
	generateEnv := func(name, prefix, value string) *provider.TestEnvironment {
		return &provider.TestEnvironment{
			Config: configuration.TestConfiguration{
				TargetPodLabels: []configuration.Label{
					{
						Name:   name,
						Prefix: prefix,
						Value:  value,
					},
				},
			},
		}
	}

	testCases := []struct {
		testName       string
		testPrefix     string
		testValue      string
		testMap        map[string]string
		expectedOutput string
		expectedErr    error
	}{
		{ // Test Case #1 - Happy path
			testName:   "test1",
			testPrefix: "prefix1",
			testValue:  "value1",
			testMap: map[string]string{
				"prefix1/test1": "value1",
			},
			expectedOutput: "prefix1/test1=value1",
		},
		{ // Test Case #2 - Missing map value
			testName:       "test1",
			testPrefix:     "prefix1",
			testValue:      "value1",
			testMap:        map[string]string{},
			expectedOutput: "",
			expectedErr:    errors.New("did not find a key and value that matching the deployment"),
		},
		{ // Test Case #3 - Happy path, no prefix
			testName:   "test1",
			testPrefix: "",
			testValue:  "value1",
			testMap: map[string]string{
				"test1": "value1",
			},
			expectedOutput: "test1=value1",
		},
	}

	for _, tc := range testCases {
		result, err := GetLabelDeploymentValue(generateEnv(tc.testName, tc.testPrefix, tc.testValue), tc.testMap)
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedOutput, result)
	}
}

func TestParseLitmusResult(t *testing.T) {
	testCases := []struct {
		testList       *unstructured.UnstructuredList
		expectedOutput bool
	}{
		{ // Test Case #1 - Two objects, return false
			testList: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{},
					},
					{
						Object: map[string]interface{}{},
					},
				},
			},
			expectedOutput: false,
		},
		{ // Test Case 2 - status not set, return false
			testList: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{
							"wrongKeyName": nil,
						},
					},
				},
			},
			expectedOutput: false,
		},
		{ // Test Case 3 - "experiments" set nil, return false
			testList: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{
							"status": map[string]interface{}{
								"experiments": nil,
							},
						},
					},
				},
			},
			expectedOutput: false,
		},
		// TODO: Add more tests here parsing the []interfaces.
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, parseLitmusResult(tc.testList))
	}
}
