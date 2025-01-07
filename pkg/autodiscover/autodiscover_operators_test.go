// Copyright (C) 2023-2024 Red Hat, Inc.

package autodiscover

import (
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
		// Generate the namespaces for the test
		var testRuntimeObjects []runtime.Object
		for _, n := range generateNamespaces(tc.testNamespaces) {
			testRuntimeObjects = append(testRuntimeObjects, n)
		}

		clientSet := fake.NewSimpleClientset(testRuntimeObjects...)
		namespaces, err := getAllNamespaces(clientSet.CoreV1())
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedNamespaces, namespaces)
	}
}
