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
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetServiceAccounts(t *testing.T) {
	generateServiceAccount := func(name, namespace string) *corev1.ServiceAccount {
		return &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		serviceAccountName      string
		serviceAccountNamespace string
		expectedServiceAccounts []*corev1.ServiceAccount
	}{
		{
			serviceAccountName:      "testServiceAccount",
			serviceAccountNamespace: "tnf",
			expectedServiceAccounts: []*corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testServiceAccount",
						Namespace: "tnf",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generateServiceAccount(tc.serviceAccountName, tc.serviceAccountNamespace))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)
		services, err := getServiceAccounts(oc.K8sClient.CoreV1(), []string{tc.serviceAccountNamespace})
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedServiceAccounts, services)
	}
}
