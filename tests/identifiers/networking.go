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
	TestICMPv4ConnectivityIdentifier             claim.Identifier
	TestICMPv4ConnectivityMultusIdentifier       claim.Identifier
	TestICMPv6ConnectivityIdentifier             claim.Identifier
	TestICMPv6ConnectivityMultusIdentifier       claim.Identifier
	TestNetworkAttachmentDefinitionSRIOVUsingMTU claim.Identifier
	TestNetworkPolicyDenyAllIdentifier           claim.Identifier
	TestOCPReservedPortsUsage                    claim.Identifier
	TestReservedExtendedPartnerPorts             claim.Identifier
	TestRestartOnRebootLabelOnPodsUsingSRIOV     claim.Identifier
	TestServiceDualStackIdentifier               claim.Identifier
	TestTLSMinimumVersionIdentifier              claim.Identifier
	TestUndeclaredContainerPortsUsage            claim.Identifier
)

//nolint:funlen
func init() {
	TestICMPv4ConnectivityIdentifier = AddCatalogEntry(
		"icmpv4-connectivity",
		common.NetworkingTestKey,
		`Checks that each workload Container is able to communicate via ICMPv4 on the Default OpenShift network. This test case requires the Deployment of the probe daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.`,                                                          //nolint:lll
		`Ensure that the workload is able to communicate via the Default OpenShift network. In some rare cases, workloads may require routing table changes in order to communicate over the Default network. To exclude a particular pod from ICMPv4 connectivity tests, add the redhat-best-practices-for-k8s.com/skip_connectivity_tests label to it. The label value is trivial, only its presence.`, //nolint:lll
		`No exceptions - must be able to communicate on default network using IPv4`,
		TestICMPv4ConnectivityIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestICMPv4ConnectivityMultusIdentifier = AddCatalogEntry(
		"icmpv4-connectivity-multus",
		common.NetworkingTestKey,
		`Checks that each workload Container is able to communicate via ICMPv4 on the Multus network(s). This test case requires the Deployment of the probe daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.`, //nolint:lll
		ICMPv4ConnectivityMultusRemediation,
		NoDocumentedProcess,
		TestICMPv4ConnectivityMultusIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestICMPv6ConnectivityIdentifier = AddCatalogEntry(
		"icmpv6-connectivity",
		common.NetworkingTestKey,
		`Checks that each workload Container is able to communicate via ICMPv6 on the Default OpenShift network. This test case requires the Deployment of the probe daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.`, //nolint:lll
		ICMPv6ConnectivityRemediation,
		NoDocumentedProcess,
		TestICMPv6ConnectivityIdentifierDocLink,
		false,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagCommon)

	TestICMPv6ConnectivityMultusIdentifier = AddCatalogEntry(
		"icmpv6-connectivity-multus",
		common.NetworkingTestKey,
		`Checks that each workload Container is able to communicate via ICMPv6 on the Multus network(s). This test case requires the Deployment of the probe daemonset and at least 2 pods connected to each network under test(one source and one destination). If no network with more than 2 pods exists this test will be skipped.`, //nolint:lll
		ICMPv6ConnectivityMultusRemediation+` Not applicable if IPv6/MULTUS is not supported.`,
		NoDocumentedProcess,
		TestICMPv6ConnectivityMultusIdentifierDocLink,
		false,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestNetworkAttachmentDefinitionSRIOVUsingMTU = AddCatalogEntry(
		"network-attachment-definition-sriov-mtu",
		common.NetworkingTestKey,
		`Ensures that MTU values are set correctly in NetworkAttachmentDefinitions for SRIOV network interfaces.`,
		SRIOVNetworkAttachmentDefinitionMTURemediation,
		NoDocumentedProcess,
		TestNetworkAttachmentDefinitionSRIOVUsingMTUDocLink,
		false,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagFarEdge)

	TestNetworkPolicyDenyAllIdentifier = AddCatalogEntry(
		"network-policy-deny-all",
		common.NetworkingTestKey,
		`Check that network policies attached to namespaces running workload pods contain a default deny-all rule for both ingress and egress traffic`,
		NetworkPolicyDenyAllRemediation,
		NoExceptionProcessForExtendedTests,
		TestNetworkPolicyDenyAllIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagCommon)

	TestOCPReservedPortsUsage = AddCatalogEntry(
		"ocp-reserved-ports-usage",
		common.NetworkingTestKey,
		`Check that containers do not listen on ports that are reserved by OpenShift`,
		OCPReservedPortsUsageRemediation,
		NoExceptions,
		TestOCPReservedPortsUsageDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)
	TestReservedExtendedPartnerPorts = AddCatalogEntry(
		"reserved-partner-ports",
		common.NetworkingTestKey,
		`Checks that pods and containers are not consuming ports designated as reserved by partner`,
		ReservedPartnerPortsRemediation,
		NoExceptionProcessForExtendedTests,
		TestReservedExtendedPartnerPortsDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagExtended)

	TestRestartOnRebootLabelOnPodsUsingSRIOV = AddCatalogEntry(
		"restart-on-reboot-sriov-pod",
		common.NetworkingTestKey,
		`Ensures that the label restart-on-reboot exists on pods that use SRIOV network interfaces.`,
		SRIOVPodsRestartOnRebootLabelRemediation,
		NoDocumentedProcess,
		TestRestartOnRebootLabelOnPodsUsingSRIOVDocLink,
		false,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagFarEdge)

	TestServiceDualStackIdentifier = AddCatalogEntry(
		"dual-stack-service",
		common.NetworkingTestKey,
		`Checks that all services in namespaces under test are either ipv6 single stack or dual stack. This test case requires the deployment of the probe daemonset.`,
		TestServiceDualStackRemediation,
		NoExceptionProcessForExtendedTests,
		TestServiceDualStackIdentifierDocLink,
		false,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagExtended)

	TestTLSMinimumVersionIdentifier = AddCatalogEntry(
		"tls-minimum-version",
		common.NetworkingTestKey,
		`Checks that TLS-enabled services in target namespaces honor the cluster's TLS security profile. `+
			`On OpenShift, the profile is read from the APIServer CR (default: Intermediate, min TLS 1.2). `+
			`On non-OpenShift clusters, Intermediate is used as default. `+
			`Validates both minimum TLS version and cipher suite compliance. `+
			`Non-TLS ports are reported as informational only. `+
			`Note: this test is skipped on OpenShift clusters running versions below 4.22.`,
		TLSMinimumVersionRemediation,
		NoExceptionProcessForExtendedTests,
		TestTLSMinimumVersionIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagExtended)

	TestUndeclaredContainerPortsUsage = AddCatalogEntry(
		"undeclared-container-ports-usage",
		common.NetworkingTestKey,
		`Check that containers do not listen on ports that weren't declared in their specification. Platforms may be configured to block undeclared ports.`,
		UndeclaredContainerPortsRemediation,
		NoExceptionProcessForExtendedTests,
		TestUndeclaredContainerPortsUsageDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagExtended)
}
