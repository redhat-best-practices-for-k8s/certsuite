// Copyright (C) 2020-2024 Red Hat, Inc.
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
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/ocplite"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	pdbv1 "github.com/redhat-best-practices-for-k8s/certsuite/tests/observability/pdb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestLoggingIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersLogging(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestCrdsStatusSubresourceIdentifier)).
		WithSkipCheckFn(testhelper.GetNoCrdsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testCrds(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestTerminationMessagePolicyIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testTerminationMessagePolicy(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodDisruptionBudgetIdentifier)).
		WithSkipCheckFn(testhelper.GetNoDeploymentsUnderTestSkipFn(&env), testhelper.GetNoStatefulSetsUnderTestSkipFn(&env)).
		WithSkipModeAll().
		WithCheckFn(func(c *checksdb.Check) error {
			testPodDisruptionBudgets(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestAPICompatibilityWithNextOCPReleaseIdentifier)).
		WithCheckFn(func(c *checksdb.Check) error {
			testAPICompatibilityWithNextOCPRelease(c, &env)
			return nil
		}))
}

// containerHasLoggingOutput helper function to get the last line of logging output from
// a container. Returns true in case some output was found, false otherwise.
func containerHasLoggingOutput(cut *provider.Container) (bool, error) {
	ocpClient := clientsholder.GetClientsHolder()

	// K8s' API will not return lines that do not have the newline termination char, so
	// We need to ask for the last two lines.
	const tailLogLines = 2
	numLogLines := int64(tailLogLines)
	podLogOptions := corev1.PodLogOptions{TailLines: &numLogLines, Container: cut.Name}
	req := ocpClient.K8sClient.CoreV1().Pods(cut.Namespace).GetLogs(cut.Podname, &podLogOptions)

	podLogsReaderCloser, err := req.Stream(context.TODO())
	if err != nil {
		return false, fmt.Errorf("unable to get log streamer, err: %v", err)
	}

	defer podLogsReaderCloser.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogsReaderCloser)
	if err != nil {
		return false, fmt.Errorf("unable to get log data, err: %v", err)
	}

	return buf.String() != "", nil
}

func testContainersLogging(check *checksdb.Check, env *provider.TestEnvironment) {
	// Iterate through all the CUTs to get their log output. The TC checks that at least
	// one log line is found.
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		hasLoggingOutput, err := containerHasLoggingOutput(cut)
		if err != nil {
			check.LogError("Failed to get %q log output, err: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Could not get log output", false))
			continue
		}

		if !hasLoggingOutput {
			check.LogError("Container %q does not have any line of log to stderr/stdout", cut)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "No log line to stderr/stdout found", false))
		} else {
			check.LogInfo("Container %q has some logging output", cut)
			compliantObjects = append(compliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Found log line to stderr/stdout", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testCrds testing if crds have a status sub resource set
func testCrds(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, crd := range env.Crds {
		check.LogInfo("Testing CRD: %s", crd.Name)
		for _, ver := range crd.Spec.Versions {
			if _, ok := ver.Schema.OpenAPIV3Schema.Properties["status"]; !ok {
				check.LogError("CRD: %s, version: %s does not have a status subresource", crd.Name, ver.Name)
				nonCompliantObjects = append(nonCompliantObjects,
					testhelper.NewReportObject("Crd does not have a status sub resource set", testhelper.CustomResourceDefinitionType, false).
						AddField(testhelper.CustomResourceDefinitionName, crd.Name).
						AddField(testhelper.CustomResourceDefinitionVersion, ver.Name))
			} else {
				check.LogInfo("CRD: %s, version: %s has a status subresource", crd.Name, ver.Name)
				compliantObjects = append(compliantObjects,
					testhelper.NewReportObject("Crd has a status sub resource set", testhelper.CustomResourceDefinitionType, true).
						AddField(testhelper.CustomResourceDefinitionName, crd.Name).
						AddField(testhelper.CustomResourceDefinitionVersion, ver.Name))
			}
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testTerminationMessagePolicy tests to make sure that pods
func testTerminationMessagePolicy(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		if cut.TerminationMessagePolicy != corev1.TerminationMessageFallbackToLogsOnError {
			check.LogError("Container %q does not have a TerminationMessagePolicy: FallbackToLogsOnError (has %s)", cut, cut.TerminationMessagePolicy)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "TerminationMessagePolicy is not FallbackToLogsOnError", false))
		} else {
			check.LogInfo("Container %q has a TerminationMessagePolicy: FallbackToLogsOnError", cut)
			compliantObjects = append(compliantObjects,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "TerminationMessagePolicy is FallbackToLogsOnError", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

//nolint:funlen
func testPodDisruptionBudgets(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Loop through all of the of Deployments and StatefulSets and check if the PDBs are valid
	for _, d := range env.Deployments {
		check.LogInfo("Testing Deployment %q", d.ToString())
		deploymentSelector := labels.Set(d.Spec.Template.Labels)
		pdbFound := false
		for pdbIndex := range env.PodDisruptionBudgets {
			pdb := &env.PodDisruptionBudgets[pdbIndex]
			if pdb.Namespace != d.Namespace {
				continue
			}
			pdbSelector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
			if err != nil {
				check.LogError("Could not convert the PDB %q label selector to selector, err: %v", pdbSelector, err)
				continue
			}
			if pdbSelector.Matches(deploymentSelector) {
				pdbFound = true
				if ok, err := pdbv1.CheckPDBIsValid(pdb, d.Spec.Replicas); !ok {
					check.LogError("PDB %q is not valid for Deployment %q, err: %v", pdb.Name, d.Name, err)
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject(fmt.Sprintf("Invalid PodDisruptionBudget config: %v", err), testhelper.DeploymentType, false).
						AddField(testhelper.DeploymentName, d.Name).
						AddField(testhelper.Namespace, d.Namespace).
						AddField(testhelper.PodDisruptionBudgetReference, pdb.Name))
				} else {
					check.LogInfo("PDB %q is valid for Deployment: %q", pdb.Name, d.Name)
					compliantObjects = append(compliantObjects, testhelper.NewReportObject("Deployment: references PodDisruptionBudget", testhelper.DeploymentType, true).
						AddField(testhelper.DeploymentName, d.Name).
						AddField(testhelper.Namespace, d.Namespace).
						AddField(testhelper.PodDisruptionBudgetReference, pdb.Name))
				}
			}
		}
		if !pdbFound {
			check.LogError("Deployment %q is missing a corresponding PodDisruptionBudget", d.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("Deployment is missing a corresponding PodDisruptionBudget", testhelper.DeploymentType, false).
				AddField(testhelper.DeploymentName, d.Name).
				AddField(testhelper.Namespace, d.Namespace))
		}
	}

	for _, s := range env.StatefulSets {
		check.LogInfo("Testing StatefulSet %q", s.ToString())
		statefulSetSelector := labels.Set(s.Spec.Template.Labels)
		pdbFound := false
		for pdbIndex := range env.PodDisruptionBudgets {
			pdb := &env.PodDisruptionBudgets[pdbIndex]
			if pdb.Namespace != s.Namespace {
				continue
			}
			pdbSelector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
			if err != nil {
				check.LogError("Could not convert the PDB %q label selector to selector, err: %v", pdbSelector, err)
				continue
			}
			if pdbSelector.Matches(statefulSetSelector) {
				pdbFound = true
				if ok, err := pdbv1.CheckPDBIsValid(pdb, s.Spec.Replicas); !ok {
					check.LogError("PDB %q is not valid for StatefulSet %q, err: %v", pdb.Name, s.Name, err)
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject(fmt.Sprintf("Invalid PodDisruptionBudget config: %v", err), testhelper.StatefulSetType, false).
						AddField(testhelper.StatefulSetName, s.Name).
						AddField(testhelper.Namespace, s.Namespace).
						AddField(testhelper.PodDisruptionBudgetReference, pdb.Name))
				} else {
					check.LogInfo("PDB %q is valid for StatefulSet: %q", pdb.Name, s.Name)
					compliantObjects = append(compliantObjects, testhelper.NewReportObject("StatefulSet: references PodDisruptionBudget", testhelper.StatefulSetType, true).
						AddField(testhelper.StatefulSetName, s.Name).
						AddField(testhelper.Namespace, s.Namespace).
						AddField(testhelper.PodDisruptionBudgetReference, pdb.Name))
				}
			}
		}
		if !pdbFound {
			check.LogError("StatefulSet %q is missing a corresponding PodDisruptionBudget", s.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("StatefulSet is missing a corresponding PodDisruptionBudget", testhelper.StatefulSetType, false).
				AddField(testhelper.StatefulSetName, s.Name).
				AddField(testhelper.Namespace, s.Namespace))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// Function to build a map from workload service accounts
// to their associated to-be-deprecated APIs and the release version
// Filters:
// - status.removedInRelease is not empty
// - Verifies if the service account is inside the workload SA list from env.ServiceAccounts
func buildServiceAccountToDeprecatedAPIMap(apiRequestCounts []ocplite.APIRequestCount, workloadServiceAccountNames map[string]struct{}) map[string]map[string]string {
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

// Function to extract unique workload-related service account names from the environment
func extractUniqueServiceAccountNames(env *provider.TestEnvironment) map[string]struct{} {
	uniqueServiceAccountNames := make(map[string]struct{})

	// Iterate over the service accounts to extract names
	for _, sa := range env.ServiceAccounts {
		uniqueServiceAccountNames[sa.Name] = struct{}{}
	}

	return uniqueServiceAccountNames
}

// Function to test API compatibility with the next OCP release
//
//nolint:funlen // The function performs a cohesive flow with dynamic client calls and evaluations.
func testAPICompatibilityWithNextOCPRelease(check *checksdb.Check, env *provider.TestEnvironment) {
	isOCP := provider.IsOCPCluster()
	check.LogInfo("Is OCP: %v", isOCP)

	if !isOCP {
		check.LogInfo("The Kubernetes distribution is not OpenShift. Skipping API compatibility test.")
		return
	}

	// Retrieve APIRequestCount using dynamic client
	oc := clientsholder.GetClientsHolder()
	apirequestGVR := schema.GroupVersionResource{Group: "apiserver.openshift.io", Version: "v1", Resource: "apirequestcounts"}
	ulist, err := oc.DynamicClient.Resource(apirequestGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		check.LogError("Error retrieving APIRequestCount objects: %s", err)
		return
	}

	// Build minimal ocplite objects from unstructured
	var apiRequestCounts []ocplite.APIRequestCount
	for i := range ulist.Items {
		u := &ulist.Items[i]
		item := ocplite.APIRequestCount{Name: u.GetName()}
		if statusObj, ok := u.Object["status"].(map[string]interface{}); ok {
			if rir, ok := statusObj["removedInRelease"].(string); ok {
				item.Status.RemovedInRelease = rir
			}
			if last24hArr, ok := statusObj["last24h"].([]interface{}); ok {
				for _, l := range last24hArr {
					lmap, ok := l.(map[string]interface{})
					if !ok {
						continue
					}
					var perRes ocplite.Last24h
					if byNodeArr, ok := lmap["byNode"].([]interface{}); ok {
						for _, bn := range byNodeArr {
							bnmap, ok := bn.(map[string]interface{})
							if !ok {
								continue
							}
							var perNode ocplite.ByNode
							if byUserArr, ok := bnmap["byUser"].([]interface{}); ok {
								for _, bu := range byUserArr {
									bumap, ok := bu.(map[string]interface{})
									if !ok {
										continue
									}
									var user ocplite.PerUserAPIRequestCount
									if un, ok := bumap["userName"].(string); ok {
										user.UserName = un
									}
									perNode.ByUser = append(perNode.ByUser, user)
								}
							}
							perRes.ByNode = append(perRes.ByNode, perNode)
						}
					}
					item.Status.Last24h = append(item.Status.Last24h, perRes)
				}
			}
		}
		apiRequestCounts = append(apiRequestCounts, item)
	}

	// Extract unique service account names from env.ServiceAccounts
	workloadServiceAccountNames := extractUniqueServiceAccountNames(env)
	check.LogInfo("Detected %d unique service account names for the workload: %v", len(workloadServiceAccountNames), workloadServiceAccountNames)

	// Build a map from service accounts to deprecated APIs
	serviceAccountToDeprecatedAPIs := buildServiceAccountToDeprecatedAPIMap(apiRequestCounts, workloadServiceAccountNames)

	// Evaluate API compliance with the next Kubernetes version
	compliantObjects, nonCompliantObjects := evaluateAPICompliance(serviceAccountToDeprecatedAPIs, env.K8sVersion, workloadServiceAccountNames)

	// Add test results
	check.SetResult(compliantObjects, nonCompliantObjects)
}
