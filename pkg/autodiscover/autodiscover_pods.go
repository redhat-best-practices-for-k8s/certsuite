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
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// findPodsMatchingAtLeastOneLabel Retrieves pods that match any provided label in a namespace
//
// The function iterates over each supplied label, querying the Kubernetes API
// for pods that have the corresponding key-value pair. It accumulates all
// matching pod objects into a single list, logging errors but continuing on
// failures. The resulting list is returned, containing every pod that satisfies
// at least one of the specified labels.
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

// FindPodsByLabels Retrieves pods matching specified labels across namespaces
//
// The function queries each provided namespace for pods, optionally filtering
// by one or more label key/value pairs. It returns two slices: runningPods
// contains only those that are not marked for deletion and either in the
// Running phase or allowed non‑running per configuration; allPods includes
// every pod found regardless of status. Errors during listing are logged and
// skipped.
func FindPodsByLabels(oc corev1client.CoreV1Interface, labels []labelObject, namespaces []string) (runningPods, allPods []corev1.Pod) {
	runningPods = []corev1.Pod{}
	allPods = []corev1.Pod{}
	allowNonRunning := configuration.GetTestParameters().AllowNonRunning
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
			if pods.Items[i].DeletionTimestamp == nil {
				if allowNonRunning || pods.Items[i].Status.Phase == corev1.PodRunning {
					runningPods = append(runningPods, pods.Items[i])
				}
			}
			allPods = append(allPods, pods.Items[i])
		}
	}

	return runningPods, allPods
}

// CountPodsByStatus Counts running versus non‑running pods
//
// The function iterates over a slice of pod objects, incrementing counters for
// those in the Running phase versus all others. It returns a map with keys
// "ready" and "non-ready" holding the respective counts. The result is used to
// track pod state before and after test execution.
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
