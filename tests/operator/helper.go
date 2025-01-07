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
	"context"
	"fmt"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/podhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func hasOperatorInstallModeSingleNamespace(installModes []v1alpha1.InstallMode) bool {
	for i := 0; i < len(installModes); i++ {
		if installModes[i].Type == v1alpha1.InstallModeTypeSingleNamespace && installModes[i].Supported {
			return true
		}
	}
	return false
}

func filterSingleNamespacedOperatorUnderTest(operators []*provider.Operator) (singleNamespacedOperators []*provider.Operator) {
	for _, operator := range operators {
		if hasOperatorInstallModeSingleNamespace(operator.Csv.Spec.InstallModes) && len(operator.TargetNamespaces) == 1 {
			singleNamespacedOperators = append(singleNamespacedOperators, operator)
		}
	}
	return singleNamespacedOperators
}

// This function checks if the namespace contains only valid operator pods
func checkIfNamespaceContainsOnlyOperatorPods(namespace string) (isOperatorOnlyNamespace bool, err error) {
	// Get all pods from the target namespace

	podsList, err := clientsholder.GetClientsHolder().K8sClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return isOperatorOnlyNamespace, err
	}
	foundCsvs := make(map[string]bool)
	foundOperatorPods := 0
	for index := range podsList.Items {
		// Get the top owners of the pod
		pod := podsList.Items[index]
		topOwners, err := podhelper.GetPodTopOwner(pod.Namespace, pod.OwnerReferences)
		if err != nil {
			return isOperatorOnlyNamespace, fmt.Errorf("could not get top owners of Pod %s (in namespace %s), err=%v", pod.Name, pod.Namespace, err)
		}

		// check if owner matches with the csv
		for _, owner := range topOwners {
			// The owner must be in the targetNamespace
			if owner.Kind == v1alpha1.ClusterServiceVersionKind && owner.Namespace == namespace {
				foundOperatorPods++
				foundCsvs[owner.Name] = true
				break
			}
		}
	}

	// Check if the found CSVs contain only valid operators under test
	allOperatorsFoundInNamespaceAreValid := true
	for _, operator := range env.Operators {
		for foundCsv := range foundCsvs {
			// Report an error only if an operator under test is found to be clusterwide
			if operator.IsClusterWide && operator.Csv.Name == foundCsv {
				allOperatorsFoundInNamespaceAreValid = false
				break
			}
		}
	}

	return len(podsList.Items) == foundOperatorPods && allOperatorsFoundInNamespaceAreValid, nil
}
