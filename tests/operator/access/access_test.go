package access

import (
	"testing"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestPermissionsHaveBadRule(t *testing.T) {
	generateSDP := func(apiGroups []string, resources []string) v1alpha1.StrategyDeploymentPermissions {
		return v1alpha1.StrategyDeploymentPermissions{
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: apiGroups,
					Resources: resources,
				},
			},
		}
	}

	testCases := []struct {
		testClusterPermissions []v1alpha1.StrategyDeploymentPermissions
		expectedResult         bool
	}{
		{ // SCC granted - this is a bad rule
			testClusterPermissions: []v1alpha1.StrategyDeploymentPermissions{
				generateSDP([]string{"security.openshift.io"}, []string{"*"}),
			},
			expectedResult: true,
		},
		{ // SCC granted - this is a bad rule
			testClusterPermissions: []v1alpha1.StrategyDeploymentPermissions{
				generateSDP([]string{"security.openshift.io"}, []string{"securitycontextconstraints"}),
			},
			expectedResult: true,
		},
		{ // SCC granted - this is a bad rule
			testClusterPermissions: []v1alpha1.StrategyDeploymentPermissions{
				generateSDP([]string{"security.openshift.io"}, []string{"*"}),
				generateSDP([]string{"security.openshift.io"}, []string{"securitycontextconstraints"}),
			},
			expectedResult: true,
		},
		{ // No bad rule
			testClusterPermissions: []v1alpha1.StrategyDeploymentPermissions{
				generateSDP([]string{"security.heathytest.io"}, []string{"*"}),
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, PermissionsHaveBadRule(tc.testClusterPermissions))
	}
}
