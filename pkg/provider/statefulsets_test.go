// Copyright (C) 2022 Red Hat, Inc.
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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:funlen
func TestIsAffinityCompliantStatefulSets(t *testing.T) {
	testCases := []struct {
		testStatefulSet StatefulSet
		resultErrStr    error
		isCompliant     bool
	}{
		{ // Test Case #1 - Affinity is nil, AffinityRequired label is set, fail
			testStatefulSet: StatefulSet{
				StatefulSet: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							AffinityRequiredKey: "true",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Affinity: nil,
							},
						},
					},
				},
			},
			resultErrStr: errors.New("has been found with an AffinityRequired flag but is missing corresponding affinity rules"),
			isCompliant:  false,
		},
		{ // Test Case #2 - Affinity is not nil, but PodAffinity/NodeAffinity are also not set, fail
			testStatefulSet: StatefulSet{
				StatefulSet: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							AffinityRequiredKey: "true",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Affinity: &corev1.Affinity{}, // not nil
							},
						},
					},
				},
			},
			resultErrStr: errors.New("has been found with an AffinityRequired flag but is missing corresponding pod/node affinity rules"),
			isCompliant:  false,
		},
		{ // Test Case #3 - Affinity is not nil, but anti-affinity rule is set which defeats the purpose of the Required flag
			testStatefulSet: StatefulSet{
				StatefulSet: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							AffinityRequiredKey: "true",
						},
					},
					Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Affinity: &corev1.Affinity{
									PodAntiAffinity: &corev1.PodAntiAffinity{}, // anti-affinity set
								},
							},
						},
					},
				},
			},
			resultErrStr: errors.New("has been found with an AffinityRequired flag but has anti-affinity rules"),
			isCompliant:  false,
		},
	}

	for _, tc := range testCases {
		result, testErr := tc.testStatefulSet.IsAffinityCompliant()
		assert.Contains(t, testErr.Error(), tc.resultErrStr.Error())
		assert.Equal(t, tc.isCompliant, result)
	}
}

func TestStatefulsetToString(t *testing.T) {
	ss := StatefulSet{
		StatefulSet: &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "testNS",
			},
		},
	}

	assert.Equal(t, "statefulset: test1 ns: testNS", ss.ToString())
}
