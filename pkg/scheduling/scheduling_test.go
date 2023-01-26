package scheduling

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

// Monkey patching is used here
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
