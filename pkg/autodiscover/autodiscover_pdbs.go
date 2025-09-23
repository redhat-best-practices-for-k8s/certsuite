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

package autodiscover

import (
	"context"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	policyv1client "k8s.io/client-go/kubernetes/typed/policy/v1"
)

// getPodDisruptionBudgets Collects pod disruption budgets across specified namespaces
//
// The function iterates over a list of namespace names, requesting the pod
// disruption budgets present in each one via the Kubernetes policy client. It
// aggregates all retrieved items into a single slice and returns them along
// with any error that occurs during listing. If an error is encountered for a
// namespace, the function aborts immediately and propagates the error.
func getPodDisruptionBudgets(oc policyv1client.PolicyV1Interface, namespaces []string) ([]policyv1.PodDisruptionBudget, error) {
	podDisruptionBudgets := []policyv1.PodDisruptionBudget{}
	for _, ns := range namespaces {
		pdbs, err := oc.PodDisruptionBudgets(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		podDisruptionBudgets = append(podDisruptionBudgets, pdbs.Items...)
	}

	return podDisruptionBudgets, nil
}
