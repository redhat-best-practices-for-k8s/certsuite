// Copyright (C) 2023-2024 Red Hat, Inc.
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
	"context"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1typed "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

// getRoleBindings retrieves all rolebindings across every namespace
//
// This function queries the Kubernetes RBAC API for RoleBinding objects in
// every namespace by using an empty string selector. It returns a slice of
// RoleBinding instances or an error if the list operation fails, logging the
// failure before propagating it.
func getRoleBindings(client rbacv1typed.RbacV1Interface) ([]rbacv1.RoleBinding, error) {
	// Get all of the rolebindings from all namespaces
	roleList, roleErr := client.RoleBindings("").List(context.TODO(), metav1.ListOptions{})
	if roleErr != nil {
		log.Error("Executing rolebinding command failed with error: %v", roleErr)
		return nil, roleErr
	}
	return roleList.Items, nil
}

// getClusterRoleBindings retrieves all cluster‑level role bindings
//
// This function calls the Kubernetes RBAC API to list every ClusterRoleBinding
// in the cluster, ignoring namespaces because they are cluster scoped. It
// returns a slice of the bindings or an error if the request fails, logging any
// failure for debugging purposes.
func getClusterRoleBindings(client rbacv1typed.RbacV1Interface) ([]rbacv1.ClusterRoleBinding, error) {
	// Get all of the clusterrolebindings from the cluster
	// These are not namespaced so we want all of them
	crbList, crbErr := client.ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if crbErr != nil {
		log.Error("Executing clusterrolebinding command failed with error: %v", crbErr)
		return nil, crbErr
	}
	return crbList.Items, nil
}

// getRoles retrieves all cluster roles
//
// The function queries the Kubernetes RBAC API to list every Role resource
// across all namespaces, returning a slice of role objects or an error if the
// request fails. It logs any errors encountered during the API call before
// propagating them to the caller.
func getRoles(client rbacv1typed.RbacV1Interface) ([]rbacv1.Role, error) {
	// Get all of the roles from all namespaces
	roleList, roleErr := client.Roles("").List(context.TODO(), metav1.ListOptions{})
	if roleErr != nil {
		log.Error("Executing roles command failed with error: %v", roleErr)
		return nil, roleErr
	}
	return roleList.Items, nil
}
