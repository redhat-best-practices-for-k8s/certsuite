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

package sysctlconfig

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSysctlSystemOutput(t *testing.T) {
	testCases := []struct {
		sysctlOutput   string
		expectedResult map[string]string
	}{
		{ // Test Case #1 - Single parameter
			sysctlOutput:   "kernel.sysrq = 16",
			expectedResult: map[string]string{"kernel.sysrq": "16"},
		},
		{ // Test Case #2 - Multiple parameters
			sysctlOutput:   "kernel.sysrq = 16\nfs.protected_symlinks = 1\nnet.core.optmem_max = 81920\n",
			expectedResult: map[string]string{"kernel.sysrq": "16", "fs.protected_symlinks": "1", "net.core.optmem_max": "81920"},
		},
		{ // Test Case #3 - Skip lines starting with "*"
			sysctlOutput:   "kernel.sysrq = 16\n* Applying ...\nfs.protected_symlinks = 1\n",
			expectedResult: map[string]string{"kernel.sysrq": "16", "fs.protected_symlinks": "1"},
		},
		{ // Test Case #4 - Skip whitespaces
			sysctlOutput:   "    kernel.sysrq    =  16 ",
			expectedResult: map[string]string{"kernel.sysrq": "16"},
		},
		{ // Test Case #5 - No whitespaces
			sysctlOutput:   "kernel.sysrq=16",
			expectedResult: map[string]string{"kernel.sysrq": "16"},
		},
		{ // Test Case #6 - No regex match
			sysctlOutput:   "kernel.sysrq -> 16",
			expectedResult: map[string]string{},
		},
		{ // Test Case #7 - Empty output
			sysctlOutput:   "",
			expectedResult: map[string]string{},
		},
	}

	for _, tc := range testCases {
		assert.True(t, maps.Equal(parseSysctlSystemOutput(tc.sysctlOutput), tc.expectedResult))
	}
}
