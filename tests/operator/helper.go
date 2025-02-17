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

// CsvResult holds the results of the splitCsv function.
type CsvResult struct {
	NameCsv   string
	Namespace string
}

// splitCsv splits the input string to extract namecsv and namespace.
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

func getAllPodsBy(namespace string, allPods []*provider.Pod) (podsInNamespace []*provider.Pod) {
	for i := range allPods {
		pod := allPods[i]
		if pod.Namespace == namespace {
			podsInNamespace = append(podsInNamespace, pod)
		}
	}
	return podsInNamespace
}

func getCsvsBy(namespace string, allCsvs []*v1alpha1.ClusterServiceVersion) (csvsInNamespace []*v1alpha1.ClusterServiceVersion) {
	for _, csv := range allCsvs {
		if csv.Namespace == namespace {
			csvsInNamespace = append(csvsInNamespace, csv)
		}
	}
	return csvsInNamespace
}

func isSingleNamespacedOperator(operatorNamespace string, targetNamespaces []string) bool {
	return len(targetNamespaces) == 1 && operatorNamespace != targetNamespaces[0]
}

func isMultiNamespacedOperator(operatorNamespace string, targetNamespaces []string) bool {
	return len(targetNamespaces) > 1 && !stringhelper.StringInSlice(targetNamespaces, operatorNamespace, false)
}

func checkIfCsvUnderTest(csv *v1alpha1.ClusterServiceVersion) bool {
	for _, testOperator := range env.Operators {
		if testOperator.Csv.Name == csv.Name {
			return true
		}
	}
	return false
}

func isCsvInNamespaceClusterWide(csvName string, allCsvs []*v1alpha1.ClusterServiceVersion) bool {
	isClusterWide := true
	for _, eachCsv := range allCsvs {
		if eachCsv.Name == csvName {
			targetNamespaces := eachCsv.Annotations["olm.targetNamespaces"]
			if targetNamespaces != "" {
				isClusterWide = false
				break
			}
		}
	}
	return isClusterWide
}

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
