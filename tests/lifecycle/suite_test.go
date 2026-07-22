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

package lifecycle

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/stretchr/testify/assert"
)

func TestNameInDeploymentSkipList(t *testing.T) {
	testCases := []struct {
		name           string
		testName       string
		testNamespace  string
		testList       []configuration.SkipScalingTestDeploymentsInfo
		expectedOutput bool
	}{
		{
			name:          "match found",
			testName:      "test1",
			testNamespace: "tnf",
			testList: []configuration.SkipScalingTestDeploymentsInfo{
				{Name: "test1", Namespace: "tnf"},
			},
			expectedOutput: true,
		},
		{
			name:          "name matches but namespace does not",
			testName:      "test1",
			testNamespace: "other-ns",
			testList: []configuration.SkipScalingTestDeploymentsInfo{
				{Name: "test1", Namespace: "tnf"},
			},
			expectedOutput: false,
		},
		{
			name:           "empty list",
			testName:       "test1",
			testNamespace:  "tnf",
			testList:       []configuration.SkipScalingTestDeploymentsInfo{},
			expectedOutput: false,
		},
		{
			name:          "no match",
			testName:      "test2",
			testNamespace: "tnf",
			testList: []configuration.SkipScalingTestDeploymentsInfo{
				{Name: "test1", Namespace: "tnf"},
			},
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedOutput, nameInDeploymentSkipList(tc.testName, tc.testNamespace, tc.testList))
		})
	}
}

func TestNameInStatefulSetSkipList(t *testing.T) {
	testCases := []struct {
		name           string
		testName       string
		testNamespace  string
		testList       []configuration.SkipScalingTestStatefulSetsInfo
		expectedOutput bool
	}{
		{
			name:          "match found",
			testName:      "test1",
			testNamespace: "tnf",
			testList: []configuration.SkipScalingTestStatefulSetsInfo{
				{Name: "test1", Namespace: "tnf"},
			},
			expectedOutput: true,
		},
		{
			name:          "name matches but namespace does not",
			testName:      "test1",
			testNamespace: "other-ns",
			testList: []configuration.SkipScalingTestStatefulSetsInfo{
				{Name: "test1", Namespace: "tnf"},
			},
			expectedOutput: false,
		},
		{
			name:           "empty list",
			testName:       "test1",
			testNamespace:  "tnf",
			testList:       []configuration.SkipScalingTestStatefulSetsInfo{},
			expectedOutput: false,
		},
		{
			name:          "no match",
			testName:      "test2",
			testNamespace: "tnf",
			testList: []configuration.SkipScalingTestStatefulSetsInfo{
				{Name: "test1", Namespace: "tnf"},
			},
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedOutput, nameInStatefulSetSkipList(tc.testName, tc.testNamespace, tc.testList))
		})
	}
}
