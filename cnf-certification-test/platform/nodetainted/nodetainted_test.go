// Copyright (C) 2021 Red Hat, Inc.
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

package nodetainted

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
)

func TestTaintsAccepted(t *testing.T) {
	testCases := []struct {
		confTaints     []configuration.AcceptedKernelTaintsInfo
		taintedModules []string
		expected       bool
	}{
		{
			confTaints: []configuration.AcceptedKernelTaintsInfo{
				{
					Module: "taint1",
				},
			},
			taintedModules: []string{
				"taint1",
			},
			expected: true,
		},
		{
			confTaints: []configuration.AcceptedKernelTaintsInfo{}, // no accepted modules
			taintedModules: []string{
				"taint1",
			},
			expected: false,
		},
		{ // We have no tainted modules, so the configuration does not matter.
			confTaints: []configuration.AcceptedKernelTaintsInfo{
				{
					Module: "taint1",
				},
			},
			taintedModules: []string{},
			expected:       true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, TaintsAccepted(tc.confTaints, tc.taintedModules))
	}
}

func TestDecodeKernelTaints(t *testing.T) {
	taint1, taint1Slice := DecodeKernelTaints(2048)
	assert.Equal(t, taint1, "workaround for bug in platform firmware applied, ")
	assert.Len(t, taint1Slice, 1)

	taint2, taint2Slice := DecodeKernelTaints(32769)
	assert.Equal(t, taint2, "proprietary module was loaded, kernel has been live patched, ")
	assert.Len(t, taint2Slice, 2)
}

func TestRemoveEmptyStrings(t *testing.T) {
	testCases := []struct {
		testSlice     []string
		expectedSlice []string
	}{
		{
			testSlice:     []string{"one", "two", "three", "", ""},
			expectedSlice: []string{"one", "two", "three"},
		},
		{ // returns a nil slice if the contents of the incoming slice are empty
			testSlice:     []string{"", ""},
			expectedSlice: nil,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedSlice, removeEmptyStrings(tc.testSlice))
	}
}
