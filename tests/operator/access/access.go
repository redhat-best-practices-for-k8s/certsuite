package access

import (
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// PermissionsHaveBadRule detects if any RBAC rule grants access to security context constraints
//
// The function iterates over a slice of cluster permissions, examining each
// rule for the presence of the security API group or a wildcard. When such a
// group is found, it then checks whether the rule targets the
// securitycontextconstraints resource or all resources. If any matching rule
// exists, the function returns true to indicate a problematic configuration;
// otherwise it returns false.
func PermissionsHaveBadRule(clusterPermissions []v1alpha1.StrategyDeploymentPermissions) bool {
	badRuleFound := false
	for permissionIndex := range clusterPermissions {
		permission := &clusterPermissions[permissionIndex]
		for ruleIndex := range permission.Rules {
			rule := &permission.Rules[ruleIndex]

			// Check whether the rule is for the security api group.
			securityGroupFound := false
			for _, group := range rule.APIGroups {
				if group == "*" || group == "security.openshift.io" {
					securityGroupFound = true
					break
				}
			}

			if !securityGroupFound {
				continue
			}

			// Now check whether it grants some access to securitycontextconstraint resources.
			for _, resource := range rule.Resources {
				if resource == "*" || resource == "securitycontextconstraints" {
					// Keep reviewing other permissions' rules so we can log all the failing ones in the claim file.
					badRuleFound = true
					break
				}
			}
		}
	}

	return badRuleFound
}
