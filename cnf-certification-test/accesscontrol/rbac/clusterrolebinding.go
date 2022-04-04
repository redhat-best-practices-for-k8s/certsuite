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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetClusterRoleBindings(serviceAccountName, podNamespace string) ([]string, error) {
	// Get all of the clusterrolebindings from all namespaces.
	clientsHolder := clientsholder.GetClientsHolder()
	crbList, crbErr := clientsHolder.K8sClient.RbacV1().ClusterRoles().List(context.TODO(), v1.ListOptions{})
	if crbErr != nil {
		logrus.Errorf("executing clusterrolebinding command failed with error: %s", crbErr)
		return nil, crbErr
	}

	clusterrolebindings := []string{}
	for index := range crbList.Items {
		// Determine if the role causes a failure of the test.
		if roleOutOfNamespace(crbList.Items[index].Namespace, podNamespace, crbList.Items[index].Name, serviceAccountName) {
			clusterrolebindings = append(clusterrolebindings, crbList.Items[index].Namespace+":"+crbList.Items[index].Name)
		}
	}
	return clusterrolebindings, nil
}
