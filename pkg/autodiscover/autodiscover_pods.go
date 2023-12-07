// Copyright (C) 2020-2023 Red Hat, Inc.
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

	"github.com/test-network-function/cnf-certification-test/internal/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

func findPodsByLabel(oc corev1client.CoreV1Interface, labels []labelObject, namespaces []string) (runningPods, allPods []corev1.Pod) {
	runningPods = []corev1.Pod{}
	allPods = []corev1.Pod{}
	for _, ns := range namespaces {
		for _, aLabelObject := range labels {
			label := aLabelObject.LabelKey + "=" + aLabelObject.LabelValue
			log.Debug("Searching Pods with label %s", label)
			pods, err := oc.Pods(ns).List(context.TODO(), metav1.ListOptions{
				LabelSelector: label,
			})
			if err != nil {
				log.Error("error when listing pods in ns=%s label=%s, err: %v", ns, label, err)
				continue
			}

			// Filter out any pod set to be deleted
			for i := 0; i < len(pods.Items); i++ {
				if pods.Items[i].ObjectMeta.DeletionTimestamp == nil &&
					pods.Items[i].Status.Phase == corev1.PodRunning {
					runningPods = append(runningPods, pods.Items[i])
				}
				allPods = append(allPods, pods.Items[i])
			}
		}
	}
	return runningPods, allPods
}
