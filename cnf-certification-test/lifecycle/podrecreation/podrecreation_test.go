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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func generatePod(name, ownerKind string) *provider.Pod {
	getIntPointer := func(val int64) *int64 {
		return &val
	}

	aPod := provider.NewPod(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: ownerKind,
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
	})
	return &aPod
}

func TestCountPodsWithDelete(t *testing.T) {
	testCases := []struct {
		testPods      []*provider.Pod
		expectedCount int
	}{
		{ // Test Case #1 - One deleted pod because one is a daemonset.
			expectedCount: 1,
			testPods: []*provider.Pod{
				generatePod("testpod1", DeploymentString),
				generatePod("testpod2", DaemonSetString),
			},
		},
		{ // Test Case #2 - Two pods deleted, both deployments
			expectedCount: 2,
			testPods: []*provider.Pod{
				generatePod("testpod1", DeploymentString),
				generatePod("testpod2", DeploymentString),
			},
		},
	}
	// Build a test clientsHolder (just for the call to delete to succeed)
	var testRuntimeObjects []runtime.Object
	// create the clientsHolder
	_ = clientsholder.GetTestClientsHolder(testRuntimeObjects)
	for _, tc := range testCases {
		result, err := CountPodsWithDelete(tc.testPods, "node1", DeleteBackground)
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedCount, result)
	}
}

func generateNode(name string) corev1.Node {
	return corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
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
		client := clientsholder.GetTestClientsHolder(testRuntimeObjects)
		err := CordonHelper("node1", tc.operation)
		assert.Nil(t, err)

		// Check that the node is actually cordoned or uncordoned
		node, err := client.K8sClient.CoreV1().Nodes().Get(context.TODO(), "node1", metav1.GetOptions{})
		assert.Nil(t, err)
		if tc.operation == Cordon {
			assert.True(t, node.Spec.Unschedulable)
		} else {
			assert.False(t, node.Spec.Unschedulable)
		}
	}
}
