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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
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
		taints := DecodeKernelTaints(tc.taintsBitMask)
		assert.Equal(t, tc.expectedTaints, taints)
	}
}

func TestGetOutOfTreeModules(t *testing.T) {
	testCases := []struct {
		testModules            []string
		expectedTaintedModules []string
		runCommandOutput       string
	}{
		{ // output is O
			testModules:            []string{"module1"},
			expectedTaintedModules: []string{"module1"},
			runCommandOutput:       "O", // O means out-of-tree
		},
		{ // output is 1 (could be anything)
			testModules:            []string{"module2"},
			expectedTaintedModules: []string{},
			runCommandOutput:       "1",
		},
	}

	for _, tc := range testCases {
		origFunc := runCommand
		runCommand = func(ctx *clientsholder.Context, cmd string) (string, error) {
			return tc.runCommandOutput, nil
		}
		nt := NewNodeTaintedTester(nil)
		assert.Equal(t, tc.expectedTaintedModules, nt.GetOutOfTreeModules(tc.testModules))
		runCommand = origFunc
	}
}

//nolint:funlen
func TestGetKernelTaintInfo(t *testing.T) {
	testCases := []struct {
		runCommandOutput   string
		runCommandError    error
		expectedTaintsMask uint64
		expectedError      error
	}{
		{
			runCommandOutput:   "0",
			runCommandError:    nil,
			expectedTaintsMask: 0,
			expectedError:      nil,
		},
		{
			runCommandOutput:   "0\n",
			runCommandError:    nil,
			expectedTaintsMask: 0,
			expectedError:      nil,
		},
		{
			runCommandOutput:   "0\r\t",
			runCommandError:    nil,
			expectedTaintsMask: 0,
			expectedError:      nil,
		},
		{
			runCommandOutput:   "1024",
			runCommandError:    nil,
			expectedTaintsMask: 1024,
			expectedError:      nil,
		},
		{
			runCommandOutput:   "65536",
			runCommandError:    nil,
			expectedTaintsMask: 65536,
			expectedError:      nil,
		},
		{
			runCommandOutput:   "test1",
			runCommandError:    errors.New("this is an error"),
			expectedTaintsMask: 0,
			expectedError:      errors.New("this is an error"),
		},
	}

	for _, tc := range testCases {
		origFunc := runCommand
		runCommand = func(ctx *clientsholder.Context, cmd string) (string, error) {
			return tc.runCommandOutput, tc.runCommandError
		}
		nt := NewNodeTaintedTester(nil)
		result, err := nt.GetKernelTaintsMask()
		assert.Equal(t, tc.expectedTaintsMask, result)
		assert.Equal(t, err, tc.expectedError)
		runCommand = origFunc
	}
}

func TestGetModulesFromNode(t *testing.T) {
	testCases := []struct {
		runCommandOutput string
		runCommandError  error
		expectedOutput   []string
	}{
		{
			runCommandOutput: "module1\nmodule2\nmodule3",
			runCommandError:  nil,
			expectedOutput:   []string{"module1", "module2", "module3"},
		},
		{
			runCommandOutput: "\tmodule1\nmodule2",
			runCommandError:  nil,
			expectedOutput:   []string{"module1", "module2"},
		},
	}

	for _, tc := range testCases {
		origFunc := runCommand
		runCommand = func(ctx *clientsholder.Context, cmd string) (string, error) {
			return tc.runCommandOutput, tc.runCommandError
		}
		nt := NewNodeTaintedTester(nil)
		assert.Equal(t, tc.expectedOutput, nt.GetModulesFromNode())

		runCommand = origFunc
	}
}

func TestGetTainterModules(t *testing.T) {
	testCases := []struct {
		runCommandOutput string
		runCommandError  error
		expectedOutput   map[string]string
	}{
		{
			runCommandOutput: "module1 O",
			runCommandError:  nil,
			expectedOutput:   map[string]string{"module1": "O"},
		},
		{
			runCommandOutput: "module1 O\nmodule2 E\n",
			runCommandError:  nil,
			expectedOutput:   map[string]string{"module1": "O", "module2": "E"},
		},
		{
			runCommandOutput: "module1 OE\nmodule2 O\nmodule3 E",
			runCommandError:  nil,
			expectedOutput:   map[string]string{"module1": "OE", "module2": "O", "module3": "E"},
		},
		{
			runCommandOutput: "",
			runCommandError:  nil,
			expectedOutput:   map[string]string{},
		},
		{
			runCommandOutput: "\n",
			runCommandError:  nil,
			expectedOutput:   map[string]string{},
		},
	}

	for _, tc := range testCases {
		origFunc := runCommand
		runCommand = func(ctx *clientsholder.Context, cmd string) (string, error) {
			return tc.runCommandOutput, tc.runCommandError
		}
		nt := NewNodeTaintedTester(nil)
		tainters, err := nt.GetTainterModules()
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedOutput, tainters)

		runCommand = origFunc
	}
}
