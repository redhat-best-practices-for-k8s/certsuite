// Copyright (C) 2020-2026 Red Hat, Inc.
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

func TestGetServices(t *testing.T) {
	generateService := func(name, namespace string) *corev1.Service {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		serviceName      string
		serviceNamespace string
		ignoreList       []string
		expectedServices []*corev1.Service
	}{
		{ // Test case #1: Service found
			serviceName:      "testService",
			serviceNamespace: "tnf",
			ignoreList:       []string{},
			expectedServices: []*corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testService",
						Namespace: "tnf",
					},
				},
			},
		},
		{ // Test case #2: Service in ignore list not found
			serviceName:      "testService",
			serviceNamespace: "tnf",
			ignoreList:       []string{"testService"},
			expectedServices: nil,
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generateService(tc.serviceName, tc.serviceNamespace))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)
		services, err := getServices(oc.K8sClient.CoreV1(), []string{tc.serviceNamespace}, tc.ignoreList)
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedServices, services)
	}
}
