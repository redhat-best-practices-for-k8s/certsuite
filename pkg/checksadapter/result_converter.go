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
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/checks"
)

// ConvertAndSetResult converts checks.CheckResult to certsuite format and sets it on the check.
func ConvertAndSetResult(check *checksdb.Check, result checks.CheckResult) {
	if result.ComplianceStatus == "Skipped" {
		check.SetResultSkipped(result.Reason)
		return
	}

	compliantObjects := []*testhelper.ReportObject{}
	nonCompliantObjects := []*testhelper.ReportObject{}

	for _, detail := range result.Details {
		reportObj := convertDetailToReportObject(detail)
		if detail.Compliant {
			compliantObjects = append(compliantObjects, reportObj)
		} else {
			nonCompliantObjects = append(nonCompliantObjects, reportObj)
		}
	}

	// If no details but status is NonCompliant, create a generic report
	if len(result.Details) == 0 && result.ComplianceStatus == "NonCompliant" {
		nonCompliantObjects = append(nonCompliantObjects,
			testhelper.NewReportObject(result.Reason, testhelper.UndefinedType, false))
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func convertDetailToReportObject(detail checks.ResourceDetail) *testhelper.ReportObject {
	var objType string
	switch detail.Kind {
	case "Pod":
		objType = testhelper.PodType
	case "Container":
		objType = testhelper.ContainerType
	case "Deployment":
		objType = testhelper.DeploymentType
	case "StatefulSet":
		objType = testhelper.StatefulSetType
	case "Service":
		objType = testhelper.ServiceType
	case "RoleBinding", "ClusterRoleBinding":
		objType = testhelper.RoleType
	case "CustomResourceDefinition":
		objType = testhelper.CustomResourceDefinitionType
	case "ClusterServiceVersion":
		objType = testhelper.OperatorType // Use OperatorType for CSVs
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
		reportObj.AddField(testhelper.Name, detail.Name)
	}

	return reportObj
}
