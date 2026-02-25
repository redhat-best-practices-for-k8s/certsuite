// Copyright (C) 2022-2026 Red Hat, Inc.
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

package preflight

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/stretchr/testify/assert"
)

func TestLabelsAllowTestRun(t *testing.T) {
	testCases := []struct {
		testLabelFilter   string
		testAllowedLabels []string
		expectedOutput    bool
	}{
		{ // Test Case #1 - Label filter is empty, nothing to test, test cases not allowed
			testLabelFilter:   "",
			testAllowedLabels: []string{common.PreflightTestKey, identifiers.TagCommon},
			expectedOutput:    false,
		},
		{ // Test Case #2 - Label filter matches other suite's test, test cases not allowed
			testLabelFilter:   "platform-alteration-isredhat-release",
			testAllowedLabels: []string{common.PreflightTestKey, identifiers.TagCommon},
			expectedOutput:    false,
		},
		{ // Test Case #3 - Label filter is a preflight test, test is allowed
			testLabelFilter:   "preflight-IsRedhatRelease",
			testAllowedLabels: []string{common.PreflightTestKey, identifiers.TagCommon},
			expectedOutput:    true,
		},
		{ // Test Case #4 - Label filter is a preflight test, test not allowed because missing allowed label
			testLabelFilter:   "preflight-IsRedhatRelease",
			testAllowedLabels: []string{identifiers.TagCommon},
			expectedOutput:    false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, labelsAllowTestRun(tc.testLabelFilter, tc.testAllowedLabels))
	}
}

// ---- TestGetUniqueTestEntriesFromContainerResults ----

func TestGetUniqueTestEntriesFromContainerResults_Empty(t *testing.T) {
	containers := []*provider.Container{}
	result := getUniqueTestEntriesFromContainerResults(containers)
	assert.Empty(t, result)
}

func TestGetUniqueTestEntriesFromContainerResults_PassedResults(t *testing.T) {
	containers := []*provider.Container{
		{
			PreflightResults: provider.PreflightResultsDB{
				Passed: []provider.PreflightTest{
					{Name: "test1", Description: "desc1", Remediation: "fix1"},
					{Name: "test2", Description: "desc2", Remediation: "fix2"},
				},
			},
		},
	}

	result := getUniqueTestEntriesFromContainerResults(containers)
	assert.Len(t, result, 2)
	assert.Equal(t, "desc1", result["test1"].Description)
	assert.Equal(t, "desc2", result["test2"].Description)
}

func TestGetUniqueTestEntriesFromContainerResults_Deduplicated(t *testing.T) {
	containers := []*provider.Container{
		{
			PreflightResults: provider.PreflightResultsDB{
				Passed: []provider.PreflightTest{
					{Name: "test1", Description: "desc1"},
				},
			},
		},
		{
			PreflightResults: provider.PreflightResultsDB{
				Passed: []provider.PreflightTest{
					{Name: "test1", Description: "desc1-dup"},
				},
			},
		},
	}

	result := getUniqueTestEntriesFromContainerResults(containers)
	assert.Len(t, result, 1)
}

func TestGetUniqueTestEntriesFromContainerResults_MixedResults(t *testing.T) {
	containers := []*provider.Container{
		{
			PreflightResults: provider.PreflightResultsDB{
				Passed: []provider.PreflightTest{
					{Name: "test1", Description: "passed desc"},
				},
				Failed: []provider.PreflightTest{
					{Name: "test2", Description: "failed desc"},
				},
				Errors: []provider.PreflightTest{
					{Name: "test3", Description: "error desc"},
				},
			},
		},
	}

	result := getUniqueTestEntriesFromContainerResults(containers)
	assert.Len(t, result, 3)
	assert.Equal(t, "passed desc", result["test1"].Description)
	assert.Equal(t, "failed desc", result["test2"].Description)
	assert.Equal(t, "error desc", result["test3"].Description)
}

// ---- TestGetUniqueTestEntriesFromOperatorResults ----

func TestGetUniqueTestEntriesFromOperatorResults_Empty(t *testing.T) {
	operators := []*provider.Operator{}
	result := getUniqueTestEntriesFromOperatorResults(operators)
	assert.Empty(t, result)
}

func TestGetUniqueTestEntriesFromOperatorResults_PassedResults(t *testing.T) {
	operators := []*provider.Operator{
		{
			Name: "op1",
			PreflightResults: provider.PreflightResultsDB{
				Passed: []provider.PreflightTest{
					{Name: "test1", Description: "desc1", Remediation: "fix1"},
				},
				Failed: []provider.PreflightTest{
					{Name: "test2", Description: "desc2", Remediation: "fix2"},
				},
			},
		},
	}

	result := getUniqueTestEntriesFromOperatorResults(operators)
	assert.Len(t, result, 2)
	assert.Equal(t, "desc1", result["test1"].Description)
	assert.Equal(t, "desc2", result["test2"].Description)
}

func TestGetUniqueTestEntriesFromOperatorResults_Deduplicated(t *testing.T) {
	operators := []*provider.Operator{
		{
			Name: "op1",
			PreflightResults: provider.PreflightResultsDB{
				Passed: []provider.PreflightTest{
					{Name: "test1", Description: "desc1"},
				},
			},
		},
		{
			Name: "op2",
			PreflightResults: provider.PreflightResultsDB{
				Passed: []provider.PreflightTest{
					{Name: "test1", Description: "desc1-dup"},
				},
			},
		},
	}

	result := getUniqueTestEntriesFromOperatorResults(operators)
	assert.Len(t, result, 1)
}
