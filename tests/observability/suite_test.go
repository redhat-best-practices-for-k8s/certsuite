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

package observability

import (
	"testing"

	apiserv1 "github.com/openshift/api/apiserver/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildServiceAccountToDeprecatedAPIMap(t *testing.T) {
	// Example map of unique service account names for the workload
	workloadServiceAccountNames := map[string]struct{}{
		"builder":                            {},
		"default":                            {},
		"deployer":                           {},
		"eventtest-operator-service-account": {},
	}

	testCases := []struct {
		name                        string
		apiRequestCounts            []apiserv1.APIRequestCount
		workloadServiceAccountNames map[string]struct{}
		expected                    map[string]map[string]string
	}{
		{
			name: "Test to ensure proper mapping of service accounts to deprecated APIs",
			apiRequestCounts: []apiserv1.APIRequestCount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "api1",
					},
					Status: apiserv1.APIRequestCountStatus{
						RemovedInRelease: "v1.20",
						Last24h: []apiserv1.PerResourceAPIRequestLog{
							{
								ByNode: []apiserv1.PerNodeAPIRequestLog{
									{
										ByUser: []apiserv1.PerUserAPIRequestCount{
											{
												UserName:  "eventtest-operator-service-account",
												UserAgent: "agent1",
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "api2",
					},
					Status: apiserv1.APIRequestCountStatus{
						RemovedInRelease: "v1.21",
						Last24h: []apiserv1.PerResourceAPIRequestLog{
							{
								ByNode: []apiserv1.PerNodeAPIRequestLog{
									{
										ByUser: []apiserv1.PerUserAPIRequestCount{
											{
												UserName:  "unknown-sa",
												UserAgent: "agent2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			workloadServiceAccountNames: workloadServiceAccountNames,
			expected: map[string]map[string]string{
				"eventtest-operator-service-account": {
					"api1": "v1.20",
				},
			},
		},
		{
			name: "Test where no matching service account names exist",
			apiRequestCounts: []apiserv1.APIRequestCount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "api3",
					},
					Status: apiserv1.APIRequestCountStatus{
						RemovedInRelease: "v1.22",
						Last24h: []apiserv1.PerResourceAPIRequestLog{
							{
								ByNode: []apiserv1.PerNodeAPIRequestLog{
									{
										ByUser: []apiserv1.PerUserAPIRequestCount{
											{
												UserName:  "non-existent-sa",
												UserAgent: "agent3",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			workloadServiceAccountNames: workloadServiceAccountNames,
			expected:                    map[string]map[string]string{}, // Expect no output since 'non-existent-sa' is not in the SA map
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := buildServiceAccountToDeprecatedAPIMap(tc.apiRequestCounts, tc.workloadServiceAccountNames)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestEvaluateAPICompliance(t *testing.T) {
	// Mock service account to deprecated APIs map
	serviceAccountToDeprecatedAPIs := map[string]map[string]string{
		"service-account-1": {
			"api1": "1.21",
		},
		"service-account-2": {
			"api2": "1.20",
		},
	}

	// Mock current Kubernetes version
	kubernetesVersion := "1.19"

	// Mock workload service account names
	workloadServiceAccountNames := map[string]struct{}{
		"service-account-1": {},
		"service-account-2": {},
	}

	// Call the function
	compliantObjects, nonCompliantObjects := evaluateAPICompliance(serviceAccountToDeprecatedAPIs, kubernetesVersion, workloadServiceAccountNames)

	// Helper to create a ReportObject with fields
	createReportObject := func(reason, objType, apiName, removedInRelease, version, serviceAccount string, isCompliant bool) *testhelper.ReportObject {
		obj := testhelper.NewReportObject(reason, objType, isCompliant)
		obj.AddField("APIName", apiName)
		obj.AddField("ServiceAccount", serviceAccount)
		if isCompliant {
			obj.AddField("ActiveInRelease", version)
		} else {
			obj.AddField("RemovedInRelease", removedInRelease)
		}
		return obj
	}

	// Expected results
	expectedCompliantObjects := []*testhelper.ReportObject{
		// API removed in 1.21 is compliant with Kubernetes 1.20 = (current version 1.19 + 1)
		createReportObject(
			"API api1 used by service account service-account-1 is compliant with Kubernetes version 1.20.0, it will be removed in release 1.21",
			"API", "api1", "1.21", "1.20.0", "service-account-1", true,
		),
	}

	expectedNonCompliantObjects := []*testhelper.ReportObject{
		// API removed in 1.20 is non-compliant with Kubernetes 1.20 = (current version 1.19 + 1)
		createReportObject(
			"API api2 used by service account service-account-2 is NOT compliant with Kubernetes version 1.20.0, it will be removed in release 1.20",
			"API", "api2", "1.20", "", "service-account-2", false,
		),
	}

	// Helper function to compare ReportObjects
	compareReportObjects := func(expected, actual []*testhelper.ReportObject) {
		assert.Len(t, actual, len(expected))
		for i, obj := range actual {
			assert.Equal(t, expected[i].ObjectType, obj.ObjectType)

			for _, ofk := range expected[i].ObjectFieldsKeys {
				assert.Contains(t, obj.ObjectFieldsKeys, ofk)
			}

			for _, ofv := range expected[i].ObjectFieldsValues {
				assert.Contains(t, obj.ObjectFieldsValues, ofv)
			}
		}
	}

	// Verify the results
	compareReportObjects(expectedCompliantObjects, compliantObjects)
	compareReportObjects(expectedNonCompliantObjects, nonCompliantObjects)

	// Test for the empty lists situation
	// Call the function with an empty map
	emptyMap := map[string]map[string]string{}
	workloadServiceAccountNames = map[string]struct{}{
		"service-account-1": {},
		"service-account-2": {},
	}
	compliantObjects, nonCompliantObjects = evaluateAPICompliance(emptyMap, kubernetesVersion, workloadServiceAccountNames)

	// Expected result for the empty case
	expectedEmptyResult := []*testhelper.ReportObject{
		testhelper.NewReportObject(
			"SA does not use any removed API",
			"ServiceAccount", true,
		).AddField("Name", "service-account-1"),
		testhelper.NewReportObject(
			"SA does not use any removed API",
			"ServiceAccount", true,
		).AddField("Name", "service-account-2"),
	}

	// Verify the result for the empty map case
	assert.Len(t, compliantObjects, len(expectedEmptyResult))
	assert.Empty(t, nonCompliantObjects)
	for i, obj := range expectedEmptyResult {
		assert.Equal(t, obj.ObjectType, compliantObjects[i].ObjectType)

		keyFound := false
		for _, ofk := range obj.ObjectFieldsKeys {
			for _, ofk2 := range compliantObjects[i].ObjectFieldsKeys {
				if ofk == ofk2 {
					keyFound = true
					break
				}
			}
		}

		assert.True(t, keyFound)

		valueFound := false
		for _, ofv := range obj.ObjectFieldsValues {
			for _, ofv2 := range compliantObjects[i].ObjectFieldsValues {
				if ofv == ofv2 {
					valueFound = true
					break
				}
			}
		}

		assert.True(t, valueFound)
	}
}
