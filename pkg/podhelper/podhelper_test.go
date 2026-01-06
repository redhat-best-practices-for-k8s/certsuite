// Copyright (C) 2024-2026 Red Hat, Inc.
package podhelper

import (
	"errors"
	"maps"
	"testing"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	k8sDynamicFake "k8s.io/client-go/dynamic/fake"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	k8stesting "k8s.io/client-go/testing"
)

func Test_followOwnerReferences(t *testing.T) {
	type args struct {
		topOwners map[string]TopOwner
		namespace string
		ownerRefs []metav1.OwnerReference
	}

	csv1 := &olmv1Alpha.ClusterServiceVersion{
		TypeMeta: metav1.TypeMeta{Kind: "ClusterServiceVersion", APIVersion: "operators.coreos.com/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "csv1",
			Namespace:       "ns1",
			OwnerReferences: []metav1.OwnerReference{},
		},
	}
	dep1 := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "dep1",
			Namespace:       "ns1",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "operators.coreos.com/v1alpha1", Kind: "ClusterServiceVersion", Name: "csv1"}},
		},
	}
	rep1 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{Kind: "ReplicaSet", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "rep1",
			Namespace:       "ns1",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Deployment", Name: "dep1"}},
		},
	}

	node1 := &corev1.Node{
		TypeMeta: metav1.TypeMeta{Kind: "Node", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
	}

	pod1 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "pod1",
			Namespace:       "ns1",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "v1", Kind: "Node", Name: "node1"}},
		},
	}

	resourceList := []*metav1.APIResourceList{
		{GroupVersion: "operators.coreos.com/v1alpha1", APIResources: []metav1.APIResource{{Name: "clusterserviceversions", Kind: "ClusterServiceVersion", Namespaced: true}}},
		{GroupVersion: "apps/v1", APIResources: []metav1.APIResource{{Name: "deployments", Kind: "Deployment", Namespaced: true}}},
		{GroupVersion: "apps/v1", APIResources: []metav1.APIResource{{Name: "replicasets", Kind: "ReplicaSet", Namespaced: true}}},
		{GroupVersion: "apps/v1", APIResources: []metav1.APIResource{{Name: "pods", Kind: "Pod", Namespaced: true}}},
		{GroupVersion: "v1", APIResources: []metav1.APIResource{{Name: "nodes", Kind: "Node", Namespaced: false}}},
		{GroupVersion: "v1", APIResources: []metav1.APIResource{{Name: "pods", Kind: "Pod", Namespaced: true}}},
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{topOwners: map[string]TopOwner{"csv1": {APIVersion: "operators.coreos.com/v1alpha1", Namespace: "ns1", Kind: "ClusterServiceVersion", Name: "csv1"}},
				namespace: "ns1",
				ownerRefs: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: "rep1"}},
			},
		},
		{
			name: "test2 - non-namespaced owner: pod owned a node",
			args: args{topOwners: map[string]TopOwner{"node1": {APIVersion: "v1", Namespace: "", Kind: "Node", Name: "node1"}},
				namespace: "ns1",
				ownerRefs: []metav1.OwnerReference{{APIVersion: "v1", Kind: "Pod", Name: "pod1"}},
			},
		},
	}

	scheme := runtime.NewScheme()
	// Add native resources to the scheme, otherwise, resources of APIVersion "v1" (not "core/v1") won't be found as unstructured resource in the type to GKV map here:
	// https://github.com/kubernetes/apimachinery/blob/96b97de8d6ba49bc192968551f2120ef3881f42d/pkg/runtime/scheme.go#L263
	err := k8sscheme.AddToScheme(scheme)
	if err != nil {
		t.Errorf("failed to ad k8s resources to scheme: %v", err)
	}

	client := k8sDynamicFake.NewSimpleDynamicClient(scheme, rep1, dep1, csv1, node1, pod1)

	// Spoof the get functions
	client.AddReactor("get", "ClusterServiceVersion", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, csv1, nil
	})
	client.AddReactor("get", "Deployment", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, dep1, nil
	})
	client.AddReactor("get", "ReplicaSet", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, rep1, nil
	})
	client.AddReactor("get", "Node", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, node1, nil
	})
	client.AddReactor("get", "Pod", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, pod1, nil
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResults := map[string]TopOwner{}
			if err := followOwnerReferences(resourceList, client, gotResults, tt.args.namespace, tt.args.ownerRefs); (err != nil) != tt.wantErr {
				t.Errorf("followOwnerReferences() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !maps.Equal(gotResults, tt.args.topOwners) {
				t.Errorf("followOwnerReferences() = %v, want %v", gotResults, tt.args.topOwners)
			}
		})
	}
}

func TestSearchAPIResource(t *testing.T) {
	testCases := []struct {
		testKind         string
		testAPIVersion   string
		testResourceList []*metav1.APIResourceList
		expectedError    error
	}{
		{ // Test Case #1 - APIResource found
			testKind:       "ClusterServiceVersion",
			testAPIVersion: "operators.coreos.com/v1alpha1",
			testResourceList: []*metav1.APIResourceList{
				{
					GroupVersion: "operators.coreos.com/v1alpha1",
					APIResources: []metav1.APIResource{
						{
							Name:       "clusterserviceversions",
							Kind:       "ClusterServiceVersion",
							Namespaced: true,
						},
					},
				},
			},
			expectedError: nil,
		},
		{ // Test Case #2 - APIResource not found
			testKind:       "ClusterServiceVersion",
			testAPIVersion: "operators.coreos.com/v1alpha1",
			testResourceList: []*metav1.APIResourceList{
				{
					GroupVersion: "operators.redhat-test.com/v1alpha1",
					APIResources: []metav1.APIResource{
						{
							Name:       "clusterserviceversions",
							Kind:       "ClusterServiceVersion",
							Namespaced: true,
						},
					},
				},
			},
			expectedError: errors.New("apiResource not found for kind=ClusterServiceVersion and APIVersion=operators.coreos.com/v1alpha1"),
		},
	}

	for _, tc := range testCases {
		resource, err := searchAPIResource(tc.testKind, tc.testAPIVersion, tc.testResourceList)
		if tc.expectedError != nil {
			assert.Equal(t, tc.expectedError, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tc.testKind, resource.Kind)
		}
	}
}

func TestGetTopOwners(t *testing.T) {
	generatePod := func(name, namespace string, ownerRefs []metav1.OwnerReference) *corev1.Pod {
		return &corev1.Pod{
			TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				Namespace:       namespace,
				OwnerReferences: ownerRefs,
			},
		}
	}

	generateDeployment := func(name, namespace string, ownerRefs []metav1.OwnerReference) *appsv1.Deployment {
		return &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				Namespace:       namespace,
				OwnerReferences: ownerRefs,
			},
		}
	}

	testDeployment := generateDeployment("dep1", "ns1", []metav1.OwnerReference{})
	testPod := generatePod("pod1", "ns1", []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Deployment", Name: "dep1"}})

	resourceList := []*metav1.APIResourceList{
		{GroupVersion: "apps/v1", APIResources: []metav1.APIResource{{Name: "deployments", Kind: "Deployment", Namespaced: true}}},
		{GroupVersion: "v1", APIResources: []metav1.APIResource{{Name: "pods", Kind: "Pod", Namespaced: true}}},
	}

	client := k8sDynamicFake.NewSimpleDynamicClient(k8sscheme.Scheme, testDeployment, testPod)

	// Spoof the get functions
	client.AddReactor("get", "Deployment", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, testDeployment, nil
	})

	client.AddReactor("get", "Pod", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, testPod, nil
	})

	// Set the test clients
	clientsholder.SetTestK8sDynamicClientsHolder(client)
	clientsholder.SetTestClientGroupResources(resourceList)

	// Get the top owner for the pod which is a deployment
	topOwners, err := GetPodTopOwner("ns1", testPod.OwnerReferences)
	assert.Nil(t, err)
	assert.Equal(t, map[string]TopOwner{"dep1": {APIVersion: "apps/v1", Namespace: "ns1", Kind: "Deployment", Name: "dep1"}}, topOwners)
}
