// Copyright (C) 2020-2026 Red Hat, Inc.
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

package checksadapter

import (
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/checks"
)

// ConvertAndSetResult converts checks.CheckResult to certsuite format and sets it on the check.
//
// The certsuite's SetResult treats empty compliant+nonCompliant lists as "skipped", so when
// the checks library returns Compliant with no resource details we synthesize a single
// compliant object to ensure the check records as "passed".
func ConvertAndSetResult(check *checksdb.Check, result checks.CheckResult) {
	if result.ComplianceStatus == checks.StatusSkipped {
		reason := result.Reason
		if reason == "" {
			reason = "check skipped"
		}
		check.SetResultSkipped(reason)
		return
	}

	compliantObjects := []*testhelper.ReportObject{}
	nonCompliantObjects := []*testhelper.ReportObject{}

	for _, detail := range result.Details {
		// Filter out non-compliant results caused by pods/containers that were
		// recreated between autodiscovery and check execution. Scoped to
		// Pod/Container kinds to avoid suppressing legitimate "not found" failures
		// from other resource types (e.g., certification database lookups).
		if !detail.Compliant && (detail.Kind == testhelper.PodType || detail.Kind == testhelper.ContainerType) &&
			strings.Contains(detail.Message, "not found") {
			continue
		}
		reportObj := convertDetailToReportObject(detail)
		if detail.Compliant {
			compliantObjects = append(compliantObjects, reportObj)
		} else {
			nonCompliantObjects = append(nonCompliantObjects, reportObj)
		}
	}

	// If NonCompliant or Error with no details, synthesize a non-compliant object.
	if (result.ComplianceStatus == checks.StatusNonCompliant || result.ComplianceStatus == checks.StatusError) && len(nonCompliantObjects) == 0 {
		nonCompliantObjects = append(nonCompliantObjects,
			testhelper.NewReportObject(result.Reason, testhelper.UndefinedType, false))
	}

	// Certsuite's SetResult treats empty compliant+nonCompliant lists as SKIPPED.
	// When the checks library validated resources, it returns compliant details --
	// those flow through as compliant objects above, producing PASS.
	// When nothing was checked (no details at all), let empty lists produce SKIP.
	// Only synthesize if details existed but were all filtered out by the "not found" filter.
	if len(compliantObjects) == 0 && len(nonCompliantObjects) == 0 && len(result.Details) > 0 {
		compliantObjects = append(compliantObjects,
			testhelper.NewReportObject("No violations found", testhelper.UndefinedType, true))
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func convertDetailToReportObject(detail checks.ResourceDetail) *testhelper.ReportObject {
	var objType string
	var nameField = testhelper.Name
	switch detail.Kind {
	case "Pod":
		objType = testhelper.PodType
		nameField = testhelper.PodName
	case "Container":
		objType = testhelper.ContainerType
		nameField = testhelper.ContainerName
	case "Deployment":
		objType = testhelper.DeploymentType
		nameField = testhelper.DeploymentName
	case "StatefulSet":
		objType = testhelper.StatefulSetType
		nameField = testhelper.StatefulSetName
	case "Service":
		objType = testhelper.ServiceType
		nameField = testhelper.ServiceName
	case "RoleBinding", "ClusterRoleBinding":
		objType = testhelper.RoleType
		nameField = testhelper.RoleBindingName
	case "CustomResourceDefinition":
		objType = testhelper.CustomResourceDefinitionType
		nameField = testhelper.CustomResourceDefinitionName
	case "ClusterServiceVersion":
		objType = testhelper.OperatorType
	case "Namespace":
		objType = testhelper.Namespace
	case "Node":
		objType = testhelper.NodeType
	case "CatalogSource":
		objType = testhelper.CatalogSourceType
	case "HelmRelease":
		objType = testhelper.HelmVersionType
	default:
		objType = testhelper.UndefinedType
	}

	reportObj := testhelper.NewReportObject(detail.Message, objType, detail.Compliant)

	if detail.Namespace != "" {
		reportObj.AddField(testhelper.Namespace, detail.Namespace)
	}
	if detail.Name != "" {
		reportObj.AddField(nameField, detail.Name)
	}

	return reportObj
}
