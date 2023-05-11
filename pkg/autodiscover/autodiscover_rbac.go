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
	"context"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getRoleBindings returns all of the rolebindings in the cluster
func getRoleBindings() ([]rbacv1.RoleBinding, error) {
	// Get all of the rolebindings from all namespaces
	clientsHolder := clientsholder.GetClientsHolder()
	roleList, roleErr := clientsHolder.K8sClient.RbacV1().RoleBindings("").List(context.TODO(), metav1.ListOptions{})
	if roleErr != nil {
		logrus.Errorf("executing rolebinding command failed with error: %v", roleErr)
		return nil, roleErr
	}
	return roleList.Items, nil
}

// getClusterRoleBindings returns all of the clusterrolebindings in the cluster
func getClusterRoleBindings() ([]rbacv1.ClusterRoleBinding, error) {
	// Get all of the clusterrolebindings from the cluster
	// These are not namespaced so we want all of them
	clientsHolder := clientsholder.GetClientsHolder()
	crbList, crbErr := clientsHolder.K8sClient.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if crbErr != nil {
		logrus.Errorf("executing clusterrolebinding command failed with error: %v", crbErr)
		return nil, crbErr
	}
	return crbList.Items, nil
}
