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
	TestCPUPinningNoExecProbes           claim.Identifier
	TestExclusiveCPUPoolIdentifier       claim.Identifier
	TestExclusiveCPUPoolSchedulingPolicy claim.Identifier
	TestIsolatedCPUPoolSchedulingPolicy  claim.Identifier
	TestLimitedUseOfExecProbesIdentifier claim.Identifier
	TestRtAppNoExecProbes                claim.Identifier
	TestSharedCPUPoolSchedulingPolicy    claim.Identifier
)

//nolint:funlen
func init() {
	TestCPUPinningNoExecProbes = AddCatalogEntry(
		"cpu-pinning-no-exec-probes",
		common.PerformanceTestKey,
		`Workloads utilizing CPU pinning (Guaranteed QoS with exclusive CPUs) should not use exec probes. Exec probes run a command within the container, which could interfere with latency-sensitive workloads and cause performance degradation.`,
		CPUPinningNoExecProbesRemediation,
		NoDocumentedProcess,
		TestCPUPinningNoExecProbesDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestExclusiveCPUPoolIdentifier = AddCatalogEntry(
		"exclusive-cpu-pool",
		common.PerformanceTestKey,
		`Ensures that if one container in a Pod selects an exclusive CPU pool the rest select the same type of CPU pool`,
		ExclusiveCPUPoolRemediation,
		NoDocumentedProcess,
		TestExclusiveCPUPoolIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagFarEdge)

	TestExclusiveCPUPoolSchedulingPolicy = AddCatalogEntry(
		"exclusive-cpu-pool-rt-scheduling-policy",
		common.PerformanceTestKey,
		`Ensures that if application workload runs in exclusive CPU pool, it chooses RT CPU schedule policy and set the priority less than 10.`,
		ExclusiveCPUPoolSchedulingPolicyRemediation,
		NoDocumentedProcess,
		TestExclusiveCPUPoolSchedulingPolicyDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagFarEdge)

	TestIsolatedCPUPoolSchedulingPolicy = AddCatalogEntry(
		"isolated-cpu-pool-rt-scheduling-policy",
		common.PerformanceTestKey,
		`Ensures that a workload running in an application-isolated exclusive CPU pool selects a RT CPU scheduling policy`,
		IsolatedCPUPoolSchedulingPolicyRemediation,
		NoDocumentedProcess,
		TestIsolatedCPUPoolSchedulingPolicyDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagFarEdge)

	TestLimitedUseOfExecProbesIdentifier = AddCatalogEntry(
		"max-resources-exec-probes",
		common.PerformanceTestKey,
		`Checks that less than 10 exec probes are configured in the cluster for this workload. Also checks that the periodSeconds parameter for each probe is superior or equal to 10.`,
		LimitedUseOfExecProbesRemediation,
		NoDocumentedProcess,
		TestLimitedUseOfExecProbesIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional},
		TagFarEdge)
	TestRtAppNoExecProbes = AddCatalogEntry(
		"rt-apps-no-exec-probes",
		common.PerformanceTestKey,
		`Ensures that if one container runs a real time application exec probes are not used`,
		RtAppNoExecProbesRemediation,
		NoDocumentedProcess,
		TestRtAppNoExecProbesDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagFarEdge)

	TestSharedCPUPoolSchedulingPolicy = AddCatalogEntry(
		"shared-cpu-pool-non-rt-scheduling-policy",
		common.PerformanceTestKey,
		`Ensures that if application workload runs in shared CPU pool, it chooses non-RT CPU schedule policy to always share the CPU with other applications and kernel threads.`,
		SharedCPUPoolSchedulingPolicyRemediation,
		NoDocumentedProcess,
		TestSharedCPUPoolSchedulingPolicyDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagFarEdge)
}
