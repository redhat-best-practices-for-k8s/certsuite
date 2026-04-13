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

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}
)

// LoadChecks loads all the checks.
//
//nolint:funlen
func LoadChecks() {
	log.Debug("Loading %s suite checks", common.AccessControlTestKey)

	checksGroup := checksdb.NewChecksGroup(common.AccessControlTestKey).
		WithBeforeEachFn(beforeEachFn)

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-security-context")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckSecurityContext).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-sys-admin-capability-check")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckSysAdmin).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-net-admin-capability-check")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckNetAdmin).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-net-raw-capability-check")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckNetRaw).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-ipc-lock-capability-check")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckIPCLock).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-bpf-capability-check")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckBPF).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-security-context-non-root-user-id-check")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckNonRootUser).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-security-context-privilege-escalation")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckPrivilegeEscalation).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-security-context-read-only-file-system")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckReadOnlyFilesystem).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-container-host-port")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckContainerHostPort).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-pod-host-network")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckHostNetwork).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-pod-host-path")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckHostPath).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-pod-host-ipc")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckHostIPC).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-pod-host-pid")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckHostPID).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-namespace")).
		WithSkipCheckFn(testhelper.GetNoNamespacesSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckNamespace).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-pod-service-account")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckServiceAccount).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-pod-role-bindings")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckRoleBindings).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-cluster-role-bindings")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckClusterRoleBindings).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-pod-automount-service-account-token")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckAutomountToken).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-one-process-per-container")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckOneProcess).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-sys-nice-realtime-capability")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithSkipCheckFn(testhelper.GetNoNodesWithRealtimeKernelSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckSysNiceRealtime).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-sys-ptrace-capability")).
		WithSkipCheckFn(testhelper.GetSharedProcessNamespacePodsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckSysPtrace).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-namespace-resource-quota")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckNamespaceResourceQuota).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-ssh-daemons")).
		WithSkipCheckFn(testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckNoSSHD).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-requests")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckPodRequests).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-no-1337-uid")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.Check1337UID).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-service-type")).
		WithSkipCheckFn(testhelper.GetNoServicesUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckNodePortService).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("access-control-crd-roles")).
		WithSkipCheckFn(testhelper.GetNoCrdsUnderTestSkipFn(&env), testhelper.GetNoNamespacesSkipFn(&env), testhelper.GetNoRolesSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckCrdRoles).MakeCheckFn(&env)))
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
