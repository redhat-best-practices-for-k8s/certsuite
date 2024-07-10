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

package scaling

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/tests/lifecycle/podsets"
	appsv1 "k8s.io/api/apps/v1"
	v1autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestScaleDeploymentFunc(t *testing.T) {
	generateDeployment := func(name string, replicas *int32) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "namespace1",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: replicas,
			},
		}
	}

	testCases := []struct {
		deploymentName string
		replicaCount   int
	}{
		{
			deploymentName: "dp1",
			replicaCount:   3,
		},
		{
			deploymentName: "dp2",
			replicaCount:   0,
		},
	}

	// Always return that the Deployment is Ready
	origFunc := podsets.WaitForDeploymentSetReady
	defer func() {
		podsets.WaitForDeploymentSetReady = origFunc
	}()
	podsets.WaitForDeploymentSetReady = func(ns, name string, timeout time.Duration, logger *log.Logger) bool {
		return true
	}

	for _, tc := range testCases {
		var runtimeObjects []runtime.Object
		intVar := new(int32)
		*intVar = int32(tc.replicaCount)
		tempDP := generateDeployment(tc.deploymentName, intVar)
		runtimeObjects = append(runtimeObjects, tempDP)
		c := clientsholder.GetTestClientsHolder(runtimeObjects)

		// Run the function
		var logArchive strings.Builder
		log.SetupLogger(&logArchive, "INFO")
		TestScaleDeployment(tempDP, 10*time.Second, log.GetLogger())

		// Get the deployment from the fake API
		dp, err := c.K8sClient.AppsV1().Deployments("namespace1").Get(context.TODO(), tc.deploymentName, metav1.GetOptions{})
		assert.Nil(t, err)
		assert.Equal(t, int32(tc.replicaCount), *dp.Spec.Replicas)
	}
}

func TestScaleHpaDeploymentFunc(t *testing.T) {
	generateDeployment := func(name string, replicas *int32) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "namespace1",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: replicas,
			},
		}
	}

	testCases := []struct {
		deploymentName       string
		replicaCount         int
		hpaMinReplicas       int32
		hpaMaxReplicas       int32
		expectedReplicaCount int
	}{
		{ // Test Case 1 - Start with 3 replicas, HPA min 1, HPA max 3, expect 3 replicas
			deploymentName:       "dp1",
			replicaCount:         3,
			hpaMinReplicas:       1,
			hpaMaxReplicas:       3,
			expectedReplicaCount: 3,
		},
		{ // Test Case 2 - Start with 1 replica, HPA min 1, HPA max 3, expect 3 replicas
			deploymentName:       "dp2",
			replicaCount:         1,
			hpaMinReplicas:       1,
			hpaMaxReplicas:       3,
			expectedReplicaCount: 3,
		},
	}

	// Always return that the Deployment is Ready
	origFunc := podsets.WaitForDeploymentSetReady
	defer func() {
		podsets.WaitForDeploymentSetReady = origFunc
	}()
	podsets.WaitForDeploymentSetReady = func(ns, name string, timeout time.Duration, logger *log.Logger) bool {
		return true
	}

	// Create int32 pointer
	int32Ptr := func(i int32) *int32 { return &i }

	for _, tc := range testCases {
		intVar := new(int32)
		*intVar = int32(tc.replicaCount)
		tempDP := generateDeployment(tc.deploymentName, intVar)

		hpatest := &v1autoscaling.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hpaName",
				Namespace: "namespace1",
			},
			Spec: v1autoscaling.HorizontalPodAutoscalerSpec{
				MinReplicas: int32Ptr(tc.hpaMinReplicas),
				MaxReplicas: tc.hpaMaxReplicas,
			},
		}

		var runtimeObjects []runtime.Object
		runtimeObjects = append(runtimeObjects, tempDP, hpatest)
		c := k8sfake.NewSimpleClientset(runtimeObjects...)
		c.Fake.AddReactor("get", "horizontalpodautoscalers", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, hpatest, nil
		})

		c.Fake.AddReactor("update", "horizontalpodautoscalers", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, hpatest, nil
		})

		// Override the clientsholder with the fake client.
		// The scaleHpaDeployment function uses the clientsholder to get the client.
		clientsholder.SetTestK8sClientsHolder(c)

		// Put the generated deployment into a provider.Deployment
		dp := &provider.Deployment{
			Deployment: tempDP,
		}

		// Run the function
		var logArchive strings.Builder
		log.SetupLogger(&logArchive, "INFO")
		TestScaleHpaDeployment(dp, hpatest, 10*time.Second, log.GetLogger())

		// Get the deployment from the fake API
		hpa, err := c.AutoscalingV1().HorizontalPodAutoscalers("namespace1").Get(context.TODO(), "hpaName", metav1.GetOptions{})
		assert.Nil(t, err)
		assert.Equal(t, int32(tc.expectedReplicaCount), hpa.Spec.MaxReplicas)
	}
}

func TestScaleHpaDeploymentHelper(t *testing.T) {
	testCases := []struct {
		getResult      error
		updateResult   error
		expectedOutput bool
	}{
		{ // Test Case 1 - No errors issuing the get or update
			getResult:      nil,
			updateResult:   nil,
			expectedOutput: true,
		},
		{ // Test Case 2 - Error updating the deployment
			getResult:      nil,
			updateResult:   errors.New("this is an error"),
			expectedOutput: false,
		},
		{ // Test Case 3 - Error getting the deployment
			getResult:      errors.New("this is an error"),
			updateResult:   nil,
			expectedOutput: false,
		},
	}

	// Clear the test clientsholder object when complete
	defer clientsholder.ClearTestClientsHolder()

	// Create int32 pointer
	int32Ptr := func(i int32) *int32 { return &i }

	for _, tc := range testCases {
		// Create a spoofed deployment to pass to the clientsholder.
		// This is only needed because the podsets.WaitForDeploymentSetReady function
		// utilizes the clientsholder straight from the function to retrieve the latest information.
		hpatest := &v1autoscaling.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name: "hpaName",
			},
			Spec: v1autoscaling.HorizontalPodAutoscalerSpec{
				MinReplicas: int32Ptr(1),
				MaxReplicas: 3,
			},
		}

		// Spoof the clientsholder with runtime objects
		var runtimeObjs []runtime.Object
		runtimeObjs = append(runtimeObjs, hpatest)
		clientsholder.GetTestClientsHolder(runtimeObjs)

		// Spoof the get and update functions
		client := k8sfake.Clientset{}
		client.Fake.AddReactor("get", "horizontalpodautoscalers", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, runtimeObjs[0], tc.getResult
		})

		client.Fake.AddReactor("update", "horizontalpodautoscalers", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, runtimeObjs[0], tc.updateResult
		})

		var logArchive strings.Builder
		log.SetupLogger(&logArchive, "INFO")
		result := scaleHpaDeploymentHelper(client.AutoscalingV1().HorizontalPodAutoscalers("ns1"), "hpaName", "dp1", "ns1", 1, 3, 10*time.Second, log.GetLogger())
		assert.Equal(t, tc.expectedOutput, result)
	}
}

func TestScaleDeploymentHelper(t *testing.T) {
	testCases := []struct {
		getResult      error
		updateResult   error
		expectedOutput bool
	}{
		{ // Test Case 1 - No errors issuing the get or update
			getResult:      nil,
			updateResult:   nil,
			expectedOutput: true,
		},
		{ // Test Case 2 - Error updating the deployment
			getResult:      nil,
			updateResult:   errors.New("this is an error"),
			expectedOutput: false,
		},
		{ // Test Case 3 - Error getting the deployment
			getResult:      errors.New("this is an error"),
			updateResult:   nil,
			expectedOutput: false,
		},
	}

	// Clear the test clientsholder object when complete
	defer clientsholder.ClearTestClientsHolder()

	for _, tc := range testCases {
		// Create a spoofed deployment to pass to the clientsholder.
		// This is only needed because the podsets.WaitForDeploymentSetReady function
		// utilizes the clientsholder straight from the function to retrieve the latest information.
		dep := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dp1",
				Namespace: "ns1",
			},
			Status: appsv1.DeploymentStatus{
				Conditions: []appsv1.DeploymentCondition{
					{
						Type: appsv1.DeploymentAvailable,
					},
				},
				Replicas:          1,
				ReadyReplicas:     1,
				UpdatedReplicas:   1,
				AvailableReplicas: 1,
			},
		}

		// Spoof the clientsholder with runtime objects
		var runtimeObjs []runtime.Object
		runtimeObjs = append(runtimeObjs, dep)
		clientsholder.GetTestClientsHolder(runtimeObjs)

		// Spoof the get and update functions
		client := k8sfake.Clientset{}
		client.Fake.AddReactor("get", "deployments", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, runtimeObjs[0], tc.getResult
		})

		client.Fake.AddReactor("update", "deployments", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, runtimeObjs[0], tc.updateResult
		})

		var logArchive strings.Builder
		log.SetupLogger(&logArchive, "INFO")
		result := scaleDeploymentHelper(client.AppsV1(), dep, 1, 10*time.Second, true, log.GetLogger())
		assert.Equal(t, tc.expectedOutput, result)
	}
}
