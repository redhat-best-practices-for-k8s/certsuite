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
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	pdbv1 "github.com/redhat-best-practices-for-k8s/certsuite/tests/observability/pdb"

	apiserv1 "github.com/openshift/api/apiserver/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}
)

// LoadChecks registers the observability checks used in the test suite.
//
// It creates a new checks group, attaches various check functions,
// and configures skip logic for different resource types. The function
// does not take any parameters or return values; it performs all setup
// via side effects on the testing framework.
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

// containerHasLoggingOutput retrieves the last line of a container's log output and reports whether any
// logging was found.
//
// It accepts a pointer to a Container, streams its logs via the Kubernetes client,
// extracts the final line, and returns true if that line contains non-empty content.
// If an error occurs during retrieval or streaming, it returns false along with the error.
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

// testContainersLogging checks that containers produce expected logging output.
//
// It takes a Check and a TestEnvironment, iterates over the containers
// in the environment, verifies each container has logging output,
// logs informational or error messages accordingly, and records the
// results in the report object. The function does not return a value.
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

// testCrds tests whether Custom Resource Definitions have a status subresource set.
//
// It receives a checksdb.Check and a TestEnvironment, performs logging,
// collects test results into report objects, and sets the overall result
// based on whether CRDs expose a status field. The function does not return
// any value; it reports its outcome via side effects on the provided check.
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

// testTerminationMessagePolicy tests that termination message handling for pods behaves as expected.
//
// testTerminationMessagePolicy verifies the termination message logic in observed pods.
// It receives a check object and a test environment, logs progress, builds container
// report objects, updates results, and ensures correct handling of pod termination
// messages. The function does not return a value; it records success or failure
// through the provided check object's result state.
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

// testPodDisruptionBudgets tests the validity of Pod Disruption Budgets in a given environment.
//
// testPodDisruptionBudgets checks that each Pod Disruption Budget in the cluster
// matches expected criteria such as labels, selectors and allowed disruptions.
// It logs information about the test progress, records any errors encountered,
// and updates a report object with the results. The function receives a
// database check context and a test environment, but does not return a value.
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

// buildServiceAccountToDeprecatedAPIMap creates a mapping from workload service accounts to APIs that will be deprecated and the release in which they are removed.
//
// It iterates over a slice of API request counts, selecting entries whose status indicates a future removal (status.removedInRelease is non‑empty). For each such entry it checks whether the service account involved belongs to the list of workload service accounts defined in env.ServiceAccounts. If so, the function records the API name and its removal release under that service account. The result is a map where keys are service account identifiers and values are maps from API names to their corresponding deprecation releases. This map can be used to identify which APIs each service account will need to replace or update before the specified release.
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

// evaluateAPICompliance evaluates the compliance of workload APIs with the next Kubernetes version.
//
// It takes a map of workloads to their API groups, a target minor version string,
// and a set of known non-compliant API groups. The function returns a slice of
// ReportObject pointers describing which workloads are compliant or not.
// For each workload it checks whether any of its API groups are present in the
// non‑compliant set for the next minor version, constructing report objects
// that include fields such as the workload name, API group, and compliance status.
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

// extractUniqueServiceAccountNames retrieves unique service account names related to workloads from the test environment.
//
// It examines the provided TestEnvironment and collects all distinct service account identifiers associated with workload components.
// The result is returned as a map keyed by service account name, where each value is an empty struct for set semantics.
func extractUniqueServiceAccountNames(env *provider.TestEnvironment) map[string]struct{} {
	uniqueServiceAccountNames := make(map[string]struct{})

	// Iterate over the service accounts to extract names
	for _, sa := range env.ServiceAccounts {
		uniqueServiceAccountNames[sa.Name] = struct{}{}
	}

	return uniqueServiceAccountNames
}

// testAPICompatibilityWithNextOCPRelease tests compatibility of cluster APIs with the next OpenShift release.
//
// It receives a checksdb.Check and a TestEnvironment, determines if the cluster is an OCP cluster,
// gathers API request counts, identifies service accounts using deprecated APIs, evaluates
// compliance, and sets the result on the check. The function logs progress and errors during execution.
func testAPICompatibilityWithNextOCPRelease(check *checksdb.Check, env *provider.TestEnvironment) {
	isOCP := provider.IsOCPCluster()
	check.LogInfo("Is OCP: %v", isOCP)

	if !isOCP {
		check.LogInfo("The Kubernetes distribution is not OpenShift. Skipping API compatibility test.")
		return
	}

	// Retrieve APIRequestCount using clientsholder
	oc := clientsholder.GetClientsHolder()
	apiRequestCounts, err := oc.ApiserverClient.ApiserverV1().APIRequestCounts().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		check.LogError("Error retrieving APIRequestCount objects: %s", err)
		return
	}

	// Extract unique service account names from env.ServiceAccounts
	workloadServiceAccountNames := extractUniqueServiceAccountNames(env)
	check.LogInfo("Detected %d unique service account names for the workload: %v", len(workloadServiceAccountNames), workloadServiceAccountNames)

	// Build a map from service accounts to deprecated APIs
	serviceAccountToDeprecatedAPIs := buildServiceAccountToDeprecatedAPIMap(apiRequestCounts.Items, workloadServiceAccountNames)

	// Evaluate API compliance with the next Kubernetes version
	compliantObjects, nonCompliantObjects := evaluateAPICompliance(serviceAccountToDeprecatedAPIs, env.K8sVersion, workloadServiceAccountNames)

	// Add test results
	check.SetResult(compliantObjects, nonCompliantObjects)
}
