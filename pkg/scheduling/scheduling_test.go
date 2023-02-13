// Copyright (C) 2023 Red Hat, Inc.
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

package scheduling

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

func TestProcessPidsCPUScheduling(t *testing.T) {
	testPids := []int{101, 102}
	testContainer := &provider.Container{}

	testCases := []struct {
		mockGetProcessCPUScheduling           func(int, *provider.Container) (string, int, error)
		check                                 string
		expectedCPUSchedulingConditionSuccess bool
	}{
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_OTHER", 0, nil
			},
			check:                                 SharedCPUScheduling,
			expectedCPUSchedulingConditionSuccess: true,
		},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_RR", 90, nil
			},
			check:                                 SharedCPUScheduling,
			expectedCPUSchedulingConditionSuccess: false,
		},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_FIFO", 9, nil
			},
			check:                                 ExclusiveCPUScheduling,
			expectedCPUSchedulingConditionSuccess: true,
		},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_FIFO", 11, nil
			},
			check:                                 ExclusiveCPUScheduling,
			expectedCPUSchedulingConditionSuccess: false,
		},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_FIFO", 50, nil
			},
			check:                                 IsolatedCPUScheduling,
			expectedCPUSchedulingConditionSuccess: true,
		},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_RR", 99, nil
			},
			check:                                 IsolatedCPUScheduling,
			expectedCPUSchedulingConditionSuccess: true,
		},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_OTHER", 0, nil
			},
			check:                                 IsolatedCPUScheduling,
			expectedCPUSchedulingConditionSuccess: false,
		},
	}
	for _, tc := range testCases {
		GetProcessCPUSchedulingFn = tc.mockGetProcessCPUScheduling
		isCheckSuccessful := ProcessPidsCPUScheduling(testPids, testContainer, make(map[*provider.Container][]int), tc.check)
		assert.Equal(t, isCheckSuccessful, tc.expectedCPUSchedulingConditionSuccess)
	}
}

func TestGetProcessCPUScheduling(t *testing.T) {
	mockSuccessStdout := `pid 476's current scheduling policy: SCHED_OTHER
	pid 476's current scheduling priority: 0`
	mockErr := fmt.Errorf(`chrt: failed to get pid 476's policy: No such process`)
	container := provider.Container{}
	testPid := 476

	testCases := []struct {
		testContainer                            *provider.Container
		mockCrcClientExecCommandContainerNSEnter func(string, *provider.Container) (string, string, error)
		expectedPolicy                           string
		expectedPriority                         int
		expectedError                            error
	}{
		{
			testContainer: &container,
			mockCrcClientExecCommandContainerNSEnter: func(command string, container *provider.Container) (string, string, error) {
				return mockSuccessStdout, "", nil
			},
			expectedPolicy:   "SCHED_OTHER",
			expectedPriority: 0,
			expectedError:    mockErr,
		},
		{
			testContainer: &container,
			mockCrcClientExecCommandContainerNSEnter: func(command string, container *provider.Container) (string, string, error) {
				return "", "", mockErr
			},
			expectedPolicy:   "",
			expectedPriority: -1,
			expectedError:    mockErr,
		},
	}
	for _, tc := range testCases {
		CrcClientExecCommandContainerNSEnter = tc.mockCrcClientExecCommandContainerNSEnter
		policy, priority, err := GetProcessCPUScheduling(testPid, tc.testContainer)

		assert.Equal(t, policy, tc.expectedPolicy)
		assert.Equal(t, priority, tc.expectedPriority)
		if err != nil {
			assert.Contains(t, err.Error(), tc.expectedError.Error())
		}
	}
}

func TestExistsSchedulingPolicyAndPriority(t *testing.T) {
	testCases := []struct {
		outputString     string
		expectedPolicy   string
		expectedPriority int
		expectedError    error
	}{
		{
			outputString: `pid 476's current scheduling policy: SCHED_OTHER
							pid 476's current scheduling priority: 0`,
			expectedPolicy:   "SCHED_OTHER",
			expectedPriority: 0,
			expectedError:    nil,
		},
		{
			outputString: `pid 476's current scheduling policy: SCHED_FIFO
							pid 476's current scheduling priority: 10`,
			expectedPolicy:   "SCHED_FIFO",
			expectedPriority: 10,
			expectedError:    nil,
		},
		{
			outputString:     `chrt: failed to get pid 2396136's policy: No such process`,
			expectedPolicy:   "",
			expectedPriority: -1,
			expectedError:    fmt.Errorf("invalid: chrt: failed to get pid 2396136's policy: No such process"),
		},
	}

	for _, tc := range testCases {
		policy, priority, err := parseSchedulingPolicyAndPriority(tc.outputString)
		assert.Equal(t, policy, tc.expectedPolicy)
		assert.Equal(t, priority, tc.expectedPriority)
		assert.Equal(t, err, tc.expectedError)
	}
}

func TestPolicyIsRT(t *testing.T) {
	testCases := []struct {
		testPolicy     string
		expectedOutput bool
	}{
		{
			testPolicy:     "SCHED_FIFO",
			expectedOutput: true,
		},
		{
			testPolicy:     "SCHED_RR",
			expectedOutput: true,
		},
		{
			testPolicy:     "",
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, PolicyIsRT(tc.testPolicy))
	}
}
