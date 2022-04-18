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

package podrecreation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func generatePod(name, kind string) corev1.Pod {
	getIntPointer := func(val int64) *int64 {
		return &val
	}

	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: kind,
				},
			},
			Labels: map[string]string{
				"pod-template-hash": "test",
			},
		},
		Spec: corev1.PodSpec{
			NodeName:                      "node1",
			TerminationGracePeriodSeconds: getIntPointer(30),
		},
	}
}

func TestCountPodsWithDelete(t *testing.T) {
	testCases := []struct {
		testPods      []corev1.Pod
		expectedCount int
	}{
		{ // Test Case #1 - One deleted pod because one is a daemonset.
			expectedCount: 1,
			testPods: []corev1.Pod{
				generatePod("testpod1", DeploymentString),
				generatePod("testpod2", DaemonSetString),
			},
		},
		{ // Test Case #2 - Two pods deleted, both deployments
			expectedCount: 2,
			testPods: []corev1.Pod{
				generatePod("testpod1", DeploymentString),
				generatePod("testpod2", DeploymentString),
			},
		},
	}

	for _, tc := range testCases {
		// Build a test clientsHolder
		var testRuntimeObjects []runtime.Object
		for i := range tc.testPods {
			x := tc.testPods[i]
			testRuntimeObjects = append(testRuntimeObjects, &x)
		}
		// Clean and recreate the clientsHolder
		clientsholder.ClearTestClientsHolder()
		_ = clientsholder.GetTestClientsHolder(testRuntimeObjects)

		result, err := CountPodsWithDelete("node1", true)
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedCount, result)
	}
}

func generateNode(name string) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
		Spec: corev1.NodeSpec{
			Unschedulable: false,
		},
	}
}

func TestCordonHelper(t *testing.T) {
	testCases := []struct {
		operation string
		testNodes []corev1.Node
	}{
		{
			operation: Cordon,
			testNodes: []corev1.Node{
				generateNode("node1"),
			},
		},
		{
			operation: Uncordon,
			testNodes: []corev1.Node{
				generateNode("node1"),
			},
		},
	}

	for _, tc := range testCases {
		// Build a test clientsHolder
		var testRuntimeObjects []runtime.Object
		for i := range tc.testNodes {
			x := tc.testNodes[i]
			testRuntimeObjects = append(testRuntimeObjects, &x)
		}
		// Clean and recreate the clientsHolder
		clientsholder.ClearTestClientsHolder()
		_ = clientsholder.GetTestClientsHolder(testRuntimeObjects)
		err := CordonHelper("node1", tc.operation)
		assert.Nil(t, err)
	}
}
