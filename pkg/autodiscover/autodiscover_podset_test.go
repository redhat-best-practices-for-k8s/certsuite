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

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	scalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
)

func TestFindDeploymentByLabel(t *testing.T) {
	generateDeployment := func(name, namespace, label string) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"testLabel": label,
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testNamespaces          []string
		expectedResults         []appsv1.Deployment
		testDeploymentName      string
		testDeploymentNamespace string
		testDeploymentLabel     string
		queryLabel              string
	}{
		{ // Test Case #1 - Happy path, labels found
			testDeploymentName:      "testName",
			testDeploymentNamespace: "testNamespace",
			testDeploymentLabel:     "mylabel",
			queryLabel:              "mylabel",

			expectedResults: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testName",
						Namespace: "testNamespace",
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"testLabel": "mylabel",
								},
							},
						},
					},
				},
			},
		},
		{ // Test Case #2 - Invalid label
			testDeploymentName:      "testName",
			testDeploymentNamespace: "testNamespace",
			testDeploymentLabel:     "testLabel",
			queryLabel:              "badlabel",

			expectedResults: []appsv1.Deployment{},
		},
	}

	for _, tc := range testCases {
		testLabel := []labelObject{{LabelKey: "testLabel", LabelValue: tc.testDeploymentLabel}}

		testNamespaces := []string{
			tc.testDeploymentNamespace,
		}
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generateDeployment(tc.testDeploymentName, tc.testDeploymentNamespace, tc.queryLabel))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)

		deployments := findDeploymentByLabel(oc.K8sClient.AppsV1(), testLabel, testNamespaces)
		assert.Equal(t, tc.expectedResults, deployments)
	}
}

func TestFindStatefulSetByLabel(t *testing.T) {
	generateStatefulSet := func(name, namespace, label string) *appsv1.StatefulSet {
		return &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"testLabel": label,
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testNamespaces           []string
		expectedResults          []appsv1.StatefulSet
		testStatefulSetName      string
		testStatefulSetNamespace string
		testStatefulSetLabel     string
		queryLabel               string
	}{
		{ // Test Case #1 - Happy path, labels found
			testStatefulSetName:      "testName",
			testStatefulSetNamespace: "testNamespace",
			testStatefulSetLabel:     "mylabel",
			queryLabel:               "mylabel",

			expectedResults: []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testName",
						Namespace: "testNamespace",
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"testLabel": "mylabel",
								},
							},
						},
					},
				},
			},
		},
		{ // Test Case #2 - Invalid label
			testStatefulSetName:      "testName",
			testStatefulSetNamespace: "testNamespace",
			testStatefulSetLabel:     "testLabel",
			queryLabel:               "badlabel",

			expectedResults: []appsv1.StatefulSet{},
		},
	}

	for _, tc := range testCases {
		testLabel := []labelObject{{LabelKey: "testLabel", LabelValue: tc.testStatefulSetLabel}}
		testNamespaces := []string{
			tc.testStatefulSetNamespace,
		}
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generateStatefulSet(tc.testStatefulSetName, tc.testStatefulSetNamespace, tc.queryLabel))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)

		statefulSets := findStatefulSetByLabel(oc.K8sClient.AppsV1(), testLabel, testNamespaces)
		assert.Equal(t, tc.expectedResults, statefulSets)
	}
}

func TestFindHpaControllers(t *testing.T) {
	generateHpa := func(name, namespace string) *scalingv1.HorizontalPodAutoscaler {
		return &scalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		testHpaName      string
		testHpaNamespace string
		expectedResults  []*scalingv1.HorizontalPodAutoscaler
	}{
		{
			testHpaName:      "testName",
			testHpaNamespace: "testNamespace",
			expectedResults: []*scalingv1.HorizontalPodAutoscaler{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testName",
						Namespace: "testNamespace",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generateHpa(tc.testHpaName, tc.testHpaNamespace))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)

		hpas := findHpaControllers(oc.K8sClient, []string{tc.testHpaNamespace})
		assert.Equal(t, tc.expectedResults, hpas)
	}
}

func TestFindDeploymentByNameByNamespace(t *testing.T) {
	generateDeployment := func(name, namespace string) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		testDeploymentName      string
		testDeploymentNamespace string
		expectedResults         *appsv1.Deployment
	}{
		{
			testDeploymentName:      "testName",
			testDeploymentNamespace: "testNamespace",
			expectedResults: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testName",
					Namespace: "testNamespace",
				},
			},
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generateDeployment(tc.testDeploymentName, tc.testDeploymentNamespace))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)

		deployment, err := FindDeploymentByNameByNamespace(oc.K8sClient.AppsV1(), tc.testDeploymentNamespace, tc.testDeploymentName)
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedResults, deployment)
	}
}

func TestFindStatefulSetByNameByNamespace(t *testing.T) {
	generateStatefulSet := func(name, namespace string) *appsv1.StatefulSet {
		return &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		testStatefulSetName      string
		testStatefulSetNamespace string
		expectedResults          *appsv1.StatefulSet
	}{
		{
			testStatefulSetName:      "testName",
			testStatefulSetNamespace: "testNamespace",
			expectedResults: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testName",
					Namespace: "testNamespace",
				},
			},
		},
	}

	for _, tc := range testCases {
		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, generateStatefulSet(tc.testStatefulSetName, tc.testStatefulSetNamespace))
		oc := clientsholder.GetTestClientsHolder(testRuntimeObjects)

		statefulSet, err := FindStatefulsetByNameByNamespace(oc.K8sClient.AppsV1(), tc.testStatefulSetNamespace, tc.testStatefulSetName)
		assert.Nil(t, err)
		assert.Equal(t, tc.expectedResults, statefulSet)
	}
}
