// Copyright (C) 2022-2024 Red Hat, Inc.
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

package rbac

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func buildServiceAccountTokenTestObjects() []runtime.Object {
	falseVar := false
	trueVar := true
	testSAwithSATokenTrue := corev1.ServiceAccount{
		AutomountServiceAccountToken: &trueVar,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "SAAutomountTrue",
			Namespace: "testNamespace",
		},
	}
	testSAwithSATokenFalse := corev1.ServiceAccount{
		AutomountServiceAccountToken: &falseVar,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "SAAutomountFalse",
			Namespace: "testNamespace",
		},
	}
	testSAwithSATokenNil := corev1.ServiceAccount{
		AutomountServiceAccountToken: nil,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "SAAutomountNil",
			Namespace: "testNamespace",
		},
	}

	var testRuntimeObjects []runtime.Object
	testRuntimeObjects = append(testRuntimeObjects, &testSAwithSATokenTrue, &testSAwithSATokenFalse, &testSAwithSATokenNil)
	return testRuntimeObjects
}

func TestEvaluateAutomountTokens(t *testing.T) {
	falseVar := false
	trueVar := true

	generatePod := func(tokenStatus, saTokenStatus *bool, saName string) provider.Pod {
		aPod := provider.NewPod(&corev1.Pod{
			Spec: corev1.PodSpec{
				NodeName:                     "worker01",
				AutomountServiceAccountToken: tokenStatus,
				ServiceAccountName:           saName,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testPod",
				Namespace: "testNamespace",
			},
		})
		var sa corev1.ServiceAccount

		sa.Name = saName
		sa.Namespace = aPod.Namespace
		sa.AutomountServiceAccountToken = saTokenStatus
		aPod.AllServiceAccountsMap = &map[string]*corev1.ServiceAccount{
			aPod.Namespace + saName: &sa,
		}
		return aPod
	}

	testCases := []struct {
		testPod        provider.Pod
		expectedMsg    string
		expectedResult bool
	}{
		{ // Test Case #1 - PASS - Automount Service Token on the pod is set to False
			testPod:        generatePod(&falseVar, &falseVar, "SAAutomountTrue"),
			expectedResult: true,
			expectedMsg:    "",
		},
		{ // Test Case #2 - FAIL - Automount Service Token on the pod is set to True
			testPod:        generatePod(&trueVar, &falseVar, "SAAutomountTrue"),
			expectedResult: false,
			expectedMsg:    "Pod testNamespace:testPod is configured with automountServiceAccountToken set to true",
		},
		{ // Test Case #3 - PASS - Pod SAT is nil, SA is false
			testPod:        generatePod(nil, &falseVar, "SAAutomountFalse"),
			expectedResult: true,
			expectedMsg:    "",
		},
		{ // Test Case #4 - FAIL - Pod SAT is nil, SA is true
			testPod:        generatePod(nil, &trueVar, "SAAutomountTrue"),
			expectedResult: false,
			expectedMsg:    "serviceaccount testNamespace:SAAutomountTrue is configured with automountServiceAccountToken set to true, impacting pod testPod",
		},
		{ // Test Case #5 - FAIL - Pod SAT is nil, SA is nil
			testPod:        generatePod(nil, nil, "SAAutomountNil"),
			expectedResult: false,
			expectedMsg:    "serviceaccount testNamespace:SAAutomountNil is not configured with automountServiceAccountToken set to false, impacting pod testPod",
		},
	}

	for _, tc := range testCases {
		client := clientsholder.GetTestClientsHolder(buildServiceAccountTokenTestObjects())
		podPassed, msg := EvaluateAutomountTokens(client.K8sClient.CoreV1(), &tc.testPod)
		assert.Equal(t, tc.expectedMsg, msg)
		assert.Equal(t, tc.expectedResult, podPassed)
	}
}
