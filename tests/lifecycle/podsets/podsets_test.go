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

package podsets

import (
	"fmt"
	"maps"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsDeploymentReady(t *testing.T) {
	type dpStatus struct {
		condition   appsv1.DeploymentConditionType
		replicas    int32
		ready       int32
		available   int32
		unavailable int32
		updated     int32
	}
	m := map[dpStatus]bool{
		{appsv1.DeploymentReplicaFailure, 10, 9, 10, 0, 0}: false,
		{appsv1.DeploymentAvailable, 10, 9, 9, 0, 10}:      false,
		{appsv1.DeploymentAvailable, 10, 10, 10, 1, 10}:    false,
		{appsv1.DeploymentAvailable, 10, 1, 10, 0, 10}:     false,
		{appsv1.DeploymentAvailable, 10, 10, 10, 0, 9}:     false,
		{appsv1.DeploymentAvailable, 10, 10, 10, 0, 10}:    true,
	}
	for key, v := range m {
		dp := provider.Deployment{
			Deployment: &appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Conditions: []appsv1.DeploymentCondition{
						{
							Type: key.condition,
						},
					},
					ReadyReplicas:       key.ready,
					AvailableReplicas:   key.available,
					UnavailableReplicas: key.unavailable,
					UpdatedReplicas:     key.updated,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &key.replicas,
				},
			},
		}
		ready := dp.IsDeploymentReady()
		assert.Equal(t, v, ready)
	}
}

func TestIsStatefulSetReady(t *testing.T) {
	type stStatus struct {
		replicas  int32
		ready     int32
		available int32
		updated   int32
		current   int32
	}
	m := map[stStatus]bool{
		{10, 9, 10, 0, 0}:    false,
		{10, 9, 9, 10, 0}:    false,
		{10, 10, 10, 11, 0}:  false,
		{10, 1, 10, 10, 3}:   false,
		{10, 10, 10, 9, 10}:  false,
		{10, 10, 10, 10, 10}: true,
	}
	for k, v := range m {
		ss := provider.StatefulSet{
			StatefulSet: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: &k.replicas,
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:     k.ready,
					AvailableReplicas: k.available,
					UpdatedReplicas:   k.updated,
					CurrentReplicas:   k.current,
				},
			},
		}
		ready := ss.IsStatefulSetReady()
		if ready != v {
			fmt.Println(" k= ", k, " should be ", v, " is ", ready)
		}
		assert.Equal(t, v, ready)
	}
}

func TestGetDeploymentsInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		deployments []*provider.Deployment
		want        []string
	}{
		{
			name: "normal list",
			deployments: []*provider.Deployment{
				{Deployment: &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "dep1", Namespace: "ns1"}}},
				{Deployment: &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "dep2", Namespace: "ns2"}}},
			},
			want: []string{"ns1:dep1", "ns2:dep2"},
		},
		{
			name:        "empty list",
			deployments: []*provider.Deployment{},
			want:        []string{},
		},
		{
			name: "single entry",
			deployments: []*provider.Deployment{
				{Deployment: &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "only", Namespace: "default"}}},
			},
			want: []string{"default:only"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDeploymentsInfo(tt.deployments)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetStatefulSetsInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		statefulSets []*provider.StatefulSet
		want         []string
	}{
		{
			name: "normal list",
			statefulSets: []*provider.StatefulSet{
				{StatefulSet: &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "sts1", Namespace: "ns1"}}},
				{StatefulSet: &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "sts2", Namespace: "ns2"}}},
			},
			want: []string{"ns1:sts1", "ns2:sts2"},
		},
		{
			name:         "empty list",
			statefulSets: []*provider.StatefulSet{},
			want:         []string{},
		},
		{
			name: "single entry",
			statefulSets: []*provider.StatefulSet{
				{StatefulSet: &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "only", Namespace: "default"}}},
			},
			want: []string{"default:only"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStatefulSetsInfo(tt.statefulSets)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPodSetNodes(t *testing.T) {
	type args struct {
		pods    []*corev1.Pod
		ssName  string
		nodesIn map[string]bool
	}
	tests := []struct {
		name string
		args args
		want map[string]bool
	}{
		{
			name: "ok",
			args: args{pods: []*corev1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "tnf",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "StatefulSet",
						Name: "stateful1",
					}},
				},
				Spec: corev1.PodSpec{
					NodeName: "node1",
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "tnf2",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "StatefulSet",
						Name: "stateful1",
					}},
				},
				Spec: corev1.PodSpec{
					NodeName: "node2",
				},
			}}, ssName: "stateful1", nodesIn: map[string]bool{}}, want: map[string]bool{"node1": true, "node2": true},
		},
		{
			name: "nok",
			args: args{pods: []*corev1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "tnf",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "StatefulSet",
						Name: "stateful1",
					}},
				},
				Spec: corev1.PodSpec{
					NodeName: "node1",
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "tnf2",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "DD",
						Name: "stateful1",
					}},
				},
				Spec: corev1.PodSpec{
					NodeName: "node2",
				},
			}}, ssName: "stateful1", nodesIn: map[string]bool{}}, want: map[string]bool{"node1": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAllNodesForAllPodSets(provider.ConvertArrayPods(tt.args.pods)); !maps.Equal(got, tt.want) {
				t.Errorf("GetAllNodesForAllPodSets() = %v, want %v", got, tt.want)
			}
		})
	}
}
