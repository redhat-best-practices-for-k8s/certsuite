// Copyright (C) 2022-2024 Red Hat, Inc.
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
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetGuaranteedPodsWithExclusiveCPUs(t *testing.T) {
	generateEnv := func(cpuLimit, cpuRequest, memLimit, memRequest, podName string) *TestEnvironment {
		return &TestEnvironment{
			Pods: []*Pod{
				{
					Containers: []*Container{
						{
							Container: &corev1.Container{
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"cpu":    resource.MustParse(cpuRequest),
										"memory": resource.MustParse(memRequest),
									},
									Limits: corev1.ResourceList{
										"cpu":    resource.MustParse(cpuLimit),
										"memory": resource.MustParse(memLimit),
									},
								},
							},
						},
					},
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: podName,
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testEnv      *TestEnvironment
		expectedPods []string
	}{
		{
			testEnv:      generateEnv(validCPULimit, validCPULimit, validMemLimit, validMemLimit, "pod1"),
			expectedPods: []string{"pod1"},
		},
		{
			testEnv:      generateEnv(invalidCPULimit1, validCPULimit, validMemLimit, validMemLimit, "pod1"), // invalid CPU limit
			expectedPods: []string{},
		},
	}

	for _, tc := range testCases {
		returnedPods := tc.testEnv.GetGuaranteedPodsWithExclusiveCPUs()
		for _, r := range returnedPods {
			assert.True(t, stringhelper.StringInSlice(tc.expectedPods, r.Name, false))
		}

		nonGuaranteedPods := tc.testEnv.GetNonGuaranteedPods()
		for _, r := range nonGuaranteedPods {
			assert.False(t, stringhelper.StringInSlice(tc.expectedPods, r.Name, false))
		}
	}
}

func TestGetGuaranteedPods(t *testing.T) {
	generateEnv := func(cpuLimit, cpuRequest, memLimit, memRequest, podName string) *TestEnvironment {
		return &TestEnvironment{
			Pods: []*Pod{
				{
					Containers: []*Container{
						{
							Container: &corev1.Container{
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"cpu":    resource.MustParse(cpuRequest),
										"memory": resource.MustParse(memRequest),
									},
									Limits: corev1.ResourceList{
										"cpu":    resource.MustParse(cpuLimit),
										"memory": resource.MustParse(memLimit),
									},
								},
							},
						},
					},
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: podName,
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testEnv      *TestEnvironment
		expectedPods []string
	}{
		{
			testEnv:      generateEnv(validCPULimit, validCPULimit, validMemLimit, validMemLimit, "pod1"),
			expectedPods: []string{"pod1"},
		},
		{
			testEnv:      generateEnv(invalidCPULimit1, validCPULimit, validMemLimit, validMemLimit, "pod1"), // invalid CPU limit
			expectedPods: []string{},
		},
	}

	for _, tc := range testCases {
		returnedPods := tc.testEnv.GetGuaranteedPods()
		for _, r := range returnedPods {
			assert.True(t, stringhelper.StringInSlice(tc.expectedPods, r.Name, false))
		}
	}
}

func TestAffinityRequiredFilters(t *testing.T) {
	generateEnv := func(affinityRequiredVal, podName string) *TestEnvironment {
		return &TestEnvironment{
			Pods: []*Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: podName,
							Labels: map[string]string{
								"AffinityRequired": affinityRequiredVal,
							},
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testEnv      *TestEnvironment
		expectedPods []string
	}{
		{
			testEnv:      generateEnv("true", "pod1"),
			expectedPods: []string{"pod1"},
		},
		{
			testEnv:      generateEnv("false", "pod1"),
			expectedPods: []string{},
		},
	}

	for _, tc := range testCases {
		returnedPods := tc.testEnv.GetAffinityRequiredPods()
		for _, r := range returnedPods {
			assert.True(t, stringhelper.StringInSlice(tc.expectedPods, r.Name, false))
		}

		p := tc.testEnv.GetPodsWithoutAffinityRequiredLabel()
		for _, r := range p {
			assert.False(t, stringhelper.StringInSlice(tc.expectedPods, r.Name, false))
		}
	}
}

func TestGetShareProcessNamespacePods(t *testing.T) {
	generateEnv := func(spn *bool, podName string) *TestEnvironment {
		return &TestEnvironment{
			Pods: []*Pod{
				{
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: podName,
						},
						Spec: corev1.PodSpec{
							ShareProcessNamespace: spn,
						},
					},
				},
			},
		}
	}

	trueVar := true
	falseVar := false

	testCases := []struct {
		testEnv      *TestEnvironment
		expectedPods []string
	}{
		{
			testEnv:      generateEnv(nil, "pod1"),
			expectedPods: []string{},
		},
		{
			testEnv:      generateEnv(&trueVar, "pod1"),
			expectedPods: []string{"pod1"},
		},
		{
			testEnv:      generateEnv(&falseVar, "pod1"),
			expectedPods: []string{},
		},
	}

	for _, tc := range testCases {
		returnedPods := tc.testEnv.GetShareProcessNamespacePods()
		for _, r := range returnedPods {
			assert.True(t, stringhelper.StringInSlice(tc.expectedPods, r.Name, false))
		}
	}
}

func TestGetHugepagesPods(t *testing.T) {
	generateEnv := func(podName string, resources, limits bool) *TestEnvironment {
		testEnv := &TestEnvironment{
			Pods: []*Pod{
				{
					Containers: []*Container{
						{
							Container: &corev1.Container{
								Name: "container1",
							},
						},
					},
					Pod: &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: podName,
						},
					},
				},
			},
		}

		if resources {
			testEnv.Pods[0].Containers[0].Resources.Requests = corev1.ResourceList{
				"hugepages": resource.MustParse(validCPULimit),
			}
		}

		if limits {
			testEnv.Pods[0].Containers[0].Resources.Limits = corev1.ResourceList{
				"hugepages": resource.MustParse(validCPULimit),
			}
		}

		return testEnv
	}

	testCases := []struct {
		testEnv      *TestEnvironment
		expectedPods []string
	}{
		{
			testEnv:      generateEnv("pod1", false, false),
			expectedPods: []string{},
		},
		{
			testEnv:      generateEnv("pod1", true, false),
			expectedPods: []string{"pod1"},
		},
		{
			testEnv:      generateEnv("pod1", false, true),
			expectedPods: []string{"pod1"},
		},
	}

	for _, tc := range testCases {
		returnedPods := tc.testEnv.GetHugepagesPods()
		for _, r := range returnedPods {
			assert.True(t, stringhelper.StringInSlice(tc.expectedPods, r.Name, false))
		}
	}
}
