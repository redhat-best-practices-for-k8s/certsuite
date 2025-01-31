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
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
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
