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
	"fmt"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/podhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"

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

func getAllPodsBy(namespace string) (podsInNamespace []*provider.Pod) {
	for i := range env.AllPods {
		pod := env.AllPods[i]
		if pod.Namespace == namespace {
			podsInNamespace = append(podsInNamespace, pod)
		}
	}
	return podsInNamespace
}

func findOperatorsFromPods(namespace string, allPods []*provider.Pod) (foundCsvs map[string]bool, podsBelongingToNoOperators []string, err error) {
	foundCsvs = make(map[string]bool)
	for index := range allPods {
		pod := allPods[index]
		topOwners, err := podhelper.GetPodTopOwner(pod.Namespace, pod.OwnerReferences)
		if err != nil {
			return foundCsvs, podsBelongingToNoOperators, err
		}

		validOwnerFound := false
		for _, owner := range topOwners {
			if owner.Kind == v1alpha1.ClusterServiceVersionKind && owner.Namespace == namespace {
				foundCsvs[owner.Name] = true
				validOwnerFound = true
				break
			}
		}
		if !validOwnerFound {
			podsBelongingToNoOperators = append(podsBelongingToNoOperators, pod.Name)
		}
	}
	return foundCsvs, podsBelongingToNoOperators, nil
}

// This function checks if the namespace contains only valid single namespaced operator without any cluster wide operator and non-operator pods
func containsValidSingleNamespacedOperatorIn(namespace string) (isOperatorOnlyNamespace bool, singleNamespacedCsvs, allNamespacedCsvs, podsBelongingToNoOperators []string, err error) {
	allPods := getAllPodsBy(namespace)

	foundCsvs, podsBelongingToNoOperators, err := findOperatorsFromPods(namespace, allPods)
	if err != nil {
		return false, singleNamespacedCsvs, allNamespacedCsvs, podsBelongingToNoOperators, err
	}

	allOperatorsFoundInNamespaceAreValid := true

	for _, operator := range env.Operators {
		for foundCsv := range foundCsvs {
			if operator.Csv.Name != foundCsv {
				continue
			}

			// Handle cluster-wide operators
			if operator.IsClusterWide {
				allOperatorsFoundInNamespaceAreValid = false
				allNamespacedCsvs = append(allNamespacedCsvs, foundCsv)
				break
			}

			if len(operator.TargetNamespaces) == 1 {
				singleNamespacedCsvs = append(singleNamespacedCsvs, foundCsv)
			}
		}
	}

	if len(podsBelongingToNoOperators) == 0 && allOperatorsFoundInNamespaceAreValid {
		return true, singleNamespacedCsvs, allNamespacedCsvs, podsBelongingToNoOperators, nil
	} else {
		return false, singleNamespacedCsvs, allNamespacedCsvs, podsBelongingToNoOperators, nil
	}
}

func generateNonCompliantMessage(singleNamespacedOperators string, allNamespacedCsvs, podsBelongingToNoOperators []string) (nonCompliantMsg string) {
	prefix := "Operator namespace"
	if singleNamespacedOperators != "" {
		prefix += " with single namespace operators (" + singleNamespacedOperators + ")"
	}

	var allNamespacedOperators string
	if len(allNamespacedCsvs) != 0 { // cluster-wide operator
		allNamespacedOperators = strings.Join(allNamespacedCsvs, ", ")
		nonCompliantMsg = fmt.Sprintf("%s contains all-namespaced operators (%s) ", prefix, allNamespacedOperators)
	}

	if len(podsBelongingToNoOperators) != 0 {
		suffix := fmt.Sprintf("contains some application pods (%s)", strings.Join(podsBelongingToNoOperators, ", "))
		if nonCompliantMsg == "" {
			nonCompliantMsg = fmt.Sprintf("%s %s", prefix, suffix)
		} else {
			nonCompliantMsg += fmt.Sprintf("and %s", suffix)
		}
	}
	return nonCompliantMsg
}
