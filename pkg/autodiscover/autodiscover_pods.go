// Copyright (C) 2020-2021 Red Hat, Inc.
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
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Filter out any pod that is not in a running state
const filterStatusRunning = "status.phase=Running"

func findPodsByLabel(oc corev1client.CoreV1Interface, labels []configuration.Label, namespaces []string) []v1.Pod {
	Pods := []v1.Pod{}
	for _, ns := range namespaces {
		for _, l := range labels {
			label := buildLabelQuery(l)
			logrus.Trace("find pods in ", ns, " using label= ", label)
			pods, err := oc.Pods(ns).List(context.TODO(), metav1.ListOptions{
				LabelSelector: label,
				FieldSelector: filterStatusRunning,
			})
			if err != nil {
				logrus.Errorln("error when listing pods in ns=", ns, " label=", label, "err: ", err)
				continue
			}

			// Filter out any pod set to be deleted
			for _, pod := range pods.Items {
				if pod.ObjectMeta.DeletionTimestamp == nil {
					Pods = append(Pods, pod)
				}
			}
		}
	}
	return Pods
}
