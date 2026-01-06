// Copyright (C) 2022-2026 Red Hat, Inc.
package resources

import (
	"strings"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	validCPULimit = "1"
	validMemLimit = "512Mi"
)

func TestHasRequestsSet(t *testing.T) {
	testCases := []struct {
		testContainer  *provider.Container
		expectedResult bool
	}{
		{ // Test Case #1 - Happy path, requests are set
			testContainer: &provider.Container{
				Container: &corev1.Container{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse(validCPULimit),
							"memory": resource.MustParse(validMemLimit),
						},
					},
				},
			},
			expectedResult: true,
		},
		{ // Test Case #2 - Success even with limits present (we only check requests now)
			testContainer: &provider.Container{
				Container: &corev1.Container{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse(validCPULimit),
							"memory": resource.MustParse(validMemLimit),
						},
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse(validCPULimit),
							"memory": resource.MustParse(validMemLimit),
						},
					},
				},
			},
			expectedResult: true,
		},
		{ // Test Case #3 - Failure due to missing memory request
			testContainer: &provider.Container{
				Container: &corev1.Container{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu": resource.MustParse(validCPULimit),
						},
					},
				},
			},
			expectedResult: false,
		},
		{ // Test Case #4 - Failure due to missing CPU request
			testContainer: &provider.Container{
				Container: &corev1.Container{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"memory": resource.MustParse(validMemLimit),
						},
					},
				},
			},
			expectedResult: false,
		},
		{ // Test Case #5 - Failure due to missing resources in general
			testContainer: &provider.Container{
				Container: &corev1.Container{},
			},
			expectedResult: false,
		},
	}

	var logArchive strings.Builder
	log.SetupLogger(&logArchive, "INFO")
	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, HasRequestsSet(tc.testContainer, log.GetLogger()))
	}
}

func TestHasExclusiveCPUsAssigned(t *testing.T) {
	testCases := []struct {
		testContainer  *provider.Container
		expectedResult bool
	}{
		{ // Test Case #1 - Exclusive CPU pool selected
			testContainer: &provider.Container{
				Container: &corev1.Container{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("2"),
							"memory": resource.MustParse("512Mi"),
						},
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("2"),
							"memory": resource.MustParse("512Mi"),
						},
					},
				},
			},
			expectedResult: true,
		},
		{ // Test Case #2 - Shared CPU pool selected (requests and limits are not equal)
			testContainer: &provider.Container{
				Container: &corev1.Container{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("1"),
							"memory": resource.MustParse("512Mi"),
						},
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("2"),
							"memory": resource.MustParse("512Mi"),
						},
					},
				},
			},
			expectedResult: false,
		},
		{ // Test Case #3 - Shared CPU pool selected (requests and limits quantities are not an integer)
			testContainer: &provider.Container{
				Container: &corev1.Container{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("500m"),
							"memory": resource.MustParse("512Mi"),
						},
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("500m"),
							"memory": resource.MustParse("512Mi"),
						},
					},
				},
			},
			expectedResult: false,
		},
		{ // Test Case #4 - Shared CPU pool selected (requests and limits quantities specified as a fractional value)
			testContainer: &provider.Container{
				Container: &corev1.Container{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("1.5"),
							"memory": resource.MustParse("512Mi"),
						},
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("1.5"),
							"memory": resource.MustParse("512Mi"),
						},
					},
				},
			},
			expectedResult: false,
		},
		{ // Test Case #5 - Shared CPU pool selected (no requests or limits set)
			testContainer: &provider.Container{
				Container: &corev1.Container{},
			},
			expectedResult: false,
		},
		{ // Test Case #6 - Shared CPU pool selected (memory requests and limits are not equal)
			testContainer: &provider.Container{
				Container: &corev1.Container{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("2"),
							"memory": resource.MustParse("512Mi"),
						},
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("2"),
							"memory": resource.MustParse("256Mi"),
						},
					},
				},
			},
			expectedResult: false,
		},
		{ // Test Case #7 - Shared CPU pool selected (no memory limits specified)
			testContainer: &provider.Container{
				Container: &corev1.Container{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu": resource.MustParse("2"),
						},
						Limits: corev1.ResourceList{
							"cpu": resource.MustParse("2"),
						},
					},
				},
			},
			expectedResult: false,
		},
	}

	var logArchive strings.Builder
	log.SetupLogger(&logArchive, "INFO")
	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, HasExclusiveCPUsAssigned(tc.testContainer, log.GetLogger()))
	}
}
