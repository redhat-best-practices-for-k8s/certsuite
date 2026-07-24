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

const (
	iptablesNftablesImplicitCheck = `Note: this test also ensures iptables and nftables are not configured by workload pods:
- NET_ADMIN and NET_RAW are required to modify nftables (namespaced) which is not desired inside pods.
nftables should be configured by an administrator outside the scope of the workload. nftables are usually configured
by operators, for instance the Performance Addon Operator (PAO) or istio.
- Privileged container are required to modify host iptables, which is not safe to perform inside pods. nftables
should be configured by an administrator outside the scope of the workload. iptables are usually configured by operators,
for instance the Performance Addon Operator (PAO) or istio.`
)

var (
	Test1337UIDIdentifier                             claim.Identifier
	TestBpfIdentifier                                 claim.Identifier
	TestContainerHostPort                             claim.Identifier
	TestCrdRoleIdentifier                             claim.Identifier
	TestDacOverrideIdentifier                         claim.Identifier
	TestDacReadSearchIdentifier                       claim.Identifier
	TestIpcLockIdentifier                             claim.Identifier
	TestNamespaceBestPracticesIdentifier              claim.Identifier
	TestNamespaceResourceQuotaIdentifier              claim.Identifier
	TestNetAdminIdentifier                            claim.Identifier
	TestNetRawIdentifier                              claim.Identifier
	TestNoSSHDaemonsAllowedIdentifier                 claim.Identifier
	TestOneProcessPerContainerIdentifier              claim.Identifier
	TestPodAutomountServiceAccountIdentifier          claim.Identifier
	TestPodClusterRoleBindingsBestPracticesIdentifier claim.Identifier
	TestPodHostIPC                                    claim.Identifier
	TestPodHostNetwork                                claim.Identifier
	TestPodHostPID                                    claim.Identifier
	TestPodHostPath                                   claim.Identifier
	TestPodRequestsIdentifier                         claim.Identifier
	TestPodRoleBindingsBestPracticesIdentifier        claim.Identifier
	TestPodServiceAccountBestPracticesIdentifier      claim.Identifier
	TestSYSNiceRealtimeCapabilityIdentifier           claim.Identifier
	TestSecConNonRootUserIDIdentifier                 claim.Identifier
	TestSecConPrivilegeEscalation                     claim.Identifier
	TestSecConReadOnlyFilesystem                      claim.Identifier
	TestSecContextIdentifier                          claim.Identifier
	TestServicesDoNotUseNodeportsIdentifier           claim.Identifier
	TestSysAdminIdentifier                            claim.Identifier
	TestSysModuleIdentifier                           claim.Identifier
	TestSysPtraceCapabilityIdentifier                 claim.Identifier
)

//nolint:funlen
func init() {
	Test1337UIDIdentifier = AddCatalogEntry(
		"no-1337-uid",
		common.AccessControlTestKey,
		`Checks that all pods are not using the securityContext UID 1337`,
		UID1337Remediation,
		NoExceptionProcessForExtendedTests,
		Test1337UIDIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagExtended)

	TestBpfIdentifier = AddCatalogEntry(
		"bpf-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use BPF capability. Workloads should avoid loading eBPF filters`,
		BpfCapabilityRemediation,
		`Exception can be considered. Must identify which container requires the capability and detail why.`,
		TestBpfIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestContainerHostPort = AddCatalogEntry(
		"container-host-port",
		common.AccessControlTestKey,
		`Verifies if containers define a hostPort.`,
		ContainerHostPortRemediation,
		"Exception for host resource access tests will only be considered in rare cases where it is absolutely needed",
		TestContainerHostPortDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestCrdRoleIdentifier = AddCatalogEntry(
		"crd-roles",
		common.AccessControlTestKey,
		"If an application creates CRDs it must supply a role to access those CRDs and no other API resources/permission. This test checks that there is at least one role present in each namespaces under test that only refers to CRDs under test.",
		"Roles providing access to CRDs should not refer to any other api or resources. Change the generation of the CRD role accordingly",
		NoExceptionProcessForExtendedTests,
		"https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-custom-role-to-access-application-crds",
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagExtended)

	TestDacOverrideIdentifier = AddCatalogEntry(
		"dac-override-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use DAC_OVERRIDE capability. DAC_OVERRIDE bypasses file permission checks and usually indicates incorrect file ownership in the container image.`,
		DacOverrideCapabilityRemediation,
		NoExceptions,
		TestDacOverrideIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestDacReadSearchIdentifier = AddCatalogEntry(
		"dac-read-search-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use DAC_READ_SEARCH capability. DAC_READ_SEARCH enables open_by_handle_at()-style access and is a known container escape risk.`,
		DacReadSearchCapabilityRemediation,
		NoExceptions,
		TestDacReadSearchIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestIpcLockIdentifier = AddCatalogEntry(
		"ipc-lock-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use IPC_LOCK capability. Workloads should avoid accessing host resources - spec.HostIpc should be false.`,
		SecConRemediation,
		`Exception possible if a workload uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and detail why.`,
		TestIpcLockIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestNamespaceBestPracticesIdentifier = AddCatalogEntry(
		"namespace",
		common.AccessControlTestKey,
		`Tests that all workload resources (PUTs and CRs) belong to valid namespaces. A valid namespace meets
the following conditions: (1) It was declared in the yaml config file under the targetNameSpaces
tag. (2) It does not have any of the following prefixes: default, openshift-, istio- and aspenmesh-`,
		NamespaceBestPracticesRemediation,
		NoExceptions,
		TestNamespaceBestPracticesIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagCommon)

	TestNamespaceResourceQuotaIdentifier = AddCatalogEntry(
		"namespace-resource-quota",
		common.AccessControlTestKey,
		`Checks to see if workload pods are running in namespaces that have resource quotas applied.`,
		NamespaceResourceQuotaRemediation,
		NoExceptionProcessForExtendedTests,
		TestNamespaceResourceQuotaIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagExtended)

	TestNetAdminIdentifier = AddCatalogEntry(
		"net-admin-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use NET_ADMIN capability. `+iptablesNftablesImplicitCheck,
		SecConRemediation,
		`Exception will be considered for user plane or networking functions (e.g. SR-IOV, Multicast). Must identify which container requires the capability and detail why.`,
		TestNetAdminIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestNetRawIdentifier = AddCatalogEntry(
		"net-raw-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use NET_RAW capability. `+iptablesNftablesImplicitCheck,
		SecConRemediation,
		`Exception will be considered for user plane or networking functions. Must identify which container requires the capability and detail why.`,
		TestNetRawIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestNoSSHDaemonsAllowedIdentifier = AddCatalogEntry(
		"ssh-daemons",
		common.AccessControlTestKey,
		`Check that pods do not run SSH daemons.`,
		NoSSHDaemonsAllowedRemediation,
		`No exceptions - special consideration can be given to certain containers which run as utility tool daemon`,
		TestNoSSHDaemonsAllowedIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestOneProcessPerContainerIdentifier = AddCatalogEntry(
		"one-process-per-container",
		common.AccessControlTestKey,
		`Check that all containers under test have only one process running`,
		OneProcessPerContainerRemediation,
		NoExceptionProcessForExtendedTests+NotApplicableSNO,
		TestOneProcessPerContainerIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagCommon)

	TestPodAutomountServiceAccountIdentifier = AddCatalogEntry(
		"pod-automount-service-account-token",
		common.AccessControlTestKey,
		`Check that all pods under test have automountServiceAccountToken set to false. Only pods that require access to the kubernetes API server should have automountServiceAccountToken set to true`,
		AutomountServiceTokenRemediation,
		`Exception will be considered if container needs to access APIs which OCP does not offer natively. Must document which container requires which API(s) and detail why existing OCP APIs cannot be used.`,
		TestPodAutomountServiceAccountIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestPodClusterRoleBindingsBestPracticesIdentifier = AddCatalogEntry(
		"cluster-role-bindings",
		common.AccessControlTestKey,
		`Tests that a Pod does not specify ClusterRoleBindings.`,
		PodClusterRoleBindingsBestPracticesRemediation,
		"Exception possible only for workloads that's cluster wide in nature and absolutely needs cluster level roles & role bindings",
		TestPodClusterRoleBindingsBestPracticesIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestPodHostIPC = AddCatalogEntry(
		"pod-host-ipc",
		common.AccessControlTestKey,
		`Verifies that the spec.HostIpc parameter is set to false`,
		PodHostIPCRemediation,
		`Exception for host resource access tests will only be considered in rare cases where it is absolutely needed`,
		TestPodHostIPCDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestPodHostNetwork = AddCatalogEntry(
		"pod-host-network",
		common.AccessControlTestKey,
		`Verifies that the spec.HostNetwork parameter is not set (not present)`,
		PodHostNetworkRemediation,
		`Exception for host resource access tests will only be considered in rare cases where it is absolutely needed`,
		TestPodHostNetworkDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestPodHostPID = AddCatalogEntry(
		"pod-host-pid",
		common.AccessControlTestKey,
		`Verifies that the spec.HostPid parameter is set to false`,
		PodHostPIDRemediation,
		`Exception for host resource access tests will only be considered in rare cases where it is absolutely needed`,
		TestPodHostPIDDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestPodHostPath = AddCatalogEntry(
		"pod-host-path",
		common.AccessControlTestKey,
		`Verifies that the spec.HostPath parameter is not set (not present)`,
		PodHostPathRemediation,
		`Exception for host resource access tests will only be considered in rare cases where it is absolutely needed`,
		TestPodHostPathDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestPodRequestsIdentifier = AddCatalogEntry(
		"requests",
		common.AccessControlTestKey,
		`Check that containers have resource requests specified in their spec. Set proper resource requests based on container use case.`,
		RequestsRemediation,
		RequestsExceptionProcess,
		TestPodRequestsIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestPodRoleBindingsBestPracticesIdentifier = AddCatalogEntry(
		"pod-role-bindings",
		common.AccessControlTestKey,
		`Ensures that a workload does not utilize RoleBinding(s) in a non-workload Namespace.`,
		PodRoleBindingsBestPracticesRemediation,
		NoExceptions,
		TestPodRoleBindingsBestPracticesIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestPodServiceAccountBestPracticesIdentifier = AddCatalogEntry(
		"pod-service-account",
		common.AccessControlTestKey,
		`Tests that each workload Pod utilizes a valid Service Account. Default or empty service account is not valid.`,
		PodServiceAccountBestPracticesRemediation,
		NoExceptions,
		TestPodServiceAccountBestPracticesIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestSYSNiceRealtimeCapabilityIdentifier = AddCatalogEntry(
		"sys-nice-realtime-capability",
		common.AccessControlTestKey,
		`Check that pods running on nodes with realtime kernel enabled have the SYS_NICE capability enabled in their spec. In the case that a workolad is running on a node using the real-time kernel, SYS_NICE will be used to allow DPDK application to switch to SCHED_FIFO.`, //nolint:lll
		SYSNiceRealtimeCapabilityRemediation,
		NoDocumentedProcess,
		TestSYSNiceRealtimeCapabilityIdentifierDocLink,
		false,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)

	TestSecConNonRootUserIDIdentifier = AddCatalogEntry(
		"security-context-non-root-user-id-check",
		common.AccessControlTestKey,
		`Checks securityContext's runAsNonRoot and runAsUser fields at pod and container level to make sure containers are not run as root.`,
		SecConRunAsNonRootUserRemediation,
		SecConNonRootUserExceptionProcess,
		TestSecConNonRootUserIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestSecConPrivilegeEscalation = AddCatalogEntry(
		"security-context-privilege-escalation",
		common.AccessControlTestKey,
		`Checks if privileged escalation is enabled (AllowPrivilegeEscalation=true).`,
		SecConPrivilegeRemediation,
		NoExceptions,
		TestSecConPrivilegeEscalationDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestSecConReadOnlyFilesystem = AddCatalogEntry(
		"security-context-read-only-root-file-system",
		common.AccessControlTestKey,
		`Checks the security context readOnlyFileSystem in containers is enabled. Containers should not try modify its own filesystem.`,
		SecConNonRootUserExceptionProcess,
		NoExceptions,
		TestSecContextIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Optional,
		},
		TagCommon)

	TestSecContextIdentifier = AddCatalogEntry(
		"security-context",
		common.AccessControlTestKey,
		`Checks the security context matches one of the 4 categories`,
		`Exception possible if a workload uses mlock(), mlockall(), shmctl(), mmap(); exception will be considered for DPDK applications. Must identify which container requires the capability and document why. If the container had the right configuration of the allowed category from the 4 approved list then the test will pass. The 4 categories are defined in [Requirement ID 94118](#security-context-categories)`, //nolint:lll
		`no exception needed for optional/extended test`,
		TestSecContextIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Optional,
			Telco:    Optional,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagExtended)

	TestServicesDoNotUseNodeportsIdentifier = AddCatalogEntry(
		"service-type",
		common.AccessControlTestKey,
		`Tests that each workload Service does not utilize NodePort(s).`,
		ServicesDoNotUseNodeportsRemediation,
		`Exception for host resource access tests will only be considered in rare cases where it is absolutely needed`,
		TestServicesDoNotUseNodeportsIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestSysAdminIdentifier = AddCatalogEntry(
		"sys-admin-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use SYS_ADMIN capability`,
		SecConRemediation+" Containers should not use the SYS_ADMIN Linux capability.",
		NoExceptions,
		TestSysAdminIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestSysModuleIdentifier = AddCatalogEntry(
		"sys-module-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use SYS_MODULE capability. SYS_MODULE allows loading kernel modules from a container and creates a host/cluster takeover risk.`,
		SysModuleCapabilityRemediation,
		NoExceptions,
		TestSysModuleIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Mandatory,
			Extended: Mandatory,
		},
		TagCommon)

	TestSysPtraceCapabilityIdentifier = AddCatalogEntry(
		"sys-ptrace-capability",
		common.AccessControlTestKey,
		`Check that if process namespace sharing is enabled for a Pod then the SYS_PTRACE capability is allowed. This capability is required when using Process Namespace Sharing. This is used when processes from one Container need to be exposed to another Container. For example, to send signals like SIGHUP from a process in a Container to another process in another Container. For more information on these capabilities refer to https://cloud.redhat.com/blog/linux-capabilities-in-openshift and https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/`, //nolint:lll
		SysPtraceCapabilityRemediation,
		NoDocumentedProcess,
		TestSysPtraceCapabilityIdentifierDocLink,
		true,
		map[string]string{
			FarEdge:  Mandatory,
			Telco:    Mandatory,
			NonTelco: Optional,
			Extended: Mandatory,
		},
		TagTelco)
}
