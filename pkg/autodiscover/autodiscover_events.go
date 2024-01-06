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

	"github.com/test-network-function/cnf-certification-test/internal/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

func findAbnormalEvents(oc corev1client.CoreV1Interface, namespaces []string) (abnormalEvents []corev1.Event) {
	abnormalEvents = []corev1.Event{}
	for _, ns := range namespaces {
		someAbnormalEvents, err := oc.Events(ns).List(context.TODO(), metav1.ListOptions{FieldSelector: "type!=Normal"})
		if err != nil {
			log.Error("failed to get event list for namespace %s, err:%s", ns, err)
			continue
		}
		abnormalEvents = append(abnormalEvents, someAbnormalEvents.Items...)
	}
	return abnormalEvents
}
