// Copyright (C) 2022-2024 Red Hat, Inc.
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

package provider

import (
	"fmt"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover"
	appsv1 "k8s.io/api/apps/v1"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
)

// StatefulSet wraps an appsv1.StatefulSet and provides utility methods.
//
// It embeds a pointer to appsv1.StatefulSet, allowing direct access to the
// underlying Kubernetes object while adding convenience functions such as
// IsStatefulSetReady to check readiness status and ToString for a formatted
// string representation of key fields.
type StatefulSet struct {
	*appsv1.StatefulSet
}

// IsStatefulSetReady reports whether the StatefulSet has all replicas ready.
//
// It checks the status of the StatefulSet and returns true if every replica is in a ready state.
// The function performs no arguments; it relies on the receiver's internal state to determine readiness.
func (ss *StatefulSet) IsStatefulSetReady() bool {
	var replicas int32
	if ss.Spec.Replicas != nil {
		replicas = *(ss.Spec.Replicas)
	} else {
		replicas = 1
	}
	if ss.Status.ReadyReplicas != replicas ||
		ss.Status.CurrentReplicas != replicas ||
		ss.Status.UpdatedReplicas != replicas {
		return false
	}
	return true
}

// ToString returns a human readable description of the StatefulSet.
//
// It formats key attributes such as name, namespace, replica count,
// pod template labels, and container names into a single string.
// The resulting string is useful for logging or debugging purposes.
func (ss *StatefulSet) ToString() string {
	return fmt.Sprintf("statefulset: %s ns: %s",
		ss.Name,
		ss.Namespace,
	)
}

// GetUpdatedStatefulset retrieves an up-to-date StatefulSet object from the cluster.
//
// It accepts a Kubernetes AppsV1Interface client, the namespace of the
// desired StatefulSet, and its name. The function internally calls
// FindStatefulsetByNameByNamespace to locate the resource and returns
// a pointer to the resulting StatefulSet along with any error encountered.
// If the StatefulSet is not found or an API call fails, it returns nil for
// the StatefulSet and an appropriate error.
func GetUpdatedStatefulset(ac appv1client.AppsV1Interface, namespace, name string) (*StatefulSet, error) {
	result, err := autodiscover.FindStatefulsetByNameByNamespace(ac, namespace, name)
	return &StatefulSet{
		result,
	}, err
}
