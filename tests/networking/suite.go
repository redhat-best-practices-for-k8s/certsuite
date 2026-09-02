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

package networking

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksadapter"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	checksfn "github.com/redhat-best-practices-for-k8s/checks/networking"
)

var env provider.TestEnvironment

func LoadChecks() {
	log.Debug("Loading %s suite checks", common.NetworkingTestKey)

	checksGroup := checksdb.NewChecksGroup(common.NetworkingTestKey).
		WithBeforeEachFn(checksdb.DefaultBeforeEachFn(func() { env = provider.GetTestEnvironment() }))

	// Default interface ICMP IPv4 test case
	checksadapter.AddCheck(checksGroup, "networking-icmpv4-connectivity", checksfn.CheckICMPv4Connectivity, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env))

	// Multus interfaces ICMP IPv4 test case
	checksadapter.AddCheck(checksGroup, "networking-icmpv4-connectivity-multus", checksfn.CheckICMPv4ConnectivityMultus, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env))

	// Default interface ICMP IPv6 test case
	checksadapter.AddCheck(checksGroup, "networking-icmpv6-connectivity", checksfn.CheckICMPv6Connectivity, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env))

	// Multus interfaces ICMP IPv6 test case
	checksadapter.AddCheck(checksGroup, "networking-icmpv6-connectivity-multus", checksfn.CheckICMPv6ConnectivityMultus, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env))

	// Undeclared container ports usage test case
	checksadapter.AddCheck(checksGroup, "networking-undeclared-container-ports-usage", checksfn.CheckUndeclaredContainerPorts, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env))

	// OCP reserved ports usage test case
	checksadapter.AddCheck(checksGroup, "networking-ocp-reserved-ports-usage", checksfn.CheckOCPReservedPorts, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env))

	// Dual stack services test case
	checksadapter.AddCheck(checksGroup, "networking-dual-stack-service", checksfn.CheckDualStackService, &env,
		testhelper.GetNoServicesUnderTestSkipFn(&env))

	// Network policy deny all test case
	checksadapter.AddCheck(checksGroup, "networking-network-policy-deny-all", checksfn.CheckNetworkPolicyDenyAll, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	// Extended partner ports test case
	checksadapter.AddCheck(checksGroup, "networking-reserved-partner-ports", checksfn.CheckReservedPartnerPorts, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env))

	// Restart on reboot label test case
	checksadapter.AddCheck(checksGroup, "networking-restart-on-reboot-sriov-pod", checksfn.CheckSRIOVRestartLabel, &env,
		testhelper.GetNoSRIOVPodsSkipFn(&env))

	// SRIOV MTU test case
	checksadapter.AddCheck(checksGroup, "networking-network-attachment-definition-sriov-mtu", checksfn.CheckSRIOVNetworkAttachmentDefinitionMTU, &env,
		testhelper.GetNoSRIOVPodsSkipFn(&env))

	// TLS minimum version test case
	checksadapter.AddCheck(checksGroup, "networking-tls-minimum-version", checksfn.CheckTLSMinimumVersion, &env,
		testhelper.GetNoServicesUnderTestSkipFn(&env),
		testhelper.GetDaemonSetFailedToSpawnSkipFn(&env),
		testhelper.GetOCPVersionBelowSkipFn(&env, checksfn.OCPTLSProfileEnforcementVersion))

	// Unsecured container ports test case
	checksadapter.AddCheck(checksGroup, "networking-unsecured-container-ports", checksfn.CheckUnsecuredContainerPorts, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env),
		testhelper.GetDaemonSetFailedToSpawnSkipFn(&env),
		testhelper.GetNoPodsUnderTestSkipFn(&env))
}
