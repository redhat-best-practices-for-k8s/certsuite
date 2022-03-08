package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleOutOfNamespace(t *testing.T) {
	testCases := []struct {
		testRoleNS             string
		testPodNS              string
		testRoleName           string
		testServiceAccountName string
		expectedOutOfNS        bool
	}{
		{ // Test Case #1 - Pod and Role are in the same namespace.
			testRoleNS:             "ns1",
			testPodNS:              "ns1",
			testRoleName:           "sa1",
			testServiceAccountName: "sa1",

			expectedOutOfNS: false,
		},
		{ // Test Case #2 - Pod and Role are in different namespaces.
			testRoleNS:             "ns1",
			testPodNS:              "ns2",
			testRoleName:           "sa1",
			testServiceAccountName: "sa1",

			expectedOutOfNS: true,
		},
		{ // Test Case #3 - Pod, Role names don't match and are in different namespaces.
			testRoleNS:             "ns1",
			testPodNS:              "ns2",
			testRoleName:           "sa1",
			testServiceAccountName: "sa2",

			expectedOutOfNS: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutOfNS, roleOutOfNamespace(tc.testRoleNS, tc.testPodNS, tc.testRoleName, tc.testServiceAccountName))
	}
}
