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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// getResourceQuotas retrieves all ResourceQuota objects from the cluster.
//
// It takes a CoreV1 client interface and returns a slice of ResourceQuota resources along with an error if any occurs during listing.
// The function calls the client's ResourceQuotas method and lists all items in the default namespace. The returned slice contains the full
// ResourceQuota objects, which can be inspected for hard and used limits. Errors from the List call are propagated to the caller.
func getResourceQuotas(oc corev1client.CoreV1Interface) ([]corev1.ResourceQuota, error) {
	rql, err := oc.ResourceQuotas("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return rql.Items, nil
}
