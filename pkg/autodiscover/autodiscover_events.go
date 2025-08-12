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

package autodiscover

import (
	"context"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// findAbnormalEvents retrieves events from the provided client that are considered abnormal based on specific criteria.
//
// It takes a CoreV1Interface to access the Kubernetes API and a slice of namespace names to limit the search.
// The function lists all events in each given namespace, filters them for abnormal conditions (e.g., high severity or specific reason),
// and returns a slice of corev1.Event objects that match those conditions. If an error occurs while listing events, it is logged but not returned,
// allowing the caller to proceed with any successfully retrieved events.
func findAbnormalEvents(oc corev1client.CoreV1Interface, namespaces []string) (abnormalEvents []corev1.Event) {
	abnormalEvents = []corev1.Event{}
	for _, ns := range namespaces {
		someAbnormalEvents, err := oc.Events(ns).List(context.TODO(), metav1.ListOptions{FieldSelector: "type!=Normal"})
		if err != nil {
			log.Error("Failed to get event list for namespace %q, err: %v", ns, err)
			continue
		}
		abnormalEvents = append(abnormalEvents, someAbnormalEvents.Items...)
	}
	return abnormalEvents
}
