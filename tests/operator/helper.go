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

	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
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

func isInstallModeSingleNamespace(installModes []v1alpha1.InstallMode) bool {
	for i := 0; i < len(installModes); i++ {
		if installModes[i].Type == v1alpha1.InstallModeTypeSingleNamespace {
			return true
		}
	}
	return false
}

func findOperatorGroup(name, namespace string, groups []*operatorsv1.OperatorGroup) *operatorsv1.OperatorGroup {
	for _, group := range groups {
		if group.Name == name && group.Namespace == namespace {
			return group
		}
	}
	return nil
}

func checkOperatorInstallationCompliance(opGroupTargetNamespaces []string, operatorNamespace string, targetNamespaces []string, isSingleNamespaceInstallMode bool) bool {
	if isSingleNamespaceInstallMode {
		return len(opGroupTargetNamespaces) == 1 && len(targetNamespaces) == 1 && opGroupTargetNamespaces[0] == targetNamespaces[0]
	}
	return stringhelper.StringInSlice(opGroupTargetNamespaces, operatorNamespace, false) // false in the function arg indicates equals check
}
