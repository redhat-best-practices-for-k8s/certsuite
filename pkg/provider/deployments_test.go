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

package provider

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentToString(t *testing.T) {
	dp := Deployment{
		Deployment: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "testNS",
			},
		},
	}

	assert.Equal(t, "deployment: test1 ns: testNS", dp.ToString())
}

func TestGetUpdatedDeployment(t *testing.T) {
	testCases := []struct {
		exists      bool
		expectedErr error
	}{
		{exists: true, expectedErr: nil},
		{exists: false, expectedErr: fmt.Errorf("deployments.apps \"test1\" not found")},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "testNS",
				},
			})
		}

		fakeClient := k8sfake.NewClientset(runtimeObjects...)

		deployment, err := GetUpdatedDeployment(fakeClient.AppsV1(), "testNS", "test1")

		if testCase.expectedErr != nil {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, deployment)
		}
	}
}

func TestIsDeploymentReady(t *testing.T) {
	generateDeployment := func(readyReplicas, unavailableReplicas, availableReplicas, updatedReplicas, replicas int32, conditions []appsv1.DeploymentCondition) *appsv1.Deployment {
		return &appsv1.Deployment{
			Status: appsv1.DeploymentStatus{
				ReadyReplicas:       readyReplicas,
				UnavailableReplicas: unavailableReplicas,
				AvailableReplicas:   availableReplicas,
				UpdatedReplicas:     updatedReplicas,
				Conditions:          conditions,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
		}
	}

	testCases := []struct {
		testDeployment *appsv1.Deployment
		expectedResult bool
	}{
		{ // Test Case #1 - Deployment condition is available
			testDeployment: generateDeployment(1, 0, 1, 1, 1, []appsv1.DeploymentCondition{
				{
					Type: appsv1.DeploymentAvailable,
				},
			}),
			expectedResult: true,
		},
		{ // Test Case #2 - Deployment condition is not available
			testDeployment: generateDeployment(1, 0, 1, 1, 1, []appsv1.DeploymentCondition{
				{
					Type: appsv1.DeploymentProgressing,
				},
			}),
			expectedResult: false,
		},
		{ // Test Case #3 - Unavailable replicas are not 0
			testDeployment: generateDeployment(1, 1, 1, 1, 1, []appsv1.DeploymentCondition{
				{
					Type: appsv1.DeploymentAvailable,
				},
			}),
			expectedResult: false,
		},
		{ // Test Case #4 - Ready replicas do not match total replicas
			testDeployment: generateDeployment(0, 0, 1, 1, 1, []appsv1.DeploymentCondition{
				{
					Type: appsv1.DeploymentAvailable,
				},
			}),
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		deployment := Deployment{
			Deployment: testCase.testDeployment,
		}

		assert.Equal(t, testCase.expectedResult, deployment.IsDeploymentReady())
	}
}
