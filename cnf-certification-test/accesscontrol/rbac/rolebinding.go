// Copyright (C) 2022 Red Hat, Inc.
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

package rbac

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetRoleBindings returns any role bindings extracted from the desired pod.
func GetRoleBindings(podNamespace, serviceAccountName string) ([]string, error) {
	// Get all of the rolebindings from all namespaces.
	clientsHolder := clientsholder.GetClientsHolder()
	roleList, roleErr := clientsHolder.K8sClient.RbacV1().Roles("").List(context.TODO(), metav1.ListOptions{})
	if roleErr != nil {
		logrus.Errorf("executing rolebinding command failed with error: %v", roleErr)
		return nil, roleErr
	}

	rolebindings := []string{}
	for index := range roleList.Items {
		// Determine if the role causes a failure of the test.
		if roleOutOfNamespace(roleList.Items[index].Namespace, podNamespace, roleList.Items[index].Name, serviceAccountName) {
			rolebindings = append(rolebindings, roleList.Items[index].Namespace+":"+roleList.Items[index].Name)
		}
	}
	return rolebindings, nil
}
