// Copyright (C) 2023 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package autodiscover

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func buildTestObjects() []runtime.Object {
	// ClusterRoleBinding Objects
	testCRB1 := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "testNS",
			Name:      "testCRB",
		},
	}
	testSA1 := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "testNS",
			Name:      "testCR1",
		},
	}
	testCR1 := rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "testNS",
			Name:      "testCR",
		},
	}

	// RoleBinding Objects
	testRB2 := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testRB",
			Namespace: "testNS",
		},
	}
	testSA2 := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "testNS",
			Name:      "testCR2",
		},
	}
	testCR2 := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "testNS",
			Name:      "testRole",
		},
	}
	var testRuntimeObjects []runtime.Object
	testRuntimeObjects = append(testRuntimeObjects, &testCRB1, &testSA1, &testCR1, &testRB2, &testSA2, &testCR2)
	return testRuntimeObjects
}

func TestGetClusterRoleBinding(t *testing.T) {
	client := clientsholder.GetTestClientsHolder(buildTestObjects())
	gatheredCRBs, err := getClusterRoleBindings(client.K8sClient.RbacV1())
	assert.Nil(t, err)
	assert.Equal(t, "testCRB", gatheredCRBs[0].Name)
}

func TestGetRoleBinding(t *testing.T) {
	client := clientsholder.GetTestClientsHolder(buildTestObjects())
	gatheredRBs, err := getRoleBindings(client.K8sClient.RbacV1())
	assert.Nil(t, err)
	assert.Equal(t, "testRB", gatheredRBs[0].Name)
}

func TestGetRoles(t *testing.T) {
	client := clientsholder.GetTestClientsHolder(buildTestObjects())
	gatheredRoles, err := getRoles(client.K8sClient.RbacV1())
	assert.Nil(t, err)
	assert.Equal(t, "testRole", gatheredRoles[0].Name)
}
