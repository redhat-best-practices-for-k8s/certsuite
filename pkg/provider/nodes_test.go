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

package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHasWorkloadDeployed(t *testing.T) {
	generateNode := func(nodeName string) *Node {
		return &Node{
			Data: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: nodeName,
				},
			},
		}
	}

	testCases := []struct {
		testNodeName string
		testPods     []*Pod
		expected     bool
	}{
		{
			testNodeName: "node1",
			testPods: []*Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pod1",
						},
						Spec: corev1.PodSpec{
							NodeName: "node1",
						},
					},
				},
			},
			expected: true,
		},
		{
			testNodeName: "node1",
			testPods: []*Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pod1",
						},
						Spec: corev1.PodSpec{
							NodeName: "node2",
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		n := generateNode(testCase.testNodeName)
		assert.Equal(t, testCase.expected, n.HasWorkloadDeployed(testCase.testPods))
	}
}
