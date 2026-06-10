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

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}
)

func LoadChecks() {
	log.Debug("Loading %s suite checks", common.NetworkingTestKey)

	checksGroup := checksdb.NewChecksGroup(common.NetworkingTestKey).
		WithBeforeEachFn(beforeEachFn)

	// Default interface ICMP IPv4 test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-icmpv4-connectivity")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckICMPv4Connectivity).MakeCheckFn(&env)))

	// Multus interfaces ICMP IPv4 test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-icmpv4-connectivity-multus")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckICMPv4ConnectivityMultus).MakeCheckFn(&env)))

	// Default interface ICMP IPv6 test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-icmpv6-connectivity")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckICMPv6Connectivity).MakeCheckFn(&env)))

	// Multus interfaces ICMP IPv6 test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-icmpv6-connectivity-multus")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckICMPv6ConnectivityMultus).MakeCheckFn(&env)))

	// Undeclared container ports usage test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-undeclared-container-ports-usage")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckUndeclaredContainerPorts).MakeCheckFn(&env)))

	// OCP reserved ports usage test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-ocp-reserved-ports-usage")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckOCPReservedPorts).MakeCheckFn(&env)))

	// Dual stack services test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-dual-stack-service")).
		WithSkipCheckFn(testhelper.GetNoServicesUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckDualStackService).MakeCheckFn(&env)))

	// Network policy deny all test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-network-policy-deny-all")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckNetworkPolicyDenyAll).MakeCheckFn(&env)))

	// Extended partner ports test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-reserved-partner-ports")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckReservedPartnerPorts).MakeCheckFn(&env)))

	// Restart on reboot label test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-restart-on-reboot-sriov-pod")).
		WithSkipCheckFn(testhelper.GetNoSRIOVPodsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckSRIOVRestartLabel).MakeCheckFn(&env)))

	// SRIOV MTU test case
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("networking-network-attachment-definition-sriov-mtu")).
		WithSkipCheckFn(testhelper.GetNoSRIOVPodsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckSRIOVNetworkAttachmentDefinitionMTU).MakeCheckFn(&env)))
}
