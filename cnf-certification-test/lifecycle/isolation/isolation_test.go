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

package isolation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	validCPULimit    = "1"
	validMemLimit    = "512Mi"
	invalidCPULimit1 = "0.5"
	invalidMemLimit1 = "64.5"
	invalidCPULimit2 = "2"
	invalidMemLimit2 = "65"
)

//nolint:funlen
func TestCPUIsolation(t *testing.T) {
	testClassName := "testRuntimeClassName"
	testCases := []struct {
		testPod                  *provider.Pod
		resourcesIdenticalResult bool
		wholeUnitsResult         bool
		runtimeClassNameResult   bool
		loadBalancingResult      bool
	}{
		{ // Test Case #1 - Happy Path, all CPU isolation requirements are met
			testPod: &provider.Pod{
				Containers: []*provider.Container{
					{
						Data: &v1.Container{
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse(validCPULimit),
									"memory": resource.MustParse(validMemLimit),
								},
								Limits: v1.ResourceList{
									"cpu":    resource.MustParse(validCPULimit),
									"memory": resource.MustParse(validMemLimit),
								},
							},
						},
					},
				},
				Data: &v1.Pod{
					Spec: v1.PodSpec{
						RuntimeClassName: &testClassName,
					},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"cpu-load-balancing.crio.io": "disable",
							"irq-load-balancing.crio.io": "disable",
						},
					},
				},
			},
			resourcesIdenticalResult: true,
			wholeUnitsResult:         true,
			runtimeClassNameResult:   true,
			loadBalancingResult:      true,
		},
		{ // Test Case #2 - Resources not identical
			testPod: &provider.Pod{
				Containers: []*provider.Container{
					{
						Data: &v1.Container{
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse(validCPULimit),
									"memory": resource.MustParse(validMemLimit),
								},
								Limits: v1.ResourceList{
									"cpu":    resource.MustParse(invalidCPULimit2),
									"memory": resource.MustParse(invalidMemLimit2),
								},
							},
						},
					},
				},
				Data: &v1.Pod{
					Spec: v1.PodSpec{
						RuntimeClassName: &testClassName,
					},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"cpu-load-balancing.crio.io": "disable",
							"irq-load-balancing.crio.io": "disable",
						},
					},
				},
			},
			resourcesIdenticalResult: false,
			wholeUnitsResult:         true,
			runtimeClassNameResult:   true,
			loadBalancingResult:      true,
		},
		{ // Test Case #3 - Resources are not whole units
			testPod: &provider.Pod{
				Containers: []*provider.Container{
					{
						Data: &v1.Container{
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse(invalidCPULimit1),
									"memory": resource.MustParse(invalidMemLimit1),
								},
								Limits: v1.ResourceList{
									"cpu":    resource.MustParse(invalidCPULimit1),
									"memory": resource.MustParse(invalidMemLimit1),
								},
							},
						},
					},
				},
				Data: &v1.Pod{
					Spec: v1.PodSpec{
						RuntimeClassName: &testClassName,
					},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"cpu-load-balancing.crio.io": "disable",
							"irq-load-balancing.crio.io": "disable",
						},
					},
				},
			},
			resourcesIdenticalResult: true,
			wholeUnitsResult:         false,
			runtimeClassNameResult:   true,
			loadBalancingResult:      true,
		},
		{ // Test Case #4 - runtimeClassName not set
			testPod: &provider.Pod{
				Containers: []*provider.Container{
					{
						Data: &v1.Container{
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse(validCPULimit),
									"memory": resource.MustParse(validMemLimit),
								},
								Limits: v1.ResourceList{
									"cpu":    resource.MustParse(validCPULimit),
									"memory": resource.MustParse(validMemLimit),
								},
							},
						},
					},
				},
				Data: &v1.Pod{
					Spec: v1.PodSpec{
						RuntimeClassName: nil,
					},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"cpu-load-balancing.crio.io": "disable",
							"irq-load-balancing.crio.io": "disable",
						},
					},
				},
			},
			resourcesIdenticalResult: true,
			wholeUnitsResult:         true,
			runtimeClassNameResult:   false,
			loadBalancingResult:      true,
		},
		{ // Test Case #5 - Annotations not set
			testPod: &provider.Pod{
				Containers: []*provider.Container{
					{
						Data: &v1.Container{
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse(validCPULimit),
									"memory": resource.MustParse(validMemLimit),
								},
								Limits: v1.ResourceList{
									"cpu":    resource.MustParse(validCPULimit),
									"memory": resource.MustParse(validMemLimit),
								},
							},
						},
					},
				},
				Data: &v1.Pod{
					Spec: v1.PodSpec{
						RuntimeClassName: &testClassName,
					},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"cpu-load-balancing.crio.io": "enable",
							"irq-load-balancing.crio.io": "disable",
						},
					},
				},
			},
			resourcesIdenticalResult: true,
			wholeUnitsResult:         true,
			runtimeClassNameResult:   true,
			loadBalancingResult:      false,
		},
	}

	for _, tc := range testCases {
		it := NewCPUIsolationTester(tc.testPod)
		assert.Equal(t, tc.resourcesIdenticalResult, it.AreResourcesIdentical())
		assert.Equal(t, tc.wholeUnitsResult, it.AreCPUResourcesWholeUnits())
		assert.Equal(t, tc.runtimeClassNameResult, it.IsRuntimeClassNameSpecified())
		assert.Equal(t, tc.loadBalancingResult, it.LoadBalancingDisabled())
	}
}
