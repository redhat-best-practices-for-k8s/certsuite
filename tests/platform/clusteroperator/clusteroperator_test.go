package clusteroperator

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/ocplite"
	"github.com/stretchr/testify/assert"
)

func TestIsClusterOperatorAvailable(t *testing.T) {
	generateClusterOperator := func(
		availableStatus string,
		degradedStatus string,
		progressingStatus string) *ocplite.ClusterOperator {
		return &ocplite.ClusterOperator{
			Name: "test-cluster-operator",
			Status: ocplite.ClusterOperatorStatus{
				Conditions: []ocplite.ClusterOperatorStatusCondition{
					{
						Type:   ocplite.OperatorAvailable,
						Status: availableStatus,
					},
					{
						Type:   ocplite.OperatorDegraded,
						Status: degradedStatus,
					},
					{
						Type:   ocplite.OperatorProgressing,
						Status: progressingStatus,
					},
				},
			},
		}
	}

	testCases := []struct {
		testAvailableStatus   string
		testDegradedStatus    string
		testProgressingStatus string
		expectedResult        bool
	}{
		{
			testAvailableStatus:   "True",
			testDegradedStatus:    "False",
			testProgressingStatus: "False",
			expectedResult:        true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, IsClusterOperatorAvailable(generateClusterOperator(tc.testAvailableStatus, tc.testDegradedStatus, tc.testProgressingStatus)))
	}
}
