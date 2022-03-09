package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1core "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func buildTestObjects() []runtime.Object {
	// ClusterRoleBinding Objects
	testCRB1 := v1rbac.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "testNS",
			Name:      "testCRB",
		},
	}
	testSA1 := v1core.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "testNS",
			Name:      "testCR1",
		},
	}
	testCR1 := v1rbac.ClusterRole{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "testNS",
			Name:      "testCR",
		},
	}

	// RoleBinding Objects
	testRB2 := v1rbac.RoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name:      "testRB",
			Namespace: "testNS",
		},
	}
	testSA2 := v1core.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "testNS",
			Name:      "testCR2",
		},
	}
	testCR2 := v1rbac.Role{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "testNS",
			Name:      "testRole",
		},
	}
	var testRuntimeObjects []runtime.Object
	testRuntimeObjects = append(testRuntimeObjects, &testCRB1, &testSA1, &testCR1, &testRB2, &testSA2, &testCR2)
	return testRuntimeObjects
}

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
