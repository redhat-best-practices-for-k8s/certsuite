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

package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	v1core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func buildServiceAccountTokenTestObjects() []runtime.Object {
	falseVar := false
	trueVar := true
	testSAwithSATokenTrue := v1core.ServiceAccount{
		AutomountServiceAccountToken: &trueVar,
		ObjectMeta: v1.ObjectMeta{
			Name:      "SAAutomountTrue",
			Namespace: "testNamespace",
		},
	}
	testSAwithSATokenFalse := v1core.ServiceAccount{
		AutomountServiceAccountToken: &falseVar,
		ObjectMeta: v1.ObjectMeta{
			Name:      "SAAutomountFalse",
			Namespace: "testNamespace",
		},
	}
	testSAwithSATokenNil := v1core.ServiceAccount{
		AutomountServiceAccountToken: nil,
		ObjectMeta: v1.ObjectMeta{
			Name:      "SAAutomountNil",
			Namespace: "testNamespace",
		},
	}

	var testRuntimeObjects []runtime.Object
	testRuntimeObjects = append(testRuntimeObjects, &testSAwithSATokenTrue, &testSAwithSATokenFalse, &testSAwithSATokenNil)
	return testRuntimeObjects
}

func TestAutomountServiceAccountSetOnSA(t *testing.T) {
	testCases := []struct {
		automountServiceTokenSet bool
	}{
		{
			automountServiceTokenSet: true,
		},
		{
			automountServiceTokenSet: false,
		},
	}

	for _, tc := range testCases {
		testSA := v1core.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "podNS",
				Name:      "testSA",
			},
			AutomountServiceAccountToken: &tc.automountServiceTokenSet,
		}

		var testRuntimeObjects []runtime.Object
		testRuntimeObjects = append(testRuntimeObjects, &testSA)

		obj := NewAutomountTokenTester(clientsholder.GetTestClientsHolder(testRuntimeObjects))
		assert.NotNil(t, obj)
		isSet, err := obj.AutomountServiceAccountSetOnSA("testSA", "podNS")
		assert.Nil(t, err)
		assert.Equal(t, tc.automountServiceTokenSet, *isSet)
	}
}

//nolint:funlen
func TestEvaluateTokens(t *testing.T) {
	falseVar := false
	trueVar := true

	generatePod := func(tokenStatus *bool, saName string) *v1core.Pod {
		return &v1core.Pod{
			Spec: v1core.PodSpec{
				NodeName:                     "worker01",
				AutomountServiceAccountToken: tokenStatus,
				ServiceAccountName:           saName,
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "testPod",
				Namespace: "testNamespace",
			},
		}
	}

	testCases := []struct {
		testPod        *v1core.Pod
		expectedMsg    string
		expectedResult bool
	}{
		{ // Test Case #1 - PASS - Automount Service Token on the pod is set to False
			testPod:        generatePod(&falseVar, "SAAutomountTrue"),
			expectedResult: true,
			expectedMsg:    "",
		},
		{ // Test Case #2 - FAIL - Automount Service Token on the pod is set to True
			testPod:        generatePod(&trueVar, "SAAutomountTrue"),
			expectedResult: false,
			expectedMsg:    "Pod testNamespace:testPod is configured with automountServiceAccountToken set to true",
		},
		{ // Test Case #3 - PASS - Pod SAT is nil, SA is false
			testPod:        generatePod(nil, "SAAutomountFalse"),
			expectedResult: true,
			expectedMsg:    "",
		},
		{ // Test Case #4 - FAIL - Pod SAT is nil, SA is true
			testPod:        generatePod(nil, "SAAutomountTrue"),
			expectedResult: false,
			expectedMsg:    "serviceaccount testNamespace:SAAutomountTrue is configured with automountServiceAccountToken set to true, impacting pod testPod",
		},
		{ // Test Case #5 - FAIL - Pod SAT is nil, SA is nil
			testPod:        generatePod(nil, "SAAutomountNil"),
			expectedResult: false,
			expectedMsg:    "serviceaccount testNamespace:SAAutomountNil is not configured with automountServiceAccountToken set to false, impacting pod testPod",
		},
	}

	for _, tc := range testCases {
		at := NewAutomountTokenTester(clientsholder.GetTestClientsHolder(buildServiceAccountTokenTestObjects()))
		podPassed, msg := at.EvaluateTokens(tc.testPod)
		assert.Equal(t, tc.expectedMsg, msg)
		assert.Equal(t, tc.expectedResult, podPassed)
	}
}
