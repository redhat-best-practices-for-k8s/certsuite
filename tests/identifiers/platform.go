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
	TestClusterOperatorHealth                claim.Identifier
	TestHugepagesNotManuallyManipulated      claim.Identifier
	TestHyperThreadEnable                    claim.Identifier
	TestIsRedHatReleaseIdentifier            claim.Identifier
	TestIsSELinuxEnforcingIdentifier         claim.Identifier
	TestNodeOperatingSystemIdentifier        claim.Identifier
	TestNonTaintedNodeKernelsIdentifier      claim.Identifier
	TestOCPLifecycleIdentifier               claim.Identifier
	TestPodHugePages1G                       claim.Identifier
	TestPodHugePages2M                       claim.Identifier
	TestServiceMeshIdentifier                claim.Identifier
	TestSysctlConfigsIdentifier              claim.Identifier
	TestUnalteredBaseImageIdentifier         claim.Identifier
	TestUnalteredStartupBootParamsIdentifier claim.Identifier
)

//nolint:funlen
func init() {
	TestClusterOperatorHealth = AddCatalogEntry(
		"cluster-operator-health",
		common.PlatformAlterationTestKey,
		`Tests that all cluster operators are healthy.`,
		ClusterOperatorHealthRemediation,
		NoExceptions,
		TestClusterOperatorHealthDocLink,
		false,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagCommon)

	TestHugepagesNotManuallyManipulated = AddCatalogEntry(
		"hugepages-config",
		common.PlatformAlterationTestKey,
		`Checks to see that HugePage settings have been configured through MachineConfig, and not manually on the underlying Node. This test case applies only to Nodes that are labeled as workers with the standard label "node-role.kubernetes.io/worker". First, the MachineConfig is inspected for hugepage settings in systemd units. If not, the MC's .spec.kernelArguments are inspected for hugepage settings. The sizes and page numbers are compared, and the test passes only if they are the same than then ones in node's /sys/kernel/mm/hugepages/hugepages-X folders.`, //nolint:lll
		HugepagesNotManuallyManipulatedRemediation,
		NoExceptions,
		TestHugepagesNotManuallyManipulatedDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestHyperThreadEnable = AddCatalogEntry(
		"hyperthread-enable",
		common.PlatformAlterationTestKey,
		`Check that baremetal workers have hyperthreading enabled`,
		HyperThreadEnable,
		NoDocumentedProcess,
		TestHyperThreadEnableDocLink,
		false,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagExtended)

	TestIsRedHatReleaseIdentifier = AddCatalogEntry(
		"isredhat-release",
		common.PlatformAlterationTestKey,
		`verifies if the container base image is redhat.`,
		IsRedHatReleaseRemediation,
		NoExceptions,
		TestIsRedHatReleaseIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestIsSELinuxEnforcingIdentifier = AddCatalogEntry(
		"is-selinux-enforcing",
		common.PlatformAlterationTestKey,
		`verifies that all openshift platform/cluster nodes have selinux in "Enforcing" mode.`,
		IsSELinuxEnforcingRemediation,
		NoExceptions,
		TestIsSELinuxEnforcingIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)
	TestNodeOperatingSystemIdentifier = AddCatalogEntry(
		"ocp-node-os-lifecycle",
		common.PlatformAlterationTestKey,
		`Tests that the nodes running in the cluster have operating systems that are compatible with the deployed version of OpenShift.`,
		NodeOperatingSystemRemediation,
		NoExceptions,
		TestNodeOperatingSystemIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestNonTaintedNodeKernelsIdentifier = AddCatalogEntry(
		"tainted-node-kernel",
		common.PlatformAlterationTestKey,
		`Ensures that the Node(s) hosting workloads do not utilize tainted kernels. This test case is especially
important to support Highly Available workloads, since when a workload is re-instantiated on a backup Node,
that Node's kernel may not have the same hacks.'`,
		NonTaintedNodeKernelsRemediation,
		`If taint is necessary, document details of the taint and why it's needed by workload or environment.`,
		TestNonTaintedNodeKernelsIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestOCPLifecycleIdentifier = AddCatalogEntry(
		"ocp-lifecycle",
		common.PlatformAlterationTestKey,
		`Tests that the running OCP version is not end of life.`,
		OCPLifecycleRemediation,
		NoExceptions,
		TestOCPLifecycleIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestPodHugePages1G = AddCatalogEntry(
		"hugepages-1g-only",
		common.PlatformAlterationTestKey,
		`Check that pods using hugepages only use 1Gi size`,
		PodHugePages1GRemediation,
		NoDocumentedProcess,
		TestPodHugePages1GDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagFarEdge)

	TestPodHugePages2M = AddCatalogEntry(
		"hugepages-2m-only",
		common.PlatformAlterationTestKey,
		`Check that pods using hugepages only use 2Mi size`,
		PodHugePages2MRemediation,
		NoExceptionProcessForExtendedTests,
		TestPodHugePages2MDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagExtended)

	TestServiceMeshIdentifier = AddCatalogEntry(
		"service-mesh-usage",
		common.PlatformAlterationTestKey,
		`Checks if the istio namespace ("istio-system") is present. If it is present, checks that the istio sidecar is present in all pods under test.`,
		ServiceMeshRemediation,
		NoExceptionProcessForExtendedTests,
		TestServiceMeshIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagExtended)

	TestSysctlConfigsIdentifier = AddCatalogEntry(
		"sysctl-config",
		common.PlatformAlterationTestKey,
		`Tests that no one has changed the node's sysctl configs after the node was created, the tests works by checking if the sysctl configs are consistent with the MachineConfig CR which defines how the node should be configured`,
		SysctlConfigsRemediation,
		NoExceptions,
		TestSysctlConfigsIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestUnalteredBaseImageIdentifier = AddCatalogEntry(
		"base-image",
		common.PlatformAlterationTestKey,
		`Ensures that the Container Base Image is not altered post-startup. This test is a heuristic, and ensures that there are no changes to the following directories: 1) /var/lib/rpm 2) /var/lib/dpkg 3) /bin 4) /sbin 5) /lib 6) /lib64 7) /usr/bin 8) /usr/sbin 9) /usr/lib 10) /usr/lib64`, //nolint:lll
		UnalteredBaseImageRemediation,
		NoExceptions,
		TestUnalteredBaseImageIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestUnalteredStartupBootParamsIdentifier = AddCatalogEntry(
		"boot-params",
		common.PlatformAlterationTestKey,
		`Tests that boot parameters are set through the MachineConfigOperator, and not set manually on the Node.`,
		UnalteredStartupBootParamsRemediation,
		NoExceptions,
		TestUnalteredStartupBootParamsIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)
}
