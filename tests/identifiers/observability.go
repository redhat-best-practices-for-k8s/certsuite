// Copyright (C) 2021-2026 Red Hat, Inc.
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

package identifiers

import (
	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
)

var (
	TestAPICompatibilityWithNextOCPReleaseIdentifier claim.Identifier
	TestCrdsStatusSubresourceIdentifier              claim.Identifier
	TestLoggingIdentifier                            claim.Identifier
	TestPodDisruptionBudgetIdentifier                claim.Identifier
	TestTerminationMessagePolicyIdentifier           claim.Identifier
)

//nolint:funlen
func init() {
	TestAPICompatibilityWithNextOCPReleaseIdentifier = AddCatalogEntry(
		"compatibility-with-next-ocp-release",
		common.ObservabilityTestKey,
		`Checks to ensure if the APIs the workload uses are compatible with the next OCP version`,
		APICompatibilityWithNextOCPReleaseRemediation,
		NoExceptions,
		TestAPICompatibilityWithNextOCPReleaseIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagCommon)
	TestCrdsStatusSubresourceIdentifier = AddCatalogEntry(
		"crd-status",
		common.ObservabilityTestKey,
		`Checks that all CRDs have a status sub-resource specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties["status"]).`,
		CrdsStatusSubresourceRemediation,
		NoExceptions,
		TestCrdsStatusSubresourceIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestLoggingIdentifier = AddCatalogEntry(
		"container-logging",
		common.ObservabilityTestKey,
		`Check that all containers under test use standard input output and standard error when logging. A container must provide APIs for the platform to observe the container health and act accordingly. These APIs include health checks (liveness and readiness), logging to stderr and stdout for log aggregation (by tools such as Logstash or Filebeat), and integrate with tracing and metrics-gathering libraries (such as Prometheus or Metricbeat).`, //nolint:lll
		LoggingRemediation,
		NoDocumentedProcess,
		TestLoggingIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestPodDisruptionBudgetIdentifier = AddCatalogEntry(
		"pod-disruption-budget",
		common.ObservabilityTestKey,
		`Checks to see if pod disruption budgets have allowed values for minAvailable and maxUnavailable, `+
			`and verifies that PDBs are zone-aware (can tolerate an entire zone going offline during platform upgrades)`,
		PodDisruptionBudgetRemediation,
		NoExceptions,
		TestPodDisruptionBudgetIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestTerminationMessagePolicyIdentifier = AddCatalogEntry(
		"termination-policy",
		common.ObservabilityTestKey,
		`Check that all containers are using terminationMessagePolicy: FallbackToLogsOnError. There are different ways a pod can stop on an OpenShift cluster. One way is that the pod can remain alive but non-functional. Another way is that the pod can crash and become non-functional. In the first case, if the administrator has implemented liveness and readiness checks, OpenShift can stop the pod and either restart it on the same node or a different node in the cluster. For the second case, when the application in the pod stops, it should exit with a code and write suitable log entries to help the administrator diagnose what the issue was that caused the problem.`, //nolint:lll
		TerminationMessagePolicyRemediation,
		NoDocumentedProcess,
		TestTerminationMessagePolicyIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)
}
