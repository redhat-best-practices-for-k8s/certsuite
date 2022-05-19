// Copyright (C) 2020-2022 Red Hat, Inc.
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

package stringhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:funlen
func TestStringInSlice(t *testing.T) {
	testCases := []struct {
		testSlice       []string
		testString      string
		containsFeature bool
		expected        bool
	}{
		{
			testSlice: []string{
				"apples",
				"bananas",
				"oranges",
			},
			testString:      "apples",
			containsFeature: false,
			expected:        true,
		},
		{
			testSlice: []string{
				"apples",
				"bananas",
				"oranges",
			},
			testString:      "tacos",
			containsFeature: false,
			expected:        false,
		},
		{
			testSlice: []string{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
			testString:      "intree:",
			containsFeature: true, // Note: Turn 'on' the contains check
			expected:        true,
		},
		{
			testSlice: []string{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
			testString:      "intree:",
			containsFeature: false, // Note: Turn 'off' the contains check
			expected:        false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, StringInSlice(tc.testSlice, tc.testString, tc.containsFeature))
	}
}

func TestCompareVersion(t *testing.T) {
	// Note: These values pertain to 'kubeVersion' fields found:
	// https://charts.openshift.io/index.yaml
	testCases := []struct {
		ver1           string
		ver2           string
		expectedOutput bool
	}{
		{
			ver1:           "1.18.1",
			ver2:           ">= 1.19",
			expectedOutput: false,
		},
		{
			ver1:           "1.19.1",
			ver2:           ">= 1.19",
			expectedOutput: true,
		},
		{
			ver1:           "1.19",
			ver2:           ">= 1.16.0 < 1.22.0",
			expectedOutput: true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, CompareVersion(tc.ver1, tc.ver2))
	}
}
