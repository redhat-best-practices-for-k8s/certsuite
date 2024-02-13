// Copyright (C) 2020-2023 Red Hat, Inc.
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

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
)

func TestFindPodsUnderTest(t *testing.T) {
	generatePod := func(podname, namespace, label string) *corev1.Pod {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      podname,
				Namespace: namespace,
				Labels: map[string]string{
					"testLabel": label,
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		}
	}

	testCases := []struct {
		testNamespaces   []string
		expectedResults  []corev1.Pod
		testPodName      string
		testPodNamespace string
		testPodLabel     string
		queryLabel       string
	}{
		{ // Test Case #1 - Happy path, labels found
			testPodName:      "testPod",
			testPodNamespace: "testNamespace",
			testPodLabel:     "mylabel",
			queryLabel:       "mylabel",

			expectedResults: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testPod",
						Namespace: "testNamespace",
						Labels: map[string]string{
							"testLabel": "mylabel",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
		},
		{ // Test Case #2 - Invalid label
			testPodName:      "testPod",
			testPodNamespace: "testNamespace",
			testPodLabel:     "testLabel",
			queryLabel:       "badlabel",

			expectedResults: []corev1.Pod{},
		},
	}

	for _, tc := range testCases {
		testLabel := []labelObject{{LabelKey: "testLabel", LabelValue: tc.testPodLabel}}
		testNamespaces := []string{
			tc.testPodNamespace,
		}
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generatePod(tc.testPodName, tc.testPodNamespace, tc.queryLabel))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)

		podResult, _ := findPodsByLabels(oc.K8sClient.CoreV1(), testLabel, testNamespaces)
		assert.Equal(t, tc.expectedResults, podResult)
	}
}
