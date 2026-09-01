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

package lifecycle

import (
	"strings"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setupCheck() *checksdb.Check {
	var logArchive strings.Builder
	log.SetupLogger(&logArchive, "INFO")
	return checksdb.NewCheck("test-id", nil)
}

func strPtr(s string) *string {
	return &s
}

func TestPodNodeSelectorAndAffinityBestPractices(t *testing.T) {
	testCases := []struct {
		name           string
		pods           []*provider.Pod
		expectedResult checksdb.CheckResult
	}{
		{
			name: "pod with no nodeSelector and no affinity is compliant",
			pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
						Spec:       corev1.PodSpec{},
					},
				},
			},
			expectedResult: checksdb.CheckResultPassed,
		},
		{
			name: "pod with nodeSelector and no runtimeClassName is non-compliant",
			pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
						Spec: corev1.PodSpec{
							NodeSelector: map[string]string{"node-role.kubernetes.io/master": ""},
						},
					},
				},
			},
			expectedResult: checksdb.CheckResultFailed,
		},
		{
			name: "pod with nodeSelector AND runtimeClassName is compliant (RuntimeClass injection)",
			pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "upf-testpod-0-0", Namespace: "upf1"},
						Spec: corev1.PodSpec{
							RuntimeClassName: strPtr("performance-performance-master"),
							NodeSelector:     map[string]string{"node-role.kubernetes.io/master": ""},
						},
					},
				},
			},
			expectedResult: checksdb.CheckResultPassed,
		},
		{
			name: "pod with nodeAffinity and no runtimeClassName is non-compliant",
			pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
						Spec: corev1.PodSpec{
							Affinity: &corev1.Affinity{
								NodeAffinity: &corev1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
										NodeSelectorTerms: []corev1.NodeSelectorTerm{
											{
												MatchExpressions: []corev1.NodeSelectorRequirement{
													{Key: "node-role.kubernetes.io/worker", Operator: corev1.NodeSelectorOpExists},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResult: checksdb.CheckResultFailed,
		},
		{
			name: "pod with nodeAffinity AND runtimeClassName still reports affinity non-compliance",
			pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
						Spec: corev1.PodSpec{
							RuntimeClassName: strPtr("performance-worker"),
							Affinity: &corev1.Affinity{
								NodeAffinity: &corev1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
										NodeSelectorTerms: []corev1.NodeSelectorTerm{
											{
												MatchExpressions: []corev1.NodeSelectorRequirement{
													{Key: "node-role.kubernetes.io/worker", Operator: corev1.NodeSelectorOpExists},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResult: checksdb.CheckResultFailed,
		},
		{
			name: "pod with both nodeSelector and runtimeClassName plus affinity reports failure for affinity only",
			pods: []*provider.Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
						Spec: corev1.PodSpec{
							RuntimeClassName: strPtr("performance-performance-master"),
							NodeSelector:     map[string]string{"node-role.kubernetes.io/master": ""},
							Affinity: &corev1.Affinity{
								NodeAffinity: &corev1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
										NodeSelectorTerms: []corev1.NodeSelectorTerm{
											{
												MatchExpressions: []corev1.NodeSelectorRequirement{
													{Key: "custom-key", Operator: corev1.NodeSelectorOpExists},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedResult: checksdb.CheckResultFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			check := setupCheck()
			testPodNodeSelectorAndAffinityBestPractices(tc.pods, check)
			assert.Equal(t, tc.expectedResult, check.Result)
		})
	}
}
