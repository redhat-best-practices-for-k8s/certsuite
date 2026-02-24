// Copyright (C) 2020-2026 Red Hat, Inc.
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

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/podsets"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestScaleStatefulSetFunc(t *testing.T) {
	generateStatefulSet := func(name string, replicas *int32) *appsv1.StatefulSet {
		return &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "namespace1",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: replicas,
			},
		}
	}

	testCases := []struct {
		statefulSetName string
		replicaCount    int
	}{
		{
			statefulSetName: "ss1",
			replicaCount:    3,
		},
		{
			statefulSetName: "ss2",
			replicaCount:    0,
		},
	}

	// Always return that the StatefulSet is Ready
	origFunc := podsets.WaitForStatefulSetReady
	defer func() {
		podsets.WaitForStatefulSetReady = origFunc
	}()
	podsets.WaitForStatefulSetReady = func(ns, name string, timeout time.Duration, logger *log.Logger) bool {
		return true
	}

	for _, tc := range testCases {
		var runtimeObjects []runtime.Object
		intVar := new(int32)
		*intVar = int32(tc.replicaCount)
		tempSS := generateStatefulSet(tc.statefulSetName, intVar)
		runtimeObjects = append(runtimeObjects, tempSS)
		c := clientsholder.GetTestClientsHolder(runtimeObjects)

		var logArchive strings.Builder
		log.SetupLogger(&logArchive, "INFO")
		TestScaleStatefulSet(tempSS, 10*time.Second, log.GetLogger())

		ss, err := c.K8sClient.AppsV1().StatefulSets("namespace1").Get(context.TODO(), tc.statefulSetName, metav1.GetOptions{})
		assert.Nil(t, err)
		assert.Equal(t, int32(tc.replicaCount), *ss.Spec.Replicas)
	}
}

func TestScaleHpaStatefulSetFunc(t *testing.T) {
	generateStatefulSet := func(name string, replicas *int32) *appsv1.StatefulSet {
		return &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "namespace1",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: replicas,
			},
		}
	}

	testCases := []struct {
		statefulSetName      string
		replicaCount         int
		hpaMinReplicas       int32
		hpaMaxReplicas       int32
		expectedReplicaCount int
	}{
		{ // Test Case 1 - Start with 3 replicas, HPA min 1, HPA max 3, expect 3 replicas
			statefulSetName:      "ss1",
			replicaCount:         3,
			hpaMinReplicas:       1,
			hpaMaxReplicas:       3,
			expectedReplicaCount: 3,
		},
		{ // Test Case 2 - Start with 1 replica, HPA min 1, HPA max 3, expect 3 replicas
			statefulSetName:      "ss2",
			replicaCount:         1,
			hpaMinReplicas:       1,
			hpaMaxReplicas:       3,
			expectedReplicaCount: 3,
		},
	}

	// Always return that the StatefulSet is Ready
	origFunc := podsets.WaitForStatefulSetReady
	defer func() {
		podsets.WaitForStatefulSetReady = origFunc
	}()
	podsets.WaitForStatefulSetReady = func(ns, name string, timeout time.Duration, logger *log.Logger) bool {
		return true
	}

	int32Ptr := func(i int32) *int32 { return &i }

	for _, tc := range testCases {
		intVar := new(int32)
		*intVar = int32(tc.replicaCount)
		tempSS := generateStatefulSet(tc.statefulSetName, intVar)

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
		runtimeObjects = append(runtimeObjects, tempSS, hpatest)
		c := k8sfake.NewClientset(runtimeObjects...)
		c.AddReactor("get", "horizontalpodautoscalers", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, hpatest, nil
		})

		c.AddReactor("update", "horizontalpodautoscalers", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, hpatest, nil
		})

		clientsholder.SetTestK8sClientsHolder(c)

		var logArchive strings.Builder
		log.SetupLogger(&logArchive, "INFO")
		TestScaleHpaStatefulSet(tempSS, hpatest, 10*time.Second, log.GetLogger())

		hpa, err := c.AutoscalingV1().HorizontalPodAutoscalers("namespace1").Get(context.TODO(), "hpaName", metav1.GetOptions{})
		assert.Nil(t, err)
		assert.Equal(t, int32(tc.expectedReplicaCount), hpa.Spec.MaxReplicas)
	}
}

func TestScaleStatefulsetHelper(t *testing.T) {
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
		{ // Test Case 2 - Error updating the statefulset
			getResult:      nil,
			updateResult:   errors.New("this is an error"),
			expectedOutput: false,
		},
		{ // Test Case 3 - Error getting the statefulset
			getResult:      errors.New("this is an error"),
			updateResult:   nil,
			expectedOutput: false,
		},
	}

	defer clientsholder.ClearTestClientsHolder()

	// Always return that the StatefulSet is Ready
	origFunc := podsets.WaitForStatefulSetReady
	defer func() {
		podsets.WaitForStatefulSetReady = origFunc
	}()
	podsets.WaitForStatefulSetReady = func(ns, name string, timeout time.Duration, logger *log.Logger) bool {
		return true
	}

	for _, tc := range testCases {
		ss := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ss1",
				Namespace: "ns1",
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		}

		var runtimeObjs []runtime.Object
		runtimeObjs = append(runtimeObjs, ss)

		// Create a fake client with reactors for both get and update
		fakeClient := k8sfake.NewClientset(runtimeObjs...)
		fakeClient.PrependReactor("get", "statefulsets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, ss, tc.getResult
		})
		fakeClient.PrependReactor("update", "statefulsets", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, ss, tc.updateResult
		})

		// Set the fake client as the clientsholder's K8s client so both ssClient and
		// clients.K8sClient.AppsV1().StatefulSets() use the same client with reactors.
		clientsholder.SetTestK8sClientsHolder(fakeClient)
		clients := clientsholder.GetClientsHolder()

		var logArchive strings.Builder
		log.SetupLogger(&logArchive, "INFO")
		result := scaleStatefulsetHelper(clients, fakeClient.AppsV1().StatefulSets("ns1"), ss, 1, 10*time.Second, log.GetLogger())
		assert.Equal(t, tc.expectedOutput, result)
	}
}

func TestScaleHpaStatefulSetHelper(t *testing.T) {
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
		{ // Test Case 2 - Error updating the HPA
			getResult:      nil,
			updateResult:   errors.New("this is an error"),
			expectedOutput: false,
		},
		{ // Test Case 3 - Error getting the HPA
			getResult:      errors.New("this is an error"),
			updateResult:   nil,
			expectedOutput: false,
		},
	}

	defer clientsholder.ClearTestClientsHolder()

	// Always return that the StatefulSet is Ready
	origFunc := podsets.WaitForStatefulSetReady
	defer func() {
		podsets.WaitForStatefulSetReady = origFunc
	}()
	podsets.WaitForStatefulSetReady = func(ns, name string, timeout time.Duration, logger *log.Logger) bool {
		return true
	}

	int32Ptr := func(i int32) *int32 { return &i }

	for _, tc := range testCases {
		hpatest := &v1autoscaling.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name: "hpaName",
			},
			Spec: v1autoscaling.HorizontalPodAutoscalerSpec{
				MinReplicas: int32Ptr(1),
				MaxReplicas: 3,
			},
		}

		var runtimeObjs []runtime.Object
		runtimeObjs = append(runtimeObjs, hpatest)
		clientsholder.GetTestClientsHolder(runtimeObjs)

		client := k8sfake.Clientset{}
		client.AddReactor("get", "horizontalpodautoscalers", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, runtimeObjs[0], tc.getResult
		})

		client.AddReactor("update", "horizontalpodautoscalers", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, runtimeObjs[0], tc.updateResult
		})

		var logArchive strings.Builder
		log.SetupLogger(&logArchive, "INFO")
		result := scaleHpaStatefulSetHelper(client.AutoscalingV1().HorizontalPodAutoscalers("ns1"), "hpaName", "ss1", "ns1", 1, 3, 10*time.Second, log.GetLogger())
		assert.Equal(t, tc.expectedOutput, result)
	}
}
