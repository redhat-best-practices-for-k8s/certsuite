// Copyright (C) 2021-2022 Red Hat, Inc.
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

//nolint:funlen
func TestDecodeKernelTaints(t *testing.T) {
	tcs := []struct {
		taintsBitMask  uint64
		expectedTaints []string
	}{
		// No taints
		{
			taintsBitMask:  0,
			expectedTaints: []string{},
		},

		// Reserved taint bit 23
		{
			taintsBitMask:  1 << 23,
			expectedTaints: []string{"reserved kernel taint bit 23"},
		},

		// Taint bit 0
		{
			taintsBitMask:  1 << 0,
			expectedTaints: []string{"proprietary module was loaded"},
		},

		// Taint bit 11
		{
			taintsBitMask:  1 << 11,
			expectedTaints: []string{"workaround for bug in platform firmware applied"},
		},

		// Bit 18
		{
			taintsBitMask:  1 << 18,
			expectedTaints: []string{"an in-kernel test has been run"},
		},

		// Bits 0 and 18
		{
			taintsBitMask:  (1 << 0) | (1 << 18),
			expectedTaints: []string{"proprietary module was loaded", "an in-kernel test has been run"},
		},

		// Bits 1, 24 and 30
		{
			taintsBitMask:  (1 << 0) | (1 << 24) | (1 << 30),
			expectedTaints: []string{"proprietary module was loaded", "reserved kernel taint bit 24", "Red Hat extension: reserved taint bit 30"},
		},

		// RH's bit 29
		{
			taintsBitMask:  1 << 29,
			expectedTaints: []string{"Red Hat extension: Technology Preview code was loaded; cf. Technology Preview features support scope description. Refer to \"TECH PREVIEW:\" kernel log entry for details."},
		},

		// RH's reserved bit 31
		{
			taintsBitMask:  1 << 31,
			expectedTaints: []string{"Red Hat extension: reserved taint bit 31"},
		},
	}

	for _, tc := range tcs {
		taints := DecodeKernelTaints(tc.taintsBitMask)
		assert.Equal(t, tc.expectedTaints, taints)
	}
}
