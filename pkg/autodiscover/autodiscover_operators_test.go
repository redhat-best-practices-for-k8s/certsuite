// Copyright (C) 2020-2022 Red Hat, Inc.
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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetAllNamespaces(t *testing.T) {
	generateNamespaces := func(ns []string) []*corev1.Namespace {
		var namespaces []*corev1.Namespace
		for _, n := range ns {
			namespaces = append(namespaces, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      n,
					Namespace: n,
				},
			})
		}
		return namespaces
	}

	testCases := []struct {
		testNamespaces     []string
		expectedNamespaces []string
	}{
		{
			testNamespaces:     []string{"ns1"},
			expectedNamespaces: []string{"ns1"},
		},
		{
			testNamespaces:     []string{"ns1", "ns2"},
			expectedNamespaces: []string{"ns1", "ns2"},
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object

		// Generate the namespaces for the test
		namespacesToTest := generateNamespaces(tc.testNamespaces)
		for i := range namespacesToTest {
			testRuntimeObjects = append(testRuntimeObjects, namespacesToTest[i])
		}

		clientSet := fake.NewSimpleClientset(testRuntimeObjects...)
		namespaces, err := getAllNamespaces(clientSet.CoreV1())
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(tc.expectedNamespaces, namespaces))
	}
}
