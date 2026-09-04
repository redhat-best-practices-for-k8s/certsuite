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
	"io"
	"testing"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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

func Test_checkForbiddenCapability(t *testing.T) {
	log.SetupLogger(io.Discard, "INFO")

	tests := []struct {
		name                  string
		containers            []*provider.Container
		capability            string
		wantCompliantCount    int
		wantNonCompliantCount int
	}{
		{
			name:                  "no containers",
			containers:            []*provider.Container{},
			capability:            "SYS_ADMIN",
			wantCompliantCount:    0,
			wantNonCompliantCount: 0,
		},
		{
			name: "container with nil SecurityContext",
			containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name:            "test-container",
						SecurityContext: nil,
					},
					Namespace: "test-ns",
					Podname:   "test-pod",
				},
			},
			capability:            "SYS_ADMIN",
			wantCompliantCount:    1,
			wantNonCompliantCount: 0,
		},
		{
			name: "container with nil Capabilities",
			containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name: "test-container",
						SecurityContext: &corev1.SecurityContext{
							Capabilities: nil,
						},
					},
					Namespace: "test-ns",
					Podname:   "test-pod",
				},
			},
			capability:            "NET_ADMIN",
			wantCompliantCount:    1,
			wantNonCompliantCount: 0,
		},
		{
			name: "container with forbidden capability",
			containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name: "test-container",
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{"SYS_ADMIN"},
							},
						},
					},
					Namespace: "test-ns",
					Podname:   "test-pod",
				},
			},
			capability:            "SYS_ADMIN",
			wantCompliantCount:    0,
			wantNonCompliantCount: 1,
		},
		{
			name: "compliant container without forbidden capability",
			containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name: "test-container",
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{"NET_RAW"},
							},
						},
					},
					Namespace: "test-ns",
					Podname:   "test-pod",
				},
			},
			capability:            "SYS_ADMIN",
			wantCompliantCount:    1,
			wantNonCompliantCount: 0,
		},
		{
			name: "multiple containers mixed compliance",
			containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name: "good-container",
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{"NET_RAW"},
							},
						},
					},
					Namespace: "test-ns",
					Podname:   "test-pod",
				},
				{
					Container: &corev1.Container{
						Name: "bad-container",
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{"SYS_ADMIN"},
							},
						},
					},
					Namespace: "test-ns",
					Podname:   "test-pod",
				},
			},
			capability:            "SYS_ADMIN",
			wantCompliantCount:    1,
			wantNonCompliantCount: 1,
		},
		{
			name: "container with SYS_MODULE is non-compliant",
			containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name: "test-container",
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{"SYS_MODULE"},
							},
						},
					},
					Namespace: "test-ns",
					Podname:   "test-pod",
				},
			},
			capability:            "SYS_MODULE",
			wantCompliantCount:    0,
			wantNonCompliantCount: 1,
		},
		{
			name: "container with DAC_OVERRIDE is non-compliant",
			containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name: "test-container",
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{"DAC_OVERRIDE"},
							},
						},
					},
					Namespace: "test-ns",
					Podname:   "test-pod",
				},
			},
			capability:            "DAC_OVERRIDE",
			wantCompliantCount:    0,
			wantNonCompliantCount: 1,
		},
		{
			name: "container with DAC_READ_SEARCH is non-compliant",
			containers: []*provider.Container{
				{
					Container: &corev1.Container{
						Name: "test-container",
						SecurityContext: &corev1.SecurityContext{
							Capabilities: &corev1.Capabilities{
								Add: []corev1.Capability{"DAC_READ_SEARCH"},
							},
						},
					},
					Namespace: "test-ns",
					Podname:   "test-pod",
				},
			},
			capability:            "DAC_READ_SEARCH",
			wantCompliantCount:    0,
			wantNonCompliantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.GetLogger()
			compliant, nonCompliant := checkForbiddenCapability(tt.containers, tt.capability, logger)
			assert.Len(t, compliant, tt.wantCompliantCount)
			assert.Len(t, nonCompliant, tt.wantNonCompliantCount)
		})
	}
}

func Test_isInstallModeMultiNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		installModes []v1alpha1.InstallMode
		want         bool
	}{
		{
			name:         "empty install modes",
			installModes: []v1alpha1.InstallMode{},
			want:         false,
		},
		{
			name: "AllNamespaces supported",
			installModes: []v1alpha1.InstallMode{
				{Type: v1alpha1.InstallModeTypeOwnNamespace, Supported: true},
				{Type: v1alpha1.InstallModeTypeAllNamespaces, Supported: true},
			},
			want: true,
		},
		{
			name: "AllNamespaces not present",
			installModes: []v1alpha1.InstallMode{
				{Type: v1alpha1.InstallModeTypeOwnNamespace, Supported: true},
				{Type: v1alpha1.InstallModeTypeSingleNamespace, Supported: true},
			},
			want: false,
		},
		{
			name: "only MultiNamespace present",
			installModes: []v1alpha1.InstallMode{
				{Type: v1alpha1.InstallModeTypeMultiNamespace, Supported: true},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isInstallModeMultiNamespace(tt.installModes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_isAllowedExtensionAPIServerRoleBinding(t *testing.T) {
	t.Parallel()

	olmPod := &provider.Pod{
		Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-ns",
				Labels: map[string]string{
					"olm.owner": "test-operator.v1.0.0",
				},
			},
		},
	}
	nonOLMPod := &provider.Pod{
		Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-ns",
				Labels:    map[string]string{"app": "my-app"},
			},
		},
	}

	tests := []struct {
		name string
		rb   *rbacv1.RoleBinding
		pod  *provider.Pod
		want bool
	}{
		{
			name: "correct namespace and name with OLM pod",
			rb: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "extension-apiserver-auth-reader",
					Namespace: extensionAPIServerNamespace,
				},
				RoleRef: rbacv1.RoleRef{
					Name: extensionAPIServerAuthReaderRoleBindingName,
				},
			},
			pod:  olmPod,
			want: true,
		},
		{
			name: "wrong namespace",
			rb: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "extension-apiserver-auth-reader",
					Namespace: "default",
				},
				RoleRef: rbacv1.RoleRef{
					Name: extensionAPIServerAuthReaderRoleBindingName,
				},
			},
			pod:  olmPod,
			want: false,
		},
		{
			name: "wrong role name",
			rb: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-role-binding",
					Namespace: extensionAPIServerNamespace,
				},
				RoleRef: rbacv1.RoleRef{
					Name: "some-other-role",
				},
			},
			pod:  olmPod,
			want: false,
		},
		{
			name: "pod not OLM-managed",
			rb: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "extension-apiserver-auth-reader",
					Namespace: extensionAPIServerNamespace,
				},
				RoleRef: rbacv1.RoleRef{
					Name: extensionAPIServerAuthReaderRoleBindingName,
				},
			},
			pod:  nonOLMPod,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedExtensionAPIServerRoleBinding(tt.rb, tt.pod)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_checkCrossNamespaceRoleBindingViolation(t *testing.T) {
	log.SetupLogger(io.Discard, "INFO")

	tests := []struct {
		name          string
		rb            *rbacv1.RoleBinding
		pod           *provider.Pod
		cnfNamespaces []string
		wantViolation bool
	}{
		{
			name: "subject in CNF namespace - allowed",
			rb: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-rb",
					Namespace: "cnf-ns",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      rbacv1.ServiceAccountKind,
						Name:      "test-sa",
						Namespace: "test-ns",
					},
				},
			},
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: "test-sa",
					},
				},
			},
			cnfNamespaces: []string{"test-ns", "cnf-ns"},
			wantViolation: false,
		},
		{
			name: "extension API server exception for OLM pod",
			rb: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "extension-apiserver-auth-reader",
					Namespace: extensionAPIServerNamespace,
				},
				RoleRef: rbacv1.RoleRef{
					Name: extensionAPIServerAuthReaderRoleBindingName,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      rbacv1.ServiceAccountKind,
						Name:      "test-sa",
						Namespace: "test-ns",
					},
				},
			},
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
						Labels: map[string]string{
							"olm.owner": "test-operator.v1.0.0",
						},
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: "test-sa",
					},
				},
			},
			cnfNamespaces: []string{"test-ns"},
			wantViolation: false,
		},
		{
			name: "cross-namespace violation detected",
			rb: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foreign-rb",
					Namespace: "foreign-ns",
				},
				RoleRef: rbacv1.RoleRef{
					Name: "some-role",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      rbacv1.ServiceAccountKind,
						Name:      "test-sa",
						Namespace: "test-ns",
					},
				},
			},
			pod: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "test-ns",
						Labels:    map[string]string{"app": "my-app"},
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: "test-sa",
					},
				},
			},
			cnfNamespaces: []string{"test-ns"},
			wantViolation: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := checksdb.NewCheck("test-check", []string{})
			result := checkCrossNamespaceRoleBindingViolation(tt.rb, tt.pod, tt.cnfNamespaces, check)
			if tt.wantViolation {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
