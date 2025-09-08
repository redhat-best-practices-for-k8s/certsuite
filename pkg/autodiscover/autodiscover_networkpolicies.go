// Copyright (C) 2022-2023 Red Hat, Inc.
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

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingv1client "k8s.io/client-go/kubernetes/typed/networking/v1"
)

// getNetworkPolicies Retrieves all network policies in the cluster
//
// The function calls the NetworkingV1 client to list network policies across
// every namespace by using an empty string for the namespace parameter. It
// returns a slice of NetworkPolicy objects and any error encountered during the
// API call.
func getNetworkPolicies(oc networkingv1client.NetworkingV1Interface) ([]networkingv1.NetworkPolicy, error) {
	nps, err := oc.NetworkPolicies("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nps.Items, nil
}
