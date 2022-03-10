// Copyright (C) 2020-2021 Red Hat, Inc.
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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v1app "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsDeploymentRead(t *testing.T) {
	type dpStatus struct {
		condition   v1app.DeploymentConditionType
		replicas    int32
		ready       int32
		available   int32
		unavailable int32
		updated     int32
	}
	m := map[dpStatus]bool{
		{v1app.DeploymentReplicaFailure, 10, 9, 10, 0, 0}: false,
		{v1app.DeploymentAvailable, 10, 9, 9, 0, 10}:      false,
		{v1app.DeploymentAvailable, 10, 10, 10, 1, 10}:    false,
		{v1app.DeploymentAvailable, 10, 1, 10, 0, 10}:     false,
		{v1app.DeploymentAvailable, 10, 10, 10, 0, 9}:     false,
		{v1app.DeploymentAvailable, 10, 10, 10, 0, 10}:    true,
	}
	for key, v := range m {
		dp := v1app.Deployment{}
		dpCondition := v1app.DeploymentCondition{Type: key.condition}
		dp.Status.Conditions = append(dp.Status.Conditions, dpCondition)
		dp.Spec.Replicas = &(key.replicas)
		dp.Status.ReadyReplicas = key.ready
		dp.Status.AvailableReplicas = key.available
		dp.Status.UnavailableReplicas = key.unavailable
		dp.Status.UpdatedReplicas = key.updated
		ready := IsDeploymentReady(&dp)
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
		statefulset := v1app.StatefulSet{}
		statefulset.Spec.Replicas = &(k.replicas)
		statefulset.Status.ReadyReplicas = k.ready
		statefulset.Status.AvailableReplicas = k.available
		statefulset.Status.UpdatedReplicas = k.updated
		statefulset.Status.CurrentReplicas = k.current
		ready := IsStatefulSetReady(&statefulset)
		if ready != v {
			fmt.Println(" k= ", k, " should be ", v, " is ", ready)
		}
		assert.Equal(t, v, ready)
	}
}

func TestGetPodSetNodes(t *testing.T) { //nolint:funlen
	type args struct {
		pods    []*v1.Pod
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
			args: args{pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "tnf",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "StatefulSet",
						Name: "stateful1",
					}},
				},
				Spec: v1.PodSpec{
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
				Spec: v1.PodSpec{
					NodeName: "node2",
				},
			}}, ssName: "stateful1", nodesIn: map[string]bool{}}, want: map[string]bool{"node1": true, "node2": true},
		},
		{
			name: "nok",
			args: args{pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "tnf",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "StatefulSet",
						Name: "stateful1",
					}},
				},
				Spec: v1.PodSpec{
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
				Spec: v1.PodSpec{
					NodeName: "node2",
				},
			}}, ssName: "stateful1", nodesIn: map[string]bool{}}, want: map[string]bool{"node1": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPodSetNodes(tt.args.pods, tt.args.ssName, tt.args.nodesIn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPodSetNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}
