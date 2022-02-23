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

package rolebinding

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleBinding holds information derived from running "oc get rolebindings" on the command line.
type RoleBinding struct {
	podNamespace       string
	serviceAccountName string
	// roleBindings []string // Output variable that stores the 'bad' RoleBindings
	ClientHolder *clientsholder.ClientsHolder
}

// NewRoleBinding creates a new RoleBinding object
func NewRoleBinding(serviceAccountName, podNamespace string, ch *clientsholder.ClientsHolder) *RoleBinding {
	// Just as a note, the old test suite ran the following command to help determine service accounts that fell outside of the pod's namespace:
	// oc get rolebindings --all-namespaces -o custom-columns='NAMESPACE:metadata.namespace,NAME:metadata.name,SERVICE_ACCOUNTS:subjects[?(@.kind=="ServiceAccount")]' | grep -E '` + serviceAccountSubString + `|SERVICE_ACCOUNTS'

	return &RoleBinding{
		serviceAccountName: serviceAccountName,
		podNamespace:       podNamespace,
		ClientHolder:       ch,
	}
}

// GetRoleBindings returns any role bindings extracted from the desired pod.
func (rb *RoleBinding) GetRoleBindings() ([]string, error) {
	// Get all of the rolebindings from all namespaces.
	roleList, roleErr := rb.ClientHolder.K8sClient.RbacV1().Roles("").List(context.TODO(), v1.ListOptions{})
	if roleErr != nil {
		logrus.Errorf("executing rolebinding command failed with error: %s", roleErr)
		return nil, roleErr
	}

	rolebindings := []string{}
	for index := range roleList.Items {
		// Determine if the role causes a failure of the test.
		if roleOutOfNamespace(roleList.Items[index].Namespace, rb.podNamespace, roleList.Items[index].Name, rb.serviceAccountName) {
			rolebindings = append(rolebindings, roleList.Items[index].Namespace+":"+roleList.Items[index].Name)
		}
	}
	return rolebindings, nil
}

func roleOutOfNamespace(roleNamespace, podNamespace, roleName, serviceAccountName string) bool {
	// Skip if the role namespace is part of the pod namespace.
	if roleNamespace == podNamespace {
		return false
	}

	// Role is in another namespace and the service account names match.
	if roleName == serviceAccountName {
		return true
	}

	return false
}
