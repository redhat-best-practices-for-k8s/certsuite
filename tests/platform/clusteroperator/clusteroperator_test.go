package clusteroperator

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsClusterOperatorAvailable(t *testing.T) {
	generateClusterOperator := func(
		availableStatus configv1.ConditionStatus,
		degradedStatus configv1.ConditionStatus,
		progressingStatus configv1.ConditionStatus) *configv1.ClusterOperator {
		return &configv1.ClusterOperator{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-cluster-operator",
			},
			Status: configv1.ClusterOperatorStatus{
				Conditions: []configv1.ClusterOperatorStatusCondition{
					{
						Type:   configv1.OperatorAvailable,
						Status: availableStatus,
					},
					{
						Type:   configv1.OperatorDegraded,
						Status: degradedStatus,
					},
					{
						Type:   configv1.OperatorProgressing,
						Status: progressingStatus,
					},
				},
			},
		}
	}

	testCases := []struct {
		testAvailableStatus   configv1.ConditionStatus
		testDegradedStatus    configv1.ConditionStatus
		testProgressingStatus configv1.ConditionStatus
		expectedResult        bool
	}{
		{
			testAvailableStatus:   configv1.ConditionTrue,
			testDegradedStatus:    configv1.ConditionFalse,
			testProgressingStatus: configv1.ConditionFalse,
			expectedResult:        true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, IsClusterOperatorAvailable(generateClusterOperator(tc.testAvailableStatus, tc.testDegradedStatus, tc.testProgressingStatus)))
	}
}
