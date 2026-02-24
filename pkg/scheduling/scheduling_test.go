// Copyright (C) 2023-2024 Red Hat, Inc.
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
	"reflect"
	"strings"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestProcessPidsCPUScheduling(t *testing.T) {
	testPids := []*crclient.Process{
		{PidNs: 2, Pid: 101, Args: "tbd command line"},
		{PidNs: 3, Pid: 102, Args: "tbd command line"},
	}
	testContainer := &provider.Container{}
	testContainer.Container = &corev1.Container{}

	testCases := []struct {
		mockGetProcessCPUScheduling func(int, *provider.Container) (string, int, error)
		check                       string
		compliant, nonCompliant     []testhelper.ReportObject
	}{
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_OTHER", 0, nil
			},
			check:     SharedCPUScheduling + "1",
			compliant: []testhelper.ReportObject{},

			nonCompliant: []testhelper.ReportObject{
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_OTHER",
						"0",
					},
				},
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_OTHER",
						"0",
					},
				},
			},
		},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_RR", 90, nil
			},
			check:     SharedCPUScheduling + "2",
			compliant: []testhelper.ReportObject{},

			nonCompliant: []testhelper.ReportObject{
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_RR",
						"90",
					},
				},
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_RR",
						"90",
					},
				},
			}},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_FIFO", 9, nil
			},
			check:     ExclusiveCPUScheduling + "1",
			compliant: []testhelper.ReportObject{},

			nonCompliant: []testhelper.ReportObject{
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_FIFO",
						"9",
					},
				},
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_FIFO",
						"9",
					},
				},
			}},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_FIFO", 11, nil
			},
			check: ExclusiveCPUScheduling + "2",
			nonCompliant: []testhelper.ReportObject{
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_FIFO",
						"11",
					},
				},
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_FIFO",
						"11",
					},
				},
			},

			compliant: []testhelper.ReportObject{}},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_FIFO", 50, nil
			},
			check:     IsolatedCPUScheduling + "1",
			compliant: []testhelper.ReportObject{},

			nonCompliant: []testhelper.ReportObject{
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_FIFO",
						"50",
					},
				},
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_FIFO",
						"50",
					},
				},
			}},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_RR", 99, nil
			},
			check:     IsolatedCPUScheduling + "2",
			compliant: []testhelper.ReportObject{},

			nonCompliant: []testhelper.ReportObject{
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_RR",
						"99",
					},
				},
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_RR",
						"99",
					},
				},
			}},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				return "SCHED_OTHER", 0, nil
			},
			check: IsolatedCPUScheduling + "3",
			nonCompliant: []testhelper.ReportObject{
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_OTHER",
						"0",
					},
				},
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_OTHER",
						"0",
					},
				},
			},

			compliant: []testhelper.ReportObject{}},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				if pid == 101 {
					return "", InvalidPriority, fmt.Errorf("command failed: %s", NoProcessFoundErrMsg)
				}
				return "SCHED_OTHER", 0, nil
			},
			check: SharedCPUScheduling + "_disappeared",
			compliant: []testhelper.ReportObject{
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process disappeared",
						"",
						"",
						"",
						"tbd command line",
						"",
						"",
					},
				},
			},
			nonCompliant: []testhelper.ReportObject{
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_OTHER",
						"0",
					},
				},
			},
		},
		{
			mockGetProcessCPUScheduling: func(pid int, container *provider.Container) (string, int, error) {
				if pid == 101 {
					return "", InvalidPriority, fmt.Errorf("connection refused")
				}
				return "SCHED_OTHER", 0, nil
			},
			check:     SharedCPUScheduling + "_other_error",
			compliant: []testhelper.ReportObject{},
			nonCompliant: []testhelper.ReportObject{
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"could not determine scheduling policy",
						"",
						"",
						"",
						"tbd command line",
						"",
						"",
					},
				},
				{
					ObjectType: "ContainerProcess",
					ObjectFieldsKeys: []string{
						testhelper.ReasonForNonCompliance,
						"Namespace",
						testhelper.PodName,
						testhelper.ContainerName,
						testhelper.ProcessCommandLine,
						testhelper.SchedulingPolicy,
						testhelper.SchedulingPriority,
					},
					ObjectFieldsValues: []string{
						"process does not satisfy: ",
						"",
						"",
						"",
						"tbd command line",
						"SCHED_OTHER",
						"0",
					},
				},
			},
		},
	}
	var logArchive strings.Builder
	log.SetupLogger(&logArchive, "INFO")
	for _, tc := range testCases {
		GetProcessCPUSchedulingFn = tc.mockGetProcessCPUScheduling
		compliant, nonCompliant := ProcessPidsCPUScheduling(testPids, testContainer, tc.check, log.GetLogger())

		fmt.Printf(
			"test=%s Actual compliant=%s,\n",
			tc.check,
			testhelper.ReportObjectTestString(compliant),
		)
		fmt.Printf(
			"test=%s Actual non-compliant=%s,\nfail=%v",
			tc.check,
			testhelper.ReportObjectTestString(nonCompliant),
			len(compliant) != len(tc.compliant),
		)

		assert.Equal(t, len(compliant), len(tc.compliant))
		assert.Equal(t, len(nonCompliant), len(tc.nonCompliant))
		for i := range compliant {
			fmt.Printf(
				"compliant test=%s fail=%v",
				tc.check,
				!reflect.DeepEqual(tc.compliant[i], *compliant[i]),
			)
			assert.Equal(t, tc.compliant[i], *compliant[i])
		}
		for i := range nonCompliant {
			fmt.Printf(
				"non compliant test=%s fail=%v",
				tc.check,
				!reflect.DeepEqual(tc.nonCompliant[i], *nonCompliant[i]),
			)
			assert.Equal(t, tc.nonCompliant[i], *nonCompliant[i])
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
			expectedError: fmt.Errorf(
				"invalid: chrt: failed to get pid 2396136's policy: No such process",
			),
		},
	}

	for _, tc := range testCases {
		policy, priority, err := parseSchedulingPolicyAndPriority(tc.outputString)
		assert.Equal(t, tc.expectedPolicy, policy)
		assert.Equal(t, tc.expectedPriority, priority)
		assert.Equal(t, tc.expectedError, err)
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
