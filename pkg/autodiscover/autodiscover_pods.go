// Copyright (C) 2020-2024 Red Hat, Inc.
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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// findPodsMatchingAtLeastOneLabel retrieves a list of pods that match at least one of the provided labels in the given namespace.
//
// It accepts a CoreV1 client interface, a slice of label objects representing key-value pairs,
// and a namespace string. The function queries the Kubernetes API for pods in that namespace
// whose labels satisfy any of the specified label criteria. The result is returned as a pointer
// to corev1.PodList. If no matching pods are found or an error occurs during listing, the
// function logs the issue using the debug logger and returns nil.
func findPodsMatchingAtLeastOneLabel(oc corev1client.CoreV1Interface, labels []labelObject, namespace string) *corev1.PodList {
	allPods := &corev1.PodList{}
	for _, l := range labels {
		log.Debug("Searching Pods in namespace %s with label %q", namespace, l)
		pods, err := oc.Pods(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: l.LabelKey + "=" + l.LabelValue,
		})
		if err != nil {
			log.Error("Error when listing pods in ns=%s label=%s, err: %v", namespace, l.LabelKey+"="+l.LabelValue, err)
			continue
		}
		allPods.Items = append(allPods.Items, pods.Items...)
	}
	return allPods
}

// FindPodsByLabels retrieves Kubernetes pods that match any of the provided label selectors.
//
// It accepts a CoreV1Interface client, a slice of label objects defining key/value pairs,
// and an optional list of namespace names to restrict the search.
// The function returns all Pods found in the specified namespaces that contain at least one
// matching label. If no namespaces are supplied, it searches across all namespaces.
func FindPodsByLabels(oc corev1client.CoreV1Interface, labels []labelObject, namespaces []string) (runningPods, allPods []corev1.Pod) {
	runningPods = []corev1.Pod{}
	allPods = []corev1.Pod{}
	// Iterate through namespaces
	for _, ns := range namespaces {
		var pods *corev1.PodList
		if len(labels) > 0 {
			pods = findPodsMatchingAtLeastOneLabel(oc, labels, ns)
		} else {
			// If labels are not provided in the namespace under test, they are tested by the CNF suite
			log.Debug("Searching Pods in namespace %s without label", ns)
			var err error
			pods, err = oc.Pods(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Error("Error when listing pods in ns=%s, err: %v", ns, err)
				continue
			}
		}
		// Filter out any pod set to be deleted
		for i := 0; i < len(pods.Items); i++ {
			if pods.Items[i].DeletionTimestamp == nil && pods.Items[i].Status.Phase == corev1.PodRunning {
				runningPods = append(runningPods, pods.Items[i])
			}
			allPods = append(allPods, pods.Items[i])
		}
	}

	return runningPods, allPods
}

// CountPodsByStatus aggregates pods by their phase and returns a count per status.
//
// It iterates over the provided slice of Pod objects, extracts each pod's Phase
// field from its Status, and tallies how many pods are in each phase.
// The returned map keys are the string representations of the pod phases,
// such as "Pending", "Running", "Succeeded", etc., with integer values indicating
// the number of pods in that state. If the input slice is empty, it returns an
// empty map.
func CountPodsByStatus(allPods []corev1.Pod) map[string]int {
	podStates := map[string]int{
		"ready":     0,
		"non-ready": 0,
	}

	for i := range allPods {
		if allPods[i].Status.Phase == corev1.PodRunning {
			podStates["ready"]++
		} else {
			podStates["non-ready"]++
		}
	}

	return podStates
}
