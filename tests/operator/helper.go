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

// CsvResult contains the parsed CSV components
//
// This structure holds the two parts produced by splitting a comma-separated
// string: one part is stored as NameCsv and the other, if prefixed with "ns=",
// is stored as Namespace. It is used to return values from the SplitCsv
// function.
type CsvResult struct {
	NameCsv   string
	Namespace string
}

// SplitCsv Separates a CSV string into its name and namespace components
//
// This function takes a comma‑delimited string, splits it into parts, trims
// whitespace, and assigns the portion prefixed with "ns=" to the Namespace
// field while the remaining part becomes NameCsv. It returns a CsvResult struct
// containing these two fields. If no namespace prefix is present, Namespace
// remains empty.
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

// OperatorInstalledMoreThanOnce Detects if the same operator appears more than once
//
// The function compares two operator instances by examining their CSV names and
// versions. It first removes the version suffix from each CSV name, then checks
// that the base names match while the versions differ. If both conditions hold,
// it reports that the operator is installed multiple times; otherwise it
// returns false.
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

// getAllPodsBy Filters pods by namespace
//
// The function iterates over a slice of pod objects, selecting only those whose
// Namespace field matches the provided namespace string. Matching pods are
// appended to a new slice that is returned to the caller. This helper
// simplifies gathering all pods within a specific namespace for further
// processing.
func getAllPodsBy(namespace string, allPods []*provider.Pod) (podsInNamespace []*provider.Pod) {
	for i := range allPods {
		pod := allPods[i]
		if pod.Namespace == namespace {
			podsInNamespace = append(podsInNamespace, pod)
		}
	}
	return podsInNamespace
}

// getCsvsBy Filters CSVs to a specific namespace
//
// This function iterates over all provided ClusterServiceVersion objects,
// selecting only those whose Namespace field matches the supplied string. The
// matching CSVs are collected into a slice that is returned to the caller. If
// no CSVs match, an empty slice is returned.
func getCsvsBy(namespace string, allCsvs []*v1alpha1.ClusterServiceVersion) (csvsInNamespace []*v1alpha1.ClusterServiceVersion) {
	for _, csv := range allCsvs {
		if csv.Namespace == namespace {
			csvsInNamespace = append(csvsInNamespace, csv)
		}
	}
	return csvsInNamespace
}

// isSingleNamespacedOperator Determines if an operator is single‑namespace scoped but targets a different namespace
//
// The function checks that the targetNamespaces slice contains exactly one
// entry and that this entry differs from the operatorNamespace. If both
// conditions hold, it returns true indicating the operator runs in its own
// namespace yet serves another namespace; otherwise it returns false.
func isSingleNamespacedOperator(operatorNamespace string, targetNamespaces []string) bool {
	return len(targetNamespaces) == 1 && operatorNamespace != targetNamespaces[0]
}

// isMultiNamespacedOperator determines if an operator targets multiple namespaces excluding its own
//
// This function checks whether the list of target namespaces for an operator
// contains more than one entry and that the operator’s own namespace is not
// among them. It returns true only when the operator is intended to operate
// across several distinct namespaces, indicating a multi‑namespaced
// deployment scenario.
func isMultiNamespacedOperator(operatorNamespace string, targetNamespaces []string) bool {
	return len(targetNamespaces) > 1 && !stringhelper.StringInSlice(targetNamespaces, operatorNamespace, false)
}

// checkIfCsvUnderTest determines if a CSV is part of the test set
//
// The function iterates through the global list of operators defined for
// testing, checking whether any entry’s CSV name matches that of the supplied
// object. If a match is found it returns true; otherwise false. This boolean
// indicates whether the given CSV should be considered under test in subsequent
// validation logic.
func checkIfCsvUnderTest(csv *v1alpha1.ClusterServiceVersion) bool {
	for _, testOperator := range env.Operators {
		if testOperator.Csv.Name == csv.Name {
			return true
		}
	}
	return false
}

// isCsvInNamespaceClusterWide determines if a CSV is cluster‑wide based on its annotations
//
// The function scans all provided ClusterServiceVersions for the one matching
// the given name. It checks whether that CSV has a nonempty
// "olm.targetNamespaces" annotation; if so, it marks the CSV as not
// cluster‑wide. The result is returned as a boolean indicating whether the
// operator applies across the entire cluster.
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

// checkValidOperatorInstallation Determines if a namespace hosts only valid single or multi‑namespace operators
//
// The function inspects all ClusterServiceVersions in the specified namespace,
// categorising them as installed under test, not under test, or targeting other
// namespaces. It also checks for non‑operator pods that do not belong to any
// operator. The return values indicate whether the namespace is dedicated to
// valid operators and provide lists of any problematic objects.
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

// findPodsNotBelongingToOperators identifies pods that are not managed by any operator in the given namespace
//
// The function retrieves all pods within a namespace, then for each pod
// determines its top-level owners using helper logic. It checks whether any
// owner is a ClusterServiceVersion belonging to the same namespace; if none
// exist, the pod name is added to the result list. The returned slice contains
// names of pods that are not controlled by an operator, along with an error if
// ownership resolution fails.
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
