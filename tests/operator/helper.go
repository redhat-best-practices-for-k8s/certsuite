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

/*
Package operator provides CNFCERT tests used to validate operator CNF facets.
*/

package operator

import (
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/podhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// CsvResult holds the results of splitting a CSV string.
//
// It contains two fields: NameCsv, which is the extracted name from the input,
// and Namespace, which represents the associated namespace.
// The values are populated by the SplitCsv function when it parses an
// input string into these components.
type CsvResult struct {
	NameCsv   string
	Namespace string
}

// SplitCsv splits a comma‑separated string into name and namespace.
//
// It takes an input string that may contain a single value or a value
// followed by a comma and a namespace. Leading and trailing spaces are
// trimmed, and any leading prefix is removed before the split.
// The function returns a CsvResult struct containing the extracted
// name and namespace components.
func SplitCsv(csv string) CsvResult {
	// Split by comma to separate components
	parts := strings.Split(csv, ",")
	var result CsvResult

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, "ns=") {
			result.Namespace = strings.TrimPrefix(part, "ns=")
		} else {
			result.NameCsv = part
		}
	}
	return result
}

// OperatorInstalledMoreThanOnce reports whether two operator instances are the same.
//
// It compares the names and namespace of two *provider.Operator objects.
// If either name or namespace differs, it returns true indicating that
// an operator appears to be installed more than once in the cluster.
func OperatorInstalledMoreThanOnce(operator1, operator2 *provider.Operator) bool {
	// Safeguard against nil operators (should not happen)
	if operator1 == nil || operator2 == nil {
		return false
	}

	log.Debug("Comparing operator %q with operator %q", operator1.Name, operator2.Name)

	// Retrieve the version from each CSV
	csv1Version := operator1.Csv.Spec.Version.String()
	csv2Version := operator2.Csv.Spec.Version.String()

	log.Debug("CSV1 Version: %s", csv1Version)
	log.Debug("CSV2 Version: %s", csv2Version)

	// Strip the version from the CSV name by removing the suffix (which should be the version)
	csv1Name := strings.TrimSuffix(operator1.Csv.Name, ".v"+csv1Version)
	csv2Name := strings.TrimSuffix(operator2.Csv.Name, ".v"+csv2Version)

	log.Debug("Comparing CSV names %q and %q", csv1Name, csv2Name)

	// The CSV name should be the same, but the version should be different
	// if the operator is installed more than once.
	if operator1.Csv != nil && operator2.Csv != nil &&
		csv1Name == csv2Name &&
		csv1Version != csv2Version {
		log.Error("Operator %q is installed more than once", operator1.Name)
		return true
	}

	return false
}

// getAllPodsBy returns a filtered slice of pods matching the given selector.
//
// It iterates over the provided list of pods, appending those whose
// annotations or labels satisfy the selector string to a new slice,
// which is then returned.
func getAllPodsBy(namespace string, allPods []*provider.Pod) (podsInNamespace []*provider.Pod) {
	for i := range allPods {
		pod := allPods[i]
		if pod.Namespace == namespace {
			podsInNamespace = append(podsInNamespace, pod)
		}
	}
	return podsInNamespace
}

// getCsvsBy returns a filtered slice of ClusterServiceVersions that match the given key.
// It iterates over the provided list of CSV objects, selecting those whose name matches the specified key,
// and appends them to a new slice which is then returned. The function does not modify the input slice.
func getCsvsBy(namespace string, allCsvs []*v1alpha1.ClusterServiceVersion) (csvsInNamespace []*v1alpha1.ClusterServiceVersion) {
	for _, csv := range allCsvs {
		if csv.Namespace == namespace {
			csvsInNamespace = append(csvsInNamespace, csv)
		}
	}
	return csvsInNamespace
}

// isSingleNamespacedOperator determines whether the operator should operate in a single‑namespace mode.
//
// It examines the supplied operator name and list of target namespaces.
// If exactly one namespace is specified (or no namespaces but a default
// single‑namespace flag is set), it returns true, indicating that the
// operator will run against a single namespace. Otherwise it returns false.
func isSingleNamespacedOperator(operatorNamespace string, targetNamespaces []string) bool {
	return len(targetNamespaces) == 1 && operatorNamespace != targetNamespaces[0]
}

// isMultiNamespacedOperator reports whether the operator can run in multiple namespaces.
//
// It receives a namespace string and a slice of allowed namespaces.
// If the slice contains more than one entry, it returns true only when
// the given namespace is present in that list; otherwise it returns false.
func isMultiNamespacedOperator(operatorNamespace string, targetNamespaces []string) bool {
	return len(targetNamespaces) > 1 && !stringhelper.StringInSlice(targetNamespaces, operatorNamespace, false)
}

// checkIfCsvUnderTest reports whether the given ClusterServiceVersion is in the test scope.
//
// It examines the metadata of the supplied ClusterServiceVersion to determine if it matches
// the criteria used by the test suite for identifying CSVs under test.
// The function returns true when the CSV should be considered part of the current test run,
// and false otherwise.
func checkIfCsvUnderTest(csv *v1alpha1.ClusterServiceVersion) bool {
	for _, testOperator := range env.Operators {
		if testOperator.Csv.Name == csv.Name {
			return true
		}
	}
	return false
}

// isCsvInNamespaceClusterWide checks whether a CSV is cluster-wide within a namespace.
// It takes a namespace string and a slice of ClusterServiceVersion pointers, then
// returns true if any CSV in the list has a target cluster scope set for that namespace,
// indicating it operates across the entire cluster rather than being confined to a single namespace.
func isCsvInNamespaceClusterWide(csvName string, allCsvs []*v1alpha1.ClusterServiceVersion) bool {
	isClusterWide := true
	for _, eachCsv := range allCsvs {
		if eachCsv.Name == csvName {
			targetNamespaces, exists := eachCsv.Annotations["olm.targetNamespaces"]
			if exists && targetNamespaces != "" {
				isClusterWide = false
				break
			}
		}
	}
	return isClusterWide
}

// checkValidOperatorInstallation validates that the operator installation matches expectations.
//
// It receives a string describing the operator and returns three values: a boolean indicating
// whether the installation is valid, a slice of strings containing any error details,
// and an error if one occurred during processing. The function checks CSVs, namespaces,
// and pod ownership to ensure the operator behaves as expected in the test environment.
func checkValidOperatorInstallation(namespace string) (isDedicatedOperatorNamespace bool, singleOrMultiNamespaceOperators,
	nonSingleOrMultiNamespaceOperators, csvsTargetingNamespace, operatorsFoundButNotUnderTest, podsNotBelongingToOperators []string, err error) {
	// 1. operator installation checks
	csvsInNamespace := getCsvsBy(namespace, env.AllCsvs)

	for _, csv := range csvsInNamespace {
		operatorNamespace := csv.Annotations["olm.operatorNamespace"]
		targetNamespacesStr := csv.Annotations["olm.targetNamespaces"]

		var targetNameSpaces []string
		if targetNamespacesStr != "" {
			targetNameSpaces = strings.Split(targetNamespacesStr, ",")
		}

		if namespace == operatorNamespace {
			if checkIfCsvUnderTest(csv) {
				isSingleOrMultiInstallation := isSingleNamespacedOperator(operatorNamespace, targetNameSpaces) || isMultiNamespacedOperator(operatorNamespace, targetNameSpaces)
				if isSingleOrMultiInstallation {
					singleOrMultiNamespaceOperators = append(singleOrMultiNamespaceOperators, csv.Name)
				} else {
					nonSingleOrMultiNamespaceOperators = append(nonSingleOrMultiNamespaceOperators, csv.Name)
				}
			} else {
				operatorsFoundButNotUnderTest = append(operatorsFoundButNotUnderTest, csv.Name)
			}
		} else {
			if !isCsvInNamespaceClusterWide(csv.Name, env.AllCsvs) { // check for non-cluster wide operators
				csvsTargetingNamespace = append(csvsTargetingNamespace, csv.Name)
			}
		}
	}

	// 2. non-operator pods check
	podsNotBelongingToOperators, err = findPodsNotBelongingToOperators(namespace)
	if err != nil {
		return false, singleOrMultiNamespaceOperators, nonSingleOrMultiNamespaceOperators, csvsTargetingNamespace, operatorsFoundButNotUnderTest, podsNotBelongingToOperators, err
	}

	var isValid bool
	if len(singleOrMultiNamespaceOperators) > 0 {
		isValid = len(nonSingleOrMultiNamespaceOperators) == 0 && len(csvsTargetingNamespace) == 0 && len(podsNotBelongingToOperators) == 0 && len(operatorsFoundButNotUnderTest) == 0
	}

	return isValid, singleOrMultiNamespaceOperators, nonSingleOrMultiNamespaceOperators, csvsTargetingNamespace, operatorsFoundButNotUnderTest, podsNotBelongingToOperators, nil
}

// findPodsNotBelongingToOperators identifies pods that are not managed by any operator within a given namespace.
//
// It takes the name of a Kubernetes namespace as input, retrieves all pods in that namespace,
// checks each pod's top-level owner reference to determine if it belongs to an operator,
// and collects the names of those pods that do not have an operator as their owner.
// The function returns a slice of these pod names and an error if any step fails.
func findPodsNotBelongingToOperators(namespace string) (podsBelongingToNoOperators []string, err error) {
	allPods := getAllPodsBy(namespace, env.AllPods)
	for index := range allPods {
		pod := allPods[index]
		topOwners, err := podhelper.GetPodTopOwner(pod.Namespace, pod.OwnerReferences)
		if err != nil {
			return podsBelongingToNoOperators, err
		}

		validOwnerFound := false
		for _, owner := range topOwners {
			if owner.Kind == v1alpha1.ClusterServiceVersionKind && owner.Namespace == namespace {
				validOwnerFound = true
				break
			}
		}
		if !validOwnerFound {
			podsBelongingToNoOperators = append(podsBelongingToNoOperators, pod.Name)
		}
	}

	return podsBelongingToNoOperators, nil
}
