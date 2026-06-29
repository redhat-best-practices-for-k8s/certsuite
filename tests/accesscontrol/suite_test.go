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

package accesscontrol

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_isContainerCapabilitySet(t *testing.T) {
	type args struct {
		containerCapabilities *corev1.Capabilities
		capability            string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil capabilities",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: nil,
			},
			want: false,
		},
		{
			name: "empty capabilities",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{},
			},
			want: false,
		},
		{
			name: "explicitly empty add list",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{Add: []corev1.Capability{}},
			},
			want: false,
		},
		{
			name: "explicitly empty drop list",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{Drop: []corev1.Capability{}},
			},
			want: false,
		},
		{
			name: "IPC_LOCK not found in any list",
			args: args{
				capability: "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{
					Add:  []corev1.Capability{"NET_CAP_BINDING"},
					Drop: []corev1.Capability{"SYS_ADMIN", "NET_ADMIN"},
				},
			},
			want: false,
		},
		{
			name: "IPC_LOCK found in the add list",
			args: args{
				capability: "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{
					Add:  []corev1.Capability{"NET_ADMIN", "IPC_LOCK"},
					Drop: []corev1.Capability{"SYS_ADMIN"}},
			},
			want: true,
		},
		{
			name: "IPC_LOCK appears in the drop list only",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{Drop: []corev1.Capability{"SYS_ADMIN", "IPC_LOCK", "NET_ADMIN"}},
			},
			want: false,
		},
		{
			// When set in both add and drop lists, k8s/openshift will compute drop first, then add, which results
			// in the capability to be finally set.
			name: "IPC_LOCK set in both add and drop lists.",
			args: args{
				capability: "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{
					Add:  []corev1.Capability{"IPC_LOCK"},
					Drop: []corev1.Capability{"SYS_ADMIN", "IPC_LOCK", "NET_ADMIN"},
				},
			},
			want: true,
		},
		{
			name: "ALL capabilities in the add list",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{Add: []corev1.Capability{"ALL"}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isContainerCapabilitySet(tt.args.containerCapabilities, tt.args.capability); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_isOwnedByOLM(t *testing.T) {
	tests := []struct {
		name string
		pod  *provider.Pod
		want bool
	}{
		{
			name: "pod with olm.owner label",
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
						Labels: map[string]string{
							"olm.owner": "test-operator.v1.0.0",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "pod with olm.owner.namespace label",
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
						Labels: map[string]string{
							"olm.owner.namespace": "openshift-operators",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "pod with olm.owner.kind label",
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
						Labels: map[string]string{
							"olm.owner.kind": "ClusterServiceVersion",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "pod with ClusterServiceVersion owner reference",
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								Kind:       "ClusterServiceVersion",
								Name:       "test-operator.v1.0.0",
								APIVersion: "operators.coreos.com/v1alpha1",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "pod with no OLM labels or owner references",
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
						Labels: map[string]string{
							"app": "my-app",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "pod with non-CSV owner reference",
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
						OwnerReferences: []metav1.OwnerReference{
							{
								Kind:       "ReplicaSet",
								Name:       "test-rs",
								APIVersion: "apps/v1",
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "pod with multiple OLM labels",
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
						Labels: map[string]string{
							"olm.owner":           "test-operator.v1.0.0",
							"olm.owner.namespace": "openshift-operators",
							"olm.owner.kind":      "ClusterServiceVersion",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "pod with both OLM label and CSV owner reference",
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
						Labels: map[string]string{
							"olm.owner": "test-operator.v1.0.0",
						},
						OwnerReferences: []metav1.OwnerReference{
							{
								Kind:       "ClusterServiceVersion",
								Name:       "test-operator.v1.0.0",
								APIVersion: "operators.coreos.com/v1alpha1",
							},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOwnedByOLM(tt.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}
