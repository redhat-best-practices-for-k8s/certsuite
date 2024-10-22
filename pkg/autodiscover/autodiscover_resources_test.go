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
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetResourceQuotas(t *testing.T) {
	generateResourceQuota := func(name string) *corev1.ResourceQuota {
		return &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: corev1.ResourceQuotaSpec{},
		}
	}

	testCases := []struct {
		rqName      string
		expectedRQs []corev1.ResourceQuota
	}{
		{
			rqName: "test1",
			expectedRQs: []corev1.ResourceQuota{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generateResourceQuota(tc.rqName))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)
		resourceQuotas, err := getResourceQuotas(oc.K8sClient.CoreV1())
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedRQs[0].Name, resourceQuotas[0].Name)
	}
}

func TestDoesAPIResourceExist(t *testing.T) {
	generateResource := func(name string) *metav1.APIResourceList {
		return &metav1.APIResourceList{
			APIResources: []metav1.APIResource{
				{
					Name: name,
				},
			},
		}
	}

	testCases := []struct {
		testResourceToSearch string
		expectedResult       bool
		runtimeObjects       []*metav1.APIResourceList
	}{
		{
			testResourceToSearch: "test1",
			expectedResult:       true,
			runtimeObjects: []*metav1.APIResourceList{
				generateResource("test1"),
			},
		},
		{
			testResourceToSearch: "test2",
			expectedResult:       false,
			runtimeObjects: []*metav1.APIResourceList{
				generateResource("test1"),
			},
		},
		{
			testResourceToSearch: "test3",
			expectedResult:       true,
			runtimeObjects: []*metav1.APIResourceList{
				generateResource("test1"),
				generateResource("test3"),
			},
		},
		{
			testResourceToSearch: "test4",
			expectedResult:       false,
			runtimeObjects:       nil,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.expectedResult, doesAPIResourceExist(testCase.runtimeObjects, testCase.testResourceToSearch))
	}
}
