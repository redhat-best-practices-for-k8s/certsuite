// Copyright (C) 2021-2026 Red Hat, Inc.
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
	"errors"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/stretchr/testify/assert"
)

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
			expectedTaints: []string{"reserved (tainted bit 23)"},
		},

		// Taint bit 0
		{
			taintsBitMask:  1 << 0,
			expectedTaints: []string{"proprietary module was loaded (tainted bit 0)"},
		},

		// Taint bit 11
		{
			taintsBitMask:  1 << 11,
			expectedTaints: []string{"workaround for bug in platform firmware applied (tainted bit 11)"},
		},

		// Bit 18
		{
			taintsBitMask:  1 << 18,
			expectedTaints: []string{"an in-kernel test has been run (tainted bit 18)"},
		},

		// Bits 0 and 18
		{
			taintsBitMask: (1 << 0) | (1 << 18),
			expectedTaints: []string{"proprietary module was loaded (tainted bit 0)",
				"an in-kernel test has been run (tainted bit 18)"},
		},

		// Bits 0, 24 and 30
		{
			taintsBitMask: (1 << 0) | (1 << 24) | (1 << 30),
			expectedTaints: []string{"proprietary module was loaded (tainted bit 0)",
				"reserved (tainted bit 24)",
				"BPF syscall has either been configured or enabled for unprivileged users/programs (tainted bit 30)"},
		},

		// RH's bit 29
		{
			taintsBitMask:  1 << 29,
			expectedTaints: []string{"Red Hat extension: Technology Preview code was loaded; cf. Technology Preview features support scope description. Refer to \"TECH PREVIEW:\" kernel log entry for details (tainted bit 29)"},
		},

		// RH's reserved bit 31
		{
			taintsBitMask:  1 << 31,
			expectedTaints: []string{"BPF syscall has either been configured or enabled for unprivileged users/programs (tainted bit 31)"},
		},
	}

	for _, tc := range tcs {
		taints := DecodeKernelTaintsFromBitMask(tc.taintsBitMask)
		assert.Equal(t, tc.expectedTaints, taints)
	}
}

func TestDecodeKernelTaintsFromLetters(t *testing.T) {
	testCases := []struct {
		letters           string
		expectedTaintBits []string
	}{
		{
			letters:           "G",
			expectedTaintBits: []string{"proprietary module was loaded (taint letter:G, bit:0)"},
		},
		{
			letters:           "E",
			expectedTaintBits: []string{"unsigned module was loaded (taint letter:E, bit:13)"},
		},
		{
			letters: "OX",
			expectedTaintBits: []string{"externally-built (\"out-of-tree\") module was loaded (taint letter:O, bit:12)",
				"auxiliary taint, defined for and used by distros (taint letter:X, bit:16)"},
		},
		// Unknown letter
		{
			letters:           "n",
			expectedTaintBits: []string{"unknown taint (letter n)"},
		},
	}

	for _, tc := range testCases {
		bits := DecodeKernelTaintsFromLetters(tc.letters)
		assert.Equal(t, tc.expectedTaintBits, bits)
	}
}

func TestGetBitPosFromLetter(t *testing.T) {
	testCases := []struct {
		letter        string
		expectedPos   int
		expectedError string
	}{
		{
			letter:      "G",
			expectedPos: 0,
		},
		{
			letter:      "E",
			expectedPos: 13,
		},
		{
			letter:      "O",
			expectedPos: 12,
		},
		{
			letter:        "OE",
			expectedError: "input string must contain one letter",
		},
		{
			letter:        "",
			expectedError: "input string must contain one letter",
		},
	}

	for _, tc := range testCases {
		bitPos, err := getBitPosFromLetter(tc.letter)
		if err != nil {
			assert.Equal(t, tc.expectedError, err.Error())
		} else {
			assert.Equal(t, tc.expectedError, "")
		}
		assert.Equal(t, tc.expectedPos, bitPos)
	}
}

func TestGetKernelTaintsMask(t *testing.T) {
	testCases := []struct {
		runCommandOutput   string
		runCommandError    error
		expectedTaintsMask uint64
		expectedErrorMsg   string
	}{
		{
			runCommandOutput:   "0",
			runCommandError:    nil,
			expectedTaintsMask: 0,
		},
		{
			runCommandOutput:   "0\n",
			runCommandError:    nil,
			expectedTaintsMask: 0,
		},
		{
			runCommandOutput:   "0\r\t",
			runCommandError:    nil,
			expectedTaintsMask: 0,
		},
		{
			runCommandOutput:   "1024",
			runCommandError:    nil,
			expectedTaintsMask: 1024,
		},
		{
			runCommandOutput:   "65536",
			runCommandError:    nil,
			expectedTaintsMask: 65536,
		},
		{
			runCommandOutput:   "test1",
			runCommandError:    errors.New("this is an error"),
			expectedTaintsMask: 0,
			expectedErrorMsg:   "this is an error",
		},
		{
			runCommandOutput:   "-1",
			runCommandError:    nil,
			expectedTaintsMask: 0,
			expectedErrorMsg:   "failed to decode taints mask \"-1\": strconv.ParseUint: parsing \"-1\": invalid syntax",
		},
	}

	for _, tc := range testCases {
		origFunc := runCommand
		runCommand = func(ctx *clientsholder.Context, cmd string) (string, error) {
			return tc.runCommandOutput, tc.runCommandError
		}
		nt := NewNodeTaintedTester(nil, "fake-node-name")
		result, err := nt.GetKernelTaintsMask()
		assert.Equal(t, tc.expectedTaintsMask, result)
		if err != nil {
			assert.Equal(t, tc.expectedErrorMsg, err.Error())
		} else {
			assert.Equal(t, tc.expectedErrorMsg, "")
		}

		runCommand = origFunc
	}
}

func TestGetAllTainterModules(t *testing.T) {
	testCases := []struct {
		runCommandOutput string
		runCommandError  error
		expectedTainters map[string]string
		expectedErrorMsg string
	}{
		{
			runCommandOutput: "module1 O",
			expectedTainters: map[string]string{"module1": "O"},
		},
		{
			runCommandOutput: "module1 O\nmodule2 E",
			expectedTainters: map[string]string{"module1": "O", "module2": "E"},
		},
		{
			runCommandOutput: "module1 OE\nmodule2 E",
			expectedTainters: map[string]string{"module1": "OE", "module2": "E"},
		},
		{
			runCommandOutput: "\n",
			expectedTainters: map[string]string{},
		},
		{
			runCommandOutput: "",
			expectedTainters: map[string]string{},
		},
		{
			runCommandOutput: "module1",
			expectedErrorMsg: "failed to parse line \"module1\" (output=module1)",
		},
		{
			runCommandOutput: "moduleAppearsTwice E\nmoduleAppearsTwice O",
			expectedErrorMsg: "module moduleAppearsTwice (taints O) has already been parsed (taints E)",
		},
		{
			runCommandError:  errors.New("fake error running command in container"),
			expectedErrorMsg: "failed to run command: fake error running command in container",
		},
	}

	for _, tc := range testCases {
		origFunc := runCommand
		runCommand = func(ctx *clientsholder.Context, cmd string) (string, error) {
			return tc.runCommandOutput, tc.runCommandError
		}
		nt := NewNodeTaintedTester(nil, "fake-node-name")
		tainters, err := nt.getAllTainterModules()
		if err != nil {
			assert.Equal(t, tc.expectedErrorMsg, err.Error())
		} else {
			assert.Equal(t, tc.expectedErrorMsg, "")
		}
		assert.Equal(t, tc.expectedTainters, tainters)

		runCommand = origFunc
	}
}

func TestGetTainterModules(t *testing.T) {
	testCases := []struct {
		runCommandOutput  string
		allowList         map[string]bool
		expectedTainters  map[string]string
		expectedTaintBits map[int]bool
		expectedErrorMsg  string
	}{
		{
			runCommandOutput:  "module1 O",
			expectedTainters:  map[string]string{"module1": "O"},
			expectedTaintBits: map[int]bool{12: true},
		},
		{
			runCommandOutput:  "module1 O\nmodule2 E\n",
			expectedTainters:  map[string]string{"module1": "O", "module2": "E"},
			expectedTaintBits: map[int]bool{12: true, 13: true},
		},
		{
			runCommandOutput:  "module1 OE\nmodule2 O\nmodule3 E",
			expectedTainters:  map[string]string{"module1": "OE", "module2": "O", "module3": "E"},
			expectedTaintBits: map[int]bool{12: true, 13: true},
		},
		{
			runCommandOutput:  "module1 OE\nmodule2 O\nmodule3 E",
			expectedTainters:  map[string]string{"module1": "OE", "module2": "O", "module3": "E"},
			expectedTaintBits: map[int]bool{12: true, 13: true},
		},
		// Allowlist usage 1
		{
			runCommandOutput:  "module1 OE\nmodule2 O\nmodule3 E",
			allowList:         map[string]bool{"module2": true},
			expectedTainters:  map[string]string{"module1": "OE", "module3": "E"},
			expectedTaintBits: map[int]bool{12: true, 13: true},
		},
		// Allowlist usage 2
		{
			runCommandOutput:  "module2 O\nmodule3 E",
			allowList:         map[string]bool{"module2": true, "module3": true},
			expectedTainters:  map[string]string{},
			expectedTaintBits: map[int]bool{12: true, 13: true},
		},
		{
			runCommandOutput:  "",
			expectedTainters:  map[string]string{},
			expectedTaintBits: map[int]bool{},
		},
		{
			runCommandOutput:  "\n",
			expectedTainters:  map[string]string{},
			expectedTaintBits: map[int]bool{},
		},
		// Error checking
		{
			runCommandOutput: "module1",
			expectedErrorMsg: "failed to get tainter modules: failed to parse line \"module1\" (output=module1)",
		},
		{
			runCommandOutput: "module1 E\nmodule2 J",
			expectedErrorMsg: "failed to get taint bits by modules: module module2 has invalid taint letter J: letter J does not belong to any known kernel taint",
		},
	}

	for _, tc := range testCases {
		origFunc := runCommand
		runCommand = func(ctx *clientsholder.Context, cmd string) (string, error) {
			// Make the command to never return error.
			return tc.runCommandOutput, nil
		}
		nt := NewNodeTaintedTester(nil, "fake-node-name")
		tainters, taintBitsByAllModules, err := nt.GetTainterModules(tc.allowList)
		if err != nil {
			assert.Equal(t, tc.expectedErrorMsg, err.Error())
		} else {
			assert.Equal(t, tc.expectedErrorMsg, "")
		}
		assert.Equal(t, tc.expectedTainters, tainters)
		assert.Equal(t, tc.expectedTaintBits, taintBitsByAllModules)

		runCommand = origFunc
	}
}

func TestRemoveAllExceptNumbers(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "module was loaded (taint letter:O, bit:12)",
			expected: "12",
		},
		{
			input:    "123",
			expected: "123",
		},
		{
			input:    "123abc",
			expected: "123",
		},
		{
			input:    "abc123",
			expected: "123",
		},
		{
			input:    "abc123abc",
			expected: "123",
		},
		{
			input:    "abc",
			expected: "",
		},
		{
			input:    "123abc123",
			expected: "123123",
		},
		{
			input:    "123abc123abc",
			expected: "123123",
		},
		{
			input:    "abc123abc123abc",
			expected: "123123",
		},
		{
			input:    "abc123abc123abc123",
			expected: "123123123",
		},
	}

	for _, tc := range testCases {
		result := RemoveAllExceptNumbers(tc.input)
		assert.Equal(t, tc.expected, result)
	}
}

func TestGetTaintedBitsByModules(t *testing.T) {
	testCases := []struct {
		modules           map[string]string
		expectedTaintBits map[int]bool
		expectedError     string
	}{
		// Bit 0 (G)
		{
			modules:           map[string]string{"module1": "G"},
			expectedTaintBits: map[int]bool{0: true},
		},
		{
			modules:           map[string]string{"module1": "P"},
			expectedTaintBits: map[int]bool{0: true},
		},
		// Bit 12 (O)
		{
			modules:           map[string]string{"module1": "O"},
			expectedTaintBits: map[int]bool{12: true},
		},
		// Bits 0 & 12
		{
			modules:           map[string]string{"module1": "GO"},
			expectedTaintBits: map[int]bool{0: true, 12: true},
		},
		// Bits 0 & 12 from two modules.
		{
			modules:           map[string]string{"module1": "GO", "module2": "O"},
			expectedTaintBits: map[int]bool{0: true, 12: true},
		},
		// Bits 0 & 12 from two modules, plus bit 15 (K)
		{
			modules:           map[string]string{"module1": "GO", "module2": "O", "module3": "K"},
			expectedTaintBits: map[int]bool{0: true, 12: true, 15: true},
		},
		// Unknown letter.
		{
			modules:           map[string]string{"module1": "n"},
			expectedTaintBits: nil,
			expectedError:     "module module1 has invalid taint letter n: letter n does not belong to any known kernel taint",
		},
		// RH's tech preview bit 29 (T)
		{
			modules:           map[string]string{"rhModule": "H"},
			expectedTaintBits: map[int]bool{28: true},
		},
	}

	for _, tc := range testCases {
		bits, err := GetTaintedBitsByModules(tc.modules)
		if err != nil {
			assert.Equal(t, tc.expectedError, err.Error())
		} else {
			assert.Equal(t, tc.expectedError, "")
		}
		assert.Equal(t, tc.expectedTaintBits, bits)
	}
}

func TestGetOtherTaintedBits(t *testing.T) {
	testCases := []struct {
		taintsMask           uint64
		taintedBitsByModules map[int]bool
		expectedBits         []int
	}{
		{
			taintsMask:           0,
			taintedBitsByModules: map[int]bool{},
			expectedBits:         []int{},
		},
		{
			taintsMask:           1 << 0,
			taintedBitsByModules: map[int]bool{0: true},
			expectedBits:         []int{},
		},
		// Bits tainted by modules: 0
		// Bits not tainted by modules: 1
		{
			taintsMask:           (1 << 0) | (1 << 1),
			taintedBitsByModules: map[int]bool{0: true},
			expectedBits:         []int{1},
		},
		// Bits tainted by modules: 0, 1
		// Bits not tainted by modules: 2
		{
			taintsMask:           (1 << 0) | (1 << 1) | (1 << 2),
			taintedBitsByModules: map[int]bool{0: true, 1: true},
			expectedBits:         []int{2},
		},
		// Bits tainted by modules: none
		// Bits not tainted by modules: 0, 1, 2
		{
			taintsMask:           (1 << 0) | (1 << 1) | (1 << 2),
			taintedBitsByModules: map[int]bool{},
			expectedBits:         []int{0, 1, 2},
		},
	}

	for _, tc := range testCases {
		bits := GetOtherTaintedBits(tc.taintsMask, tc.taintedBitsByModules)
		assert.Equal(t, tc.expectedBits, bits)
	}
}
