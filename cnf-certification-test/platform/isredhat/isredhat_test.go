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

package isredhat

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/autodiscover/pkg/clientsholder"
)

func TestIsRHEL(t *testing.T) {
	testCases := []struct {
		testOutput string
		expected   bool
	}{
		{
			testOutput: "Red Hat Enterprise Linux release 8.5 (Ootpa)",
			expected:   true,
		},
		{
			testOutput: "Unknown Base Image",
			expected:   false,
		},
		{
			testOutput: "CentOS",
			expected:   false,
		},
		{
			testOutput: "",
			expected:   false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, IsRHEL(tc.testOutput))
	}
}

func TestTestContainerIsRedHatRelease(t *testing.T) {
	testCases := []struct {
		resultStdOut string
		resultStdErr string
		resultErr    error

		expectedResult bool
		expectedErr    error
	}{
		{ // Test Case #1 - No error, RHEL release
			resultStdOut:   "Red Hat Enterprise Linux release 8.5 (Ootpa)",
			resultStdErr:   "",
			resultErr:      nil,
			expectedResult: true,
			expectedErr:    nil,
		},
		{ // Test Case #2 - Error with exec, RHEL release
			resultStdOut:   "Red Hat Enterprise Linux release 8.5 (Ootpa)",
			resultStdErr:   "",
			resultErr:      errors.New("this is an error"),
			expectedResult: false,
			expectedErr:    errors.New("this is an error"),
		},
		{ // Test Case #3 - Error with stderr, RHEL release
			resultStdOut:   "Red Hat Enterprise Linux release 8.5 (Ootpa)",
			resultStdErr:   "random error",
			resultErr:      nil,
			expectedResult: false,
			expectedErr:    errors.New("random error"),
		},
	}

	for _, tc := range testCases {
		bit := NewBaseImageTester(&clientsholder.CommandMock{
			// Mock out the return values from actually running the command.
			ExecCommandContainerFunc: func(context clientsholder.Context, s string) (string, string, error) {
				return tc.resultStdOut, tc.resultStdErr, tc.resultErr
			}}, clientsholder.Context{
			Namespace:     "testNamespace",
			Podname:       "testPodName",
			Containername: "testContainer",
		})

		result, err := bit.TestContainerIsRedHatRelease()
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedResult, result)
	}
}
