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

// StatefulSet Encapsulates a Kubernetes StatefulSet for simplified management
//
// The structure embeds the official StatefulSet type, allowing direct access to
// its fields while providing helper methods. It offers functionality to
// determine if the set is fully ready and to produce a concise string
// representation of its identity.
type StatefulSet struct {
	*appsv1.StatefulSet
}

// StatefulSet.IsStatefulSetReady Checks if all replicas of a StatefulSet are fully operational
//
// The method compares the desired number of replicas, which defaults to one if
// unspecified, against the current status fields: ready, current, and updated
// replicas. If any of these counts differ from the target, it returns false;
// otherwise, true indicates the StatefulSet is considered ready.
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

// StatefulSet.ToString Formats a StatefulSet name and namespace into a string
//
// The method builds a concise representation of the StatefulSet by combining
// its name and namespace. It uses formatting utilities to return a single
// string that identifies the resource in a human‑readable form.
func (ss *StatefulSet) ToString() string {
	return fmt.Sprintf("statefulset: %s ns: %s",
		ss.Name,
		ss.Namespace,
	)
}

// GetUpdatedStatefulset Retrieves the current StatefulSet object for a given namespace and name
//
// This function calls an internal discovery helper to fetch the latest
// statefulset from the Kubernetes API. It wraps the result in a custom
// StatefulSet type that provides additional methods, such as readiness checks.
// The returned pointer is nil if an error occurs, with the error propagated to
// the caller.
func GetUpdatedStatefulset(ac appv1client.AppsV1Interface, namespace, name string) (*StatefulSet, error) {
	result, err := autodiscover.FindStatefulsetByNameByNamespace(ac, namespace, name)
	return &StatefulSet{
		result,
	}, err
}
