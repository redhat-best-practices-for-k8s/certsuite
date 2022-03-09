// Copyright (C) 2020-2022 Red Hat, Inc.
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

package podrecreation

import (
	"reflect"
	"testing"

	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func buildTestObjects() (testRuntimeObjects []runtime.Object) {
	// Replicaset Object
	aReplicaset := appv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "replicaset1",
			Namespace: "tnf",
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "Deployment",
				Name: "deployment1",
			}},
		},
	}
	testRuntimeObjects = append(testRuntimeObjects, &aReplicaset)
	return testRuntimeObjects
}

func TestGetDeploymentNodes(t *testing.T) { //nolint:funlen
	type args struct {
		pods  []*v1.Pod
		dName string
	}
	tests := []struct { //nolint:dupl
		name      string
		args      args
		wantNodes []string
	}{
		{
			name: "ok",
			args: args{pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "tnf",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "ReplicaSet",
						Name: "replicaset1",
					}},
				},
				Spec: v1.PodSpec{
					NodeName: "node1",
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "tnf",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "ReplicaSet",
						Name: "replicaset1",
					}},
				},
				Spec: v1.PodSpec{
					NodeName: "node2",
				},
			}}, dName: "deployment1"}, wantNodes: []string{"node1", "node2"},
		},
		{
			name: "nok",
			args: args{pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "tnf",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "ReplicaSet",
						Name: "replicaset2",
					}},
				},
				Spec: v1.PodSpec{
					NodeName: "node1",
				},
			}, {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "tnf",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "ReplicaSet",
						Name: "replicaset1",
					}},
				},
				Spec: v1.PodSpec{
					NodeName: "node2",
				},
			}}, dName: "deployment1"}, wantNodes: []string{"node2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNodes := GetDeploymentNodes(tt.args.pods, tt.args.dName, clientsholder.GetTestClientsHolder(buildTestObjects())); !reflect.DeepEqual(gotNodes, tt.wantNodes) {
				t.Errorf("GetDeploymentNodes() = %v, want %v", gotNodes, tt.wantNodes)
			}
		})
	}
}

func TestGetStatefulsetNodes(t *testing.T) { //nolint:funlen
	type args struct {
		pods   []*v1.Pod
		ssName string
	}
	tests := []struct { //nolint:dupl
		name      string
		args      args
		wantNodes []string
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
			}}, ssName: "stateful1"}, wantNodes: []string{"node1", "node2"},
		},
		{
			name: "nok",
			args: args{pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "tnf",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "StatefulSet",
						Name: "stateful2",
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
			}}, ssName: "stateful1"}, wantNodes: []string{"node2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotNodes := GetStatefulsetNodes(tt.args.pods, tt.args.ssName); !reflect.DeepEqual(gotNodes, tt.wantNodes) {
				t.Errorf("GetStatefulsetNodes() = %v, want %v", gotNodes, tt.wantNodes)
			}
		})
	}
}
