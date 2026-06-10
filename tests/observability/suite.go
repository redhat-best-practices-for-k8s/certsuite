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
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksadapter"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	checksfn "github.com/redhat-best-practices-for-k8s/checks/observability"

	apiserv1 "github.com/openshift/api/apiserver/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}
)

func LoadChecks() {
	log.Debug("Loading %s suite checks", common.ObservabilityTestKey)

	checksGroup := checksdb.NewChecksGroup(common.ObservabilityTestKey).
		WithBeforeEachFn(beforeEachFn)

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("observability-container-logging")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckContainerLogging).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("observability-crd-status")).
		WithSkipCheckFn(testhelper.GetNoCrdsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckCRDStatus).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("observability-termination-policy")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckTerminationPolicy).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("observability-pod-disruption-budget")).
		WithSkipCheckFn(testhelper.GetNoDeploymentsUnderTestSkipFn(&env), testhelper.GetNoStatefulSetsUnderTestSkipFn(&env)).
		WithSkipModeAll().
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckPodDisruptionBudget).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("observability-compatibility-with-next-ocp-release")).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckAPICompatibilityWithNextOCPRelease).MakeCheckFn(&env)))
}

// Function to build a map from workload service accounts
// to their associated to-be-deprecated APIs and the release version
// Filters:
// - status.removedInRelease is not empty
// - Verifies if the service account is inside the workload SA list from env.ServiceAccounts
func buildServiceAccountToDeprecatedAPIMap(apiRequestCounts []apiserv1.APIRequestCount, workloadServiceAccountNames map[string]struct{}) map[string]map[string]string {
	// Define a map where the key is the service account name and the value is another map
	// The inner map key is the API name and the value is the release version in which it will be removed
	serviceAccountToDeprecatedAPIs := make(map[string]map[string]string)

	for i := range apiRequestCounts {
		obj := &apiRequestCounts[i]
		// Filter by non-empty removedInRelease
		if obj.Status.RemovedInRelease != "" {
			// Iterate over the last 24h usage data
			for _, last24h := range obj.Status.Last24h {
				for _, byNode := range last24h.ByNode {
					for _, byUser := range byNode.ByUser {
						// Split the username by ":" and take the last chunk to extract ServiceAccount
						// from composed structures like system:serviceaccount:default:eventtest-operator-service-account
						serviceAccountParts := strings.Split(byUser.UserName, ":")
						strippedServiceAccount := serviceAccountParts[len(serviceAccountParts)-1]

						// Check if the service account is in the workload SA list
						if _, exists := workloadServiceAccountNames[strippedServiceAccount]; exists {
							// Initialize the inner map if it does not exist
							if serviceAccountToDeprecatedAPIs[strippedServiceAccount] == nil {
								serviceAccountToDeprecatedAPIs[strippedServiceAccount] = make(map[string]string)
							}
							// Add the API and its RemovedInRelease K8s version to the map
							serviceAccountToDeprecatedAPIs[strippedServiceAccount][obj.Name] = obj.Status.RemovedInRelease
						}
					}
				}
			}
		}
	}

	return serviceAccountToDeprecatedAPIs
}

// Evaluate workload API compliance with the next Kubernetes version
func evaluateAPICompliance(
	serviceAccountToDeprecatedAPIs map[string]map[string]string,
	kubernetesVersion string,
	workloadServiceAccountNames map[string]struct{}) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	version, err := semver.NewVersion(kubernetesVersion)
	if err != nil {
		fmt.Printf("Failed to parse Kubernetes version %q: %v", kubernetesVersion, err)
		return nil, nil
	}

	// Increment the version to represent the next release for comparison
	nextK8sVersion := version.IncMinor()

	// Iterate over each service account and its deprecated APIs
	for saName, deprecatedAPIs := range serviceAccountToDeprecatedAPIs {
		for apiName, removedInRelease := range deprecatedAPIs {
			removedVersion, err := semver.NewVersion(removedInRelease)
			if err != nil {
				fmt.Printf("Failed to parse Kubernetes version from APIRequestCount.status.removedInRelease: %s\n", err)
				// Skip this API if the version parsing fails
				continue
			}

			isCompliantWithNextK8sVersion := removedVersion.Minor() > nextK8sVersion.Minor()

			// Define reasons with version information
			nonCompliantReason := fmt.Sprintf("API %s used by service account %s is NOT compliant with Kubernetes version %s, it will be removed in release %s", apiName, saName, nextK8sVersion.String(), removedInRelease)
			compliantReason := fmt.Sprintf("API %s used by service account %s is compliant with Kubernetes version %s, it will be removed in release %s", apiName, saName, nextK8sVersion.String(), removedInRelease)

			var reportObject *testhelper.ReportObject
			if isCompliantWithNextK8sVersion {
				reportObject = testhelper.NewReportObject(compliantReason, "API", true)
				reportObject.AddField("ActiveInRelease", nextK8sVersion.String())
				compliantObjects = append(compliantObjects, reportObject)
			} else {
				reportObject = testhelper.NewReportObject(nonCompliantReason, "API", false)
				reportObject.AddField("RemovedInRelease", removedInRelease)
				nonCompliantObjects = append(nonCompliantObjects, reportObject)
			}

			reportObject.AddField("APIName", apiName)
			reportObject.AddField("ServiceAccount", saName)
		}
	}

	// Force the test to pass if both lists are empty
	if len(compliantObjects) == 0 && len(nonCompliantObjects) == 0 {
		for saName := range workloadServiceAccountNames {
			reportObject := testhelper.NewReportObject("SA does not use any removed API", "ServiceAccount", true).
				AddField("Name", saName)
			compliantObjects = append(compliantObjects, reportObject)
		}
	}

	return compliantObjects, nonCompliantObjects
}
