package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleBindingOutOfNamespace(t *testing.T) {
	testCases := []struct {
		testRoleBindingNS      string
		testPodNS              string
		testRoleBindingName    string
		testServiceAccountName string
		expectedOutOfNS        bool
	}{
		{ // Test Case #1 - Pod and RoleBinding are in the same namespace.
			testRoleBindingNS:      "ns1",
			testPodNS:              "ns1",
			testRoleBindingName:    "sa1",
			testServiceAccountName: "sa1",

			expectedOutOfNS: false,
		},
		{ // Test Case #2 - Pod and RoleBinding are in different namespaces.
			testRoleBindingNS:      "ns1",
			testPodNS:              "ns2",
			testRoleBindingName:    "sa1",
			testServiceAccountName: "sa1",

			expectedOutOfNS: true,
		},
		{ // Test Case #3 - Pod, RoleBinding names do not match and are in different namespaces.
			testRoleBindingNS:      "ns1",
			testPodNS:              "ns2",
			testRoleBindingName:    "sa1",
			testServiceAccountName: "sa2",

			expectedOutOfNS: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutOfNS, roleBindingOutOfNamespace(tc.testRoleBindingNS, tc.testPodNS, tc.testRoleBindingName, tc.testServiceAccountName))
	}
}
