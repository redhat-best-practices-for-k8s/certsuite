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

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/podhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
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

func getAllPodsBy(namespace string, allPods []*provider.Pod) (podsInNamespace []*provider.Pod) {
	for i := range allPods {
		pod := allPods[i]
		if pod.Namespace == namespace {
			podsInNamespace = append(podsInNamespace, pod)
		}
	}
	return podsInNamespace
}

func getAllOperatorsBy(namespace string, operators []*provider.Operator) (operatorsInNamespace []*provider.Operator) {
	for _, operator := range operators {
		if operator.Csv.Namespace == namespace {
			operatorsInNamespace = append(operatorsInNamespace, operator)
		}
	}
	return operatorsInNamespace
}

func isSingleNamespacedOperator(operator *provider.Operator) bool {
	return len(operator.TargetNamespaces) == 1 && operator.Namespace != operator.TargetNamespaces[0]
}

func isMultiNamespacedOperator(operator *provider.Operator) bool {
	return len(operator.TargetNamespaces) > 1 && !stringhelper.StringInSlice(operator.TargetNamespaces, operator.Namespace, false)
}

func checkIfOperatorUnderTest(operator *provider.Operator) bool {
	for _, testOperator := range env.Operators {
		if testOperator.Name == operator.Name && testOperator.Namespace == operator.Namespace {
			return true
		}
	}

	return false
}

func checkValidOperatorInstallation(namespace string) (isDedicatedOperatorNamespace bool, singleOrMultiNamespaceOperators, nonSingleOrMultiNamespaceOperators, csvsFoundButNotInOperatorInstallationNamespace, operatorsFoundButNotUnderTest, podsNotBelongingToOperators []string) {
	// 1. operator installation checks
	for _, operator := range getAllOperatorsBy(namespace, env.AllOperators) {
		if namespace == operator.Csv.Annotations["olm.operatorNamespace"] {
			if checkIfOperatorUnderTest(operator) {
				if isSingleNamespacedOperator(operator) || isMultiNamespacedOperator(operator) {
					singleOrMultiNamespaceOperators = append(singleOrMultiNamespaceOperators, operator.Name)
				} else {
					nonSingleOrMultiNamespaceOperators = append(nonSingleOrMultiNamespaceOperators, operator.Name)
				}
			} else {
				operatorsFoundButNotUnderTest = append(operatorsFoundButNotUnderTest, operator.Name)
			}
		} else {
			csvsFoundButNotInOperatorInstallationNamespace = append(csvsFoundButNotInOperatorInstallationNamespace, operator.Name)
		}
	}
	// 2. existing pods check
	podsBelongingToNoOperators, err := findPodsNotBelongingToOperators(namespace)
	if err != nil {
		return false, singleOrMultiNamespaceOperators, nonSingleOrMultiNamespaceOperators, csvsFoundButNotInOperatorInstallationNamespace, operatorsFoundButNotUnderTest, podsNotBelongingToOperators
	}

	return len(podsBelongingToNoOperators) == 0 && len(singleOrMultiNamespaceOperators) != 0, singleOrMultiNamespaceOperators, nonSingleOrMultiNamespaceOperators, csvsFoundButNotInOperatorInstallationNamespace, operatorsFoundButNotUnderTest, podsNotBelongingToOperators

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
