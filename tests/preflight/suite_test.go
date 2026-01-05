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

	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/stretchr/testify/assert"
)

func TestGetUniqueTestEntriesFromContainerResults(_ *testing.T) {
	// Note: preflight lib does not expose their underlying plibRuntime.Result struct vars so I can not write unit tests for this
}

func TestGetUniqueTestEntriesFromOperatorResults(_ *testing.T) {
	// Note: preflight lib does not expose their underlying plibRuntime.Result struct vars so I can not write unit tests for this
}

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
		{ // Test Case #3 - Label filter is a preflight test, test not allowed because missing allowed label
			testLabelFilter:   "preflight-IsRedhatRelease",
			testAllowedLabels: []string{identifiers.TagCommon},
			expectedOutput:    false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, labelsAllowTestRun(tc.testLabelFilter, tc.testAllowedLabels))
	}
}
