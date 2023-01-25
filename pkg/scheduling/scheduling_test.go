package scheduling

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			expectedPriority: 0,
			expectedError:    fmt.Errorf("error in parsing chrt: failed to get pid 2396136's policy: No such process"),
		},
	}

	for _, tc := range testCases {
		policy, priority, err := GetSchedulingPolicyAndPriority(tc.outputString)
		assert.Equal(t, policy, tc.expectedPolicy)
		assert.Equal(t, priority, tc.expectedPriority)
		assert.Equal(t, err, tc.expectedError)
	}
}
