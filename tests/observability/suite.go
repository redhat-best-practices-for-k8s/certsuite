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

// LoadChecks Initializes the observability test suite
//
// The function creates a new checks group for observability and registers
// several checks related to logging, CRD status subresources, termination
// message policy, pod disruption budgets, and API compatibility with future
// OpenShift releases. Each check is configured with optional skip functions
// that determine whether the environment contains relevant objects before
// execution. Debug output records the loading of this suite.
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

// containerHasLoggingOutput Checks whether a container has produced any log output
//
// The function retrieves the last two lines of a pod’s logs via the
// Kubernetes API, reads them into memory, and returns true if any content was
// found. It handles errors from establishing the stream or copying data,
// returning false with an error in those cases. The result indicates whether
// the container produced at least one line to stdout or stderr.
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

// testContainersLogging Verifies that containers emit log output to stdout or stderr
//
// The function iterates over all containers under test, attempts to fetch their
// most recent log lines, and records whether any logs were present. Containers
// lacking logs or encountering errors are marked non‑compliant, while those
// producing at least one line are marked compliant. The results are aggregated
// into report objects for later analysis.
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

// testCrds Verifies CRD status subresource presence
//
// The function iterates over all custom resource definitions in the test
// environment, checking each version for a "status" property in its schema. For
// every missing status field it logs an error and records a non‑compliant
// report object; otherwise it logs success and records a compliant report.
// Finally, it sets the check result with lists of compliant and non‑compliant
// objects.
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

// testTerminationMessagePolicy Verifies container termination message policies
//
// The function iterates over each container in the test environment, checking
// whether its TerminationMessagePolicy is set to FallbackToLogsOnError.
// Containers that meet this requirement are recorded as compliant; others are
// marked non-compliant with an explanatory report object. After processing all
// containers, the check results are stored for reporting.
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

// testPodDisruptionBudgets Verifies that deployments and stateful sets have valid pod disruption budgets
//
// The function iterates through all deployments and stateful sets in the test
// environment, checking for a matching PodDisruptionBudget by label selector.
// It validates each found PDB against the replica count of its controller using
// an external checker. Results are recorded as compliant or non‑compliant
// report objects, which are then set on the check result.
//
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

// buildServiceAccountToDeprecatedAPIMap Creates a mapping of service accounts to APIs slated for removal
//
// The function receives a slice of API request count objects and a set of
// workload service account names. It iterates through the usage data,
// extracting each service account that appears in the workload list and
// recording any API whose removal release is specified. The result is a nested
// map where each key is a service account name and its value maps deprecated
// APIs to their corresponding Kubernetes release version.
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

// evaluateAPICompliance Assesses whether service accounts use APIs that will be removed in the next Kubernetes release
//
// The function parses the current Kubernetes version, increments it to
// determine the upcoming release, and then checks each deprecated API used by a
// service account against the removal schedule. It creates report objects
// indicating compliance or non‑compliance for each API, adding relevant
// fields such as the API name, service account, and removal or active release.
// If no APIs are detected, it generates pass reports for all workload service
// accounts.
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

// extractUniqueServiceAccountNames collects distinct service account names from the test environment
//
// It receives a test environment, iterates over its ServiceAccounts slice, and
// inserts each name into a map to ensure uniqueness. The resulting map has keys
// of type string and empty struct values, providing an efficient set
// representation for later use in compatibility checks.
func extractUniqueServiceAccountNames(env *provider.TestEnvironment) map[string]struct{} {
	uniqueServiceAccountNames := make(map[string]struct{})

	// Iterate over the service accounts to extract names
	for _, sa := range env.ServiceAccounts {
		uniqueServiceAccountNames[sa.Name] = struct{}{}
	}

	return uniqueServiceAccountNames
}

// testAPICompatibilityWithNextOCPRelease Checks whether workload APIs remain available in the upcoming OpenShift release
//
// The function first verifies that the cluster is an OpenShift distribution,
// then gathers API request usage data via the ApiserverV1 client. It maps each
// service account to any deprecated APIs it has used and compares these
// deprecation releases against the next minor Kubernetes version. Results are
// recorded as compliant or non‑compliant objects for reporting.
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
