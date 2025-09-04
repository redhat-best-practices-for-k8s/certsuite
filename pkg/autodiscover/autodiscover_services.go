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

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// getServices Retrieves services from specified namespaces while excluding ignored names
//
// The function iterates over a list of namespace strings, querying the
// Kubernetes API for services in each one. It filters out any service whose
// name appears in an ignore list using a helper that checks string membership.
// Matching services are collected into a slice and returned; if any API call
// fails, the error is propagated immediately.
func getServices(oc corev1client.CoreV1Interface, namespaces, ignoreList []string) (allServices []*corev1.Service, err error) {
	for _, ns := range namespaces {
		s, err := oc.Services(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return allServices, err
		}
		for i := range s.Items {
			if stringhelper.StringInSlice(ignoreList, s.Items[i].Name, false) {
				continue
			}
			allServices = append(allServices, &s.Items[i])
		}
	}
	return allServices, nil
}
