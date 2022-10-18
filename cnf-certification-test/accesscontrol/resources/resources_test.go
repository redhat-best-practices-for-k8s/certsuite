package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	validCPULimit    = "1"
	validMemLimit    = "512Mi"
	invalidCPULimit1 = "0.5"
	invalidMemLimit1 = "64.5"
	invalidCPULimit2 = "2"
	invalidMemLimit2 = "65"
)

//nolint:funlen
func TestHasRequestsAndLimitsSet(t *testing.T) {
	testCases := []struct {
		testContainer  *provider.Container
		expectedResult bool
	}{
		{ // Test Case #1 - Happy path, all resource are set
			testContainer: &provider.Container{
				Container: &v1.Container{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"cpu":    resource.MustParse(validCPULimit),
							"memory": resource.MustParse(validMemLimit),
						},
						Limits: v1.ResourceList{
							"cpu":    resource.MustParse(validCPULimit),
							"memory": resource.MustParse(validMemLimit),
						},
					},
				},
			},
			expectedResult: true,
		},
		{ // Test Case #2 - Failure due to missing limits
			testContainer: &provider.Container{
				Container: &v1.Container{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"cpu":    resource.MustParse(validCPULimit),
							"memory": resource.MustParse(validMemLimit),
						},
						Limits: nil,
					},
				},
			},
			expectedResult: false,
		},
		{ // Test Case #3 - Failure due to missing memory limit
			testContainer: &provider.Container{
				Container: &v1.Container{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"cpu":    resource.MustParse(validCPULimit),
							"memory": resource.MustParse(validMemLimit),
						},
						Limits: v1.ResourceList{
							"cpu": resource.MustParse(validCPULimit),
						},
					},
				},
			},
			expectedResult: false,
		},
		{ // Test Case #4 - Failure due to missing resources in general
			testContainer: &provider.Container{
				Container: &v1.Container{},
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, HasRequestsAndLimitsSet(tc.testContainer))
	}
}
