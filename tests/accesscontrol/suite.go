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

package accesscontrol

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksadapter"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	checksfn "github.com/redhat-best-practices-for-k8s/checks/accesscontrol"
	corev1 "k8s.io/api/core/v1"
)

var env provider.TestEnvironment

// LoadChecks loads all the checks.
//
//nolint:funlen
func LoadChecks() {
	log.Debug("Loading %s suite checks", common.AccessControlTestKey)

	checksGroup := checksdb.NewChecksGroup(common.AccessControlTestKey).
		WithBeforeEachFn(checksdb.DefaultBeforeEachFn(func() { env = provider.GetTestEnvironment() }))

	checksadapter.AddCheck(checksGroup, "access-control-security-context", checksfn.CheckSecurityContext, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-sys-admin-capability-check", checksfn.CheckSysAdmin, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-sys-module-capability-check", checksfn.CheckSysModule, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-net-admin-capability-check", checksfn.CheckNetAdmin, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-net-raw-capability-check", checksfn.CheckNetRaw, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-ipc-lock-capability-check", checksfn.CheckIPCLock, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-bpf-capability-check", checksfn.CheckBPF, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-dac-override-capability-check", checksfn.CheckDACOverride, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-dac-read-search-capability-check", checksfn.CheckDACReadSearch, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-security-context-non-root-user-id-check", checksfn.CheckNonRootUser, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-security-context-privilege-escalation", checksfn.CheckPrivilegeEscalation, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-security-context-read-only-root-file-system", checksfn.CheckReadOnlyFilesystem, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-container-host-port", checksfn.CheckContainerHostPort, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-pod-host-network", checksfn.CheckHostNetwork, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-pod-host-path", checksfn.CheckHostPath, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-pod-host-ipc", checksfn.CheckHostIPC, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-pod-host-pid", checksfn.CheckHostPID, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-namespace", checksfn.CheckNamespace, &env,
		testhelper.GetNoNamespacesSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-pod-service-account", checksfn.CheckServiceAccount, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-pod-role-bindings", checksfn.CheckRoleBindings, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-cluster-role-bindings", checksfn.CheckClusterRoleBindings, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-pod-automount-service-account-token", checksfn.CheckAutomountToken, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-one-process-per-container", checksfn.CheckOneProcess, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-sys-nice-realtime-capability", checksfn.CheckSysNiceRealtime, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetNoNodesWithRealtimeKernelSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-sys-ptrace-capability", checksfn.CheckSysPtrace, &env,
		testhelper.GetSharedProcessNamespacePodsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-namespace-resource-quota", checksfn.CheckNamespaceResourceQuota, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-ssh-daemons", checksfn.CheckNoSSHD, &env,
		testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-requests", checksfn.CheckPodRequests, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-no-1337-uid", checksfn.Check1337UID, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-service-type", checksfn.CheckNodePortService, &env,
		testhelper.GetNoServicesUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "access-control-crd-roles", checksfn.CheckCrdRoles, &env,
		testhelper.GetNoCrdsUnderTestSkipFn(&env), testhelper.GetNoNamespacesSkipFn(&env), testhelper.GetNoRolesSkipFn(&env))
}

// isContainerCapabilitySet checks whether a container capability was explicitly set
// in securityContext.capabilities.add list.
func isContainerCapabilitySet(containerCapabilities *corev1.Capabilities, capability string) bool {
	if containerCapabilities == nil {
		return false
	}

	if len(containerCapabilities.Add) == 0 {
		return false
	}

	if stringhelper.StringInSlice(containerCapabilities.Add, corev1.Capability("ALL"), true) ||
		stringhelper.StringInSlice(containerCapabilities.Add, corev1.Capability(capability), true) {
		return true
	}

	return false
}
