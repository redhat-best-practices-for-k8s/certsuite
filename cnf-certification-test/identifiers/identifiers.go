// Copyright (C) 2021-2023 Red Hat, Inc.
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
	"strings"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

const (
	InformativeResult        = "informative"
	NormativeResult          = "normative"
	bestPracticeDocV1dot3URL = "https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf"
	bestPracticeDocV1dot4URL = "https://TODO" // TODO: Fill in this variable with the new v1.4 document when available.
)

// shared description text
const (
	iptablesNftablesImplicitCheck = ` Note: this test ensures iptables and nftables are not configured by CNF pods:
- NET_ADMIN and NET_RAW are required to modify nftables (namespaced) which is not desired inside pods. 
nftables should be configured by an administrator outside the scope of the CNF. nftables are usually configured 
by operators, for instance the Performance Addon Operator (PAO) or istio.
- Privileged container are required to modify host iptables, which is not safe to perform inside pods. nftables 
should be configured by an administrator outside the scope of the CNF. iptables are usually configured by operators, 
for instance the Performance Addon Operator (PAO) or istio.`
)

const (
	TagCommon   = "common"
	TagExtended = "extended"
	TagTelco    = "telco"
	TagFarEdge  = "faredge"
)

func init() {
	InitCatalog()
}

func AddCatalogEntry(testID, suiteName, description, remediation, testType, exception, reference string, qe bool, categoryclassification claim.Categoryclassification, tags ...string) (aID claim.Identifier) {
	// Default Values (if missing)
	if strings.TrimSpace(exception) == "" {
		exception = NoDocumentedProcess
	}
	if strings.TrimSpace(reference) == "" {
		reference = "No Reference Document Specified"
	}
	if len(tags) == 0 {
		tags = append(tags, TagCommon)
	}

	tcDescription, aID := claim.BuildTestCaseDescription(testID, suiteName, description, remediation, testType, exception, reference, qe, categoryclassification, tags...)
	Catalog[aID] = tcDescription
	return aID
}

var (
	TestICMPv4ConnectivityIdentifier                  claim.Identifier
	TestNetworkPolicyDenyAllIdentifier                claim.Identifier
	Test1337UIDIdentifier                             claim.Identifier
	TestContainerIsCertifiedDigestIdentifier          claim.Identifier
	TestPodHugePages2M                                claim.Identifier
	TestPodHugePages1G                                claim.Identifier
	TestReservedExtendedPartnerPorts                  claim.Identifier
	TestAffinityRequiredPods                          claim.Identifier
	TestStartupIdentifier                             claim.Identifier
	TestShutdownIdentifier                            claim.Identifier
	TestDpdkCPUPinningExecProbe                       claim.Identifier
	TestSysAdminIdentifier                            claim.Identifier
	TestNetAdminIdentifier                            claim.Identifier
	TestNetRawIdentifier                              claim.Identifier
	TestIpcLockIdentifier                             claim.Identifier
	TestStorageRequiredPods                           claim.Identifier
	TestExclusiveCPUPoolIdentifier                    claim.Identifier
	TestSharedCPUPoolSchedulingPolicy                 claim.Identifier
	TestExclusiveCPUPoolSchedulingPolicy              claim.Identifier
	TestIsolatedCPUPoolSchedulingPolicy               claim.Identifier
	TestRtAppNoExecProbes                             claim.Identifier
	TestRestartOnRebootLabelOnPodsUsingSRIOV          claim.Identifier
	TestSecConNonRootUserIdentifier                   claim.Identifier
	TestSecContextIdentifier                          claim.Identifier
	TestSecConPrivilegeEscalation                     claim.Identifier
	TestContainerHostPort                             claim.Identifier
	TestPodHostNetwork                                claim.Identifier
	TestPodHostPath                                   claim.Identifier
	TestPodHostIPC                                    claim.Identifier
	TestPodHostPID                                    claim.Identifier
	TestContainerIsCertifiedIdentifier                claim.Identifier
	TestHugepagesNotManuallyManipulated               claim.Identifier
	TestICMPv6ConnectivityIdentifier                  claim.Identifier
	TestICMPv4ConnectivityMultusIdentifier            claim.Identifier
	TestICMPv6ConnectivityMultusIdentifier            claim.Identifier
	TestServiceDualStackIdentifier                    claim.Identifier
	TestNamespaceBestPracticesIdentifier              claim.Identifier
	TestNonTaintedNodeKernelsIdentifier               claim.Identifier
	TestOperatorInstallStatusSucceededIdentifier      claim.Identifier
	TestOperatorNoPrivileges                          claim.Identifier
	TestOperatorIsCertifiedIdentifier                 claim.Identifier
	TestHelmIsCertifiedIdentifier                     claim.Identifier
	TestOperatorIsInstalledViaOLMIdentifier           claim.Identifier
	TestPodNodeSelectorAndAffinityBestPractices       claim.Identifier
	TestPodHighAvailabilityBestPractices              claim.Identifier
	TestPodClusterRoleBindingsBestPracticesIdentifier claim.Identifier
	TestPodDeploymentBestPracticesIdentifier          claim.Identifier
	TestDeploymentScalingIdentifier                   claim.Identifier
	TestStateFulSetScalingIdentifier                  claim.Identifier
	TestImagePullPolicyIdentifier                     claim.Identifier
	TestPodRecreationIdentifier                       claim.Identifier
	TestPodRoleBindingsBestPracticesIdentifier        claim.Identifier
	TestPodServiceAccountBestPracticesIdentifier      claim.Identifier
	TestPodAutomountServiceAccountIdentifier          claim.Identifier
	TestServicesDoNotUseNodeportsIdentifier           claim.Identifier
	TestUnalteredBaseImageIdentifier                  claim.Identifier
	TestUnalteredStartupBootParamsIdentifier          claim.Identifier
	TestLoggingIdentifier                             claim.Identifier
	TestTerminationMessagePolicyIdentifier            claim.Identifier
	TestCrdsStatusSubresourceIdentifier               claim.Identifier
	TestSysctlConfigsIdentifier                       claim.Identifier
	TestServiceMeshIdentifier                         claim.Identifier
	TestOCPLifecycleIdentifier                        claim.Identifier
	TestNodeOperatingSystemIdentifier                 claim.Identifier
	TestIsRedHatReleaseIdentifier                     claim.Identifier
	TestIsSELinuxEnforcingIdentifier                  claim.Identifier
	TestUndeclaredContainerPortsUsage                 claim.Identifier
	TestOCPReservedPortsUsage                         claim.Identifier
	TestLivenessProbeIdentifier                       claim.Identifier
	TestReadinessProbeIdentifier                      claim.Identifier
	TestStartupProbeIdentifier                        claim.Identifier
	TestOneProcessPerContainerIdentifier              claim.Identifier
	TestSYSNiceRealtimeCapabilityIdentifier           claim.Identifier
	TestSysPtraceCapabilityIdentifier                 claim.Identifier
	TestPodRequestsAndLimitsIdentifier                claim.Identifier
	TestNamespaceResourceQuotaIdentifier              claim.Identifier
	TestPodDisruptionBudgetIdentifier                 claim.Identifier
	TestPodTolerationBypassIdentifier                 claim.Identifier
	TestPersistentVolumeReclaimPolicyIdentifier       claim.Identifier
	TestContainersImageTag                            claim.Identifier
	TestNoSSHDaemonsAllowedIdentifier                 claim.Identifier
	TestCPUIsolationIdentifier                        claim.Identifier
	TestContainerPortNameFormat                       claim.Identifier
	TestCrdScalingIdentifier                          claim.Identifier
	// Chaos Testing
	TestPodDeleteIdentifier claim.Identifier
)

//nolint:funlen
func InitCatalog() map[claim.Identifier]claim.TestCaseDescription {
	TestICMPv4ConnectivityIdentifier = AddCatalogEntry(
		"icmpv4-connectivity",
		common.NetworkingTestKey,
		`Checks that each CNF Container is able to communicate via ICMPv4 on the Default OpenShift network.
This test case requires the Deployment of the debug daemonset.`,
		`Ensure that the CNF is able to communicate via the Default OpenShift network. In some rare cases,
CNFs may require routing table changes in order to communicate over the Default network. To exclude
a particular pod from ICMPv4 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it.
The label value is not important, only its presence.`,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestNetworkPolicyDenyAllIdentifier = AddCatalogEntry(
		"network-policy-deny-all",
		common.NetworkingTestKey,
		`Check that network policies attached to namespaces running CNF pods contain a default deny-all rule for both ingress and egress traffic`,
		NetworkPolicyDenyAllRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 10.6",
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagCommon)

	Test1337UIDIdentifier = AddCatalogEntry(
		"no-1337-uid",
		common.AccessControlTestKey,
		`Checks that all pods are not using the securityContext UID 1337`,
		UID1337Remediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 4.6.24",
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagExtended)

	// TestContainerIsCertifiedDigestIdentifier tests whether the container has passed Container Certification.
	TestContainerIsCertifiedDigestIdentifier = AddCatalogEntry(
		"container-is-certified-digest",
		common.AffiliatedCertTestKey,
		`Tests whether container images that are autodiscovered have passed the Red Hat Container Certification Program by their digest(CCP).`,
		ContainerIsCertifiedRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 5.3.7",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagExtended, TagTelco)

	TestPodHugePages2M = AddCatalogEntry(
		"hugepages-2m-only",
		common.PlatformAlterationTestKey,
		`Check that pods using hugepages only use 2Mi size`,
		"Modify pod to consume 2Mi hugepages only",
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 3.5.4",
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagExtended)

	TestPodHugePages1G = AddCatalogEntry(
		"hugepages-1g-only",
		common.PlatformAlterationTestKey,
		`Check that pods using hugepages only use 1Gi size`,
		"Modify pod to consume 1Gi hugepages only",
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL, // TODO: link Far Edge spec document
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagFarEdge)

	TestReservedExtendedPartnerPorts = AddCatalogEntry(
		"reserved-partner-ports",
		common.NetworkingTestKey,
		`Checks that pods and containers are not consuming ports designated as reserved by partner`,
		ReservedPartnerPortsRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 4.6.24",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagExtended)

	TestAffinityRequiredPods = AddCatalogEntry(
		"affinity-required-pods",
		common.LifecycleTestKey,
		`Checks that affinity rules are in place if AffinityRequired: 'true' labels are set on Pods.`,
		AffinityRequiredRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 4.6.24",
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagExtended, TagTelco)

	TestStorageRequiredPods = AddCatalogEntry(
		"storage-required-pods",
		common.LifecycleTestKey,
		`Checks that pods do not place persistent volumes on local storage.`,
		StorageRequiredPods,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 4.6.24",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagExtended)

	TestStorageRequiredPods = AddCatalogEntry(
		"no-pvs-on-localstorage",
		common.LifecycleTestKey,
		`Checks that pods do not place persistent volumes on local storage.`,
		StorageRequiredPods,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 4.6.24",
		false,
		CatagoryClassification{
			FarEdge:   "Mandatory",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagExtended)

	TestStartupIdentifier = AddCatalogEntry(
		"container-startup",
		common.LifecycleTestKey,
		`Ensure that the containers lifecycle postStart management feature is configured.`,
		`PostStart is normally used to configure the container, set up dependencies, and
record the new creation. You could use this event to check that a required
API is available before the container’s main work begins. Kubernetes will
not change the container’s state to Running until the PostStart script has
executed successfully. For details, see https://www.containiq.com/post/kubernetes-container-lifecycle-events-and-hooks and
https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks`,
		NormativeResult,
		`Identify which pod is not conforming to the process and submit information as to why it cannot use a postStart startup specification.`,
		bestPracticeDocV1dot3URL+" Section 5.1.3, 12.2 and 12.5",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestShutdownIdentifier = AddCatalogEntry(
		"container-shutdown",
		common.LifecycleTestKey,
		`Ensure that the containers lifecycle preStop management feature is configured.`,
		`The preStop can be used to gracefully stop the container and clean resources (e.g., DB connection).
For details, see https://www.containiq.com/post/kubernetes-container-lifecycle-events-and-hooks and
https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks`,
		NormativeResult,
		`Identify which pod is not conforming to the process and submit information as to why it cannot use a preStop shutdown specification.`,
		bestPracticeDocV1dot3URL+" Section 5.1.3, 12.2 and 12.5",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestDpdkCPUPinningExecProbe = AddCatalogEntry(
		"dpdk-cpu-pinning-exec-probe",
		common.NetworkingTestKey,
		`If a CNF is doing CPI pinning, exec probes may not be used.`,
		DpdkCPUPinningExecProbeRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 4.6.24",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagExtended, TagTelco)

	TestNetAdminIdentifier = AddCatalogEntry(
		"net-admin-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use NET_ADMIN capability. `+iptablesNftablesImplicitCheck,
		SecConRemediation,
		NormativeResult,
		SecConCapabilitiesExceptionProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestSysAdminIdentifier = AddCatalogEntry(
		"sys-admin-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use SYS_ADMIN capability`,
		SecConRemediation,
		NormativeResult,
		SecConCapabilitiesExceptionProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestIpcLockIdentifier = AddCatalogEntry(
		"ipc-lock-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use IPC_LOCK capability`,
		SecConRemediation,
		NormativeResult,
		SecConCapabilitiesExceptionProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestNetRawIdentifier = AddCatalogEntry(
		"net-raw-capability-check",
		common.AccessControlTestKey,
		`Ensures that containers do not use NET_RAW capability. `+iptablesNftablesImplicitCheck,
		SecConRemediation,
		NormativeResult,
		SecConCapabilitiesExceptionProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestExclusiveCPUPoolIdentifier = AddCatalogEntry(
		"exclusive-cpu-pool",
		common.PerformanceTestKey,
		`Ensures that if one container in a Pod selects an exclusive CPU pool the rest select the same type of CPU pool`,
		ExclusiveCPUPoolRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL, // TODO: link Far Edge spec document
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagFarEdge)

	TestSharedCPUPoolSchedulingPolicy = AddCatalogEntry(
		"shared-cpu-pool-non-rt-scheduling-policy",
		common.PerformanceTestKey,
		`Ensures that if application workload runs in shared CPU pool, it chooses non-RT CPU schedule policy to always share the CPU with other applications and kernel threads.`,
		SharedCPUPoolSchedulingPolicyRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL, // TODO: link Far Edge spec document
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagFarEdge)

	TestExclusiveCPUPoolSchedulingPolicy = AddCatalogEntry(
		"exclusive-cpu-pool-rt-scheduling-policy",
		common.PerformanceTestKey,
		`Ensures that if application workload runs in exclusive CPU pool, it chooses RT CPU schedule policy and set the priority less than 10.`,
		ExclusiveCPUPoolSchedulingPolicyRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL, // TODO: link Far Edge spec document
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagFarEdge)

	TestIsolatedCPUPoolSchedulingPolicy = AddCatalogEntry(
		"isolated-cpu-pool-rt-scheduling-policy",
		common.PerformanceTestKey,
		`Ensures that a workload running in an application-isolated exclusive CPU pool selects a RT CPU scheduling policy`,
		IsolatedCPUPoolSchedulingPolicyRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL, // TODO: link Far Edge spec document
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagFarEdge)

	TestRtAppNoExecProbes = AddCatalogEntry(
		"rt-apps-no-exec-probes",
		common.PerformanceTestKey,
		`Ensures that if one container runs a real time application exec probes are not used`,
		RtAppNoExecProbesRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL, // TODO: link Far Edge spec document
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagFarEdge)

	TestRestartOnRebootLabelOnPodsUsingSRIOV = AddCatalogEntry(
		"restart-on-reboot-sriov-pod",
		common.NetworkingTestKey,
		`Ensures that the label restart-on-reboot exists on pods that use SRIOV network interfaces.`,
		SRIOVPodsRestartOnRebootLabelRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL, // TODO: link Far Edge spec document
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagFarEdge)

	TestSecConNonRootUserIdentifier = AddCatalogEntry(
		"security-context-non-root-user-check",
		common.AccessControlTestKey,
		`Checks the security context runAsUser parameter in pods and containers to make sure it is not set to uid root(0)`,
		SecConNonRootUserRemediation,
		NormativeResult,
		SecConNonRootUserExceptionProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		false,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagCommon)

	TestSecContextIdentifier = AddCatalogEntry(
		"security-context",
		common.AccessControlTestKey,
		`Checks the security context matches one of the 4 categories`,
		SecConRemediation,
		NormativeResult,
		SecConExceptionProcess,
		bestPracticeDocV1dot4URL+" Section 4.5",
		false,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagExtended)

	TestSecConPrivilegeEscalation = AddCatalogEntry(
		"security-context-privilege-escalation",
		common.AccessControlTestKey,
		`Checks if privileged escalation is enabled (AllowPrivilegeEscalation=true)`,
		SecConPrivilegeRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		false,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagCommon)

	TestContainerHostPort = AddCatalogEntry(
		"container-host-port",
		common.AccessControlTestKey,
		`Verifies if containers define a hostPort.`,
		ContainerHostPortRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.3.6",
		false,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagCommon)

	TestPodHostNetwork = AddCatalogEntry(
		"pod-host-network",
		common.AccessControlTestKey,
		`Verifies that the spec.HostNetwork parameter is not set (not present)`,
		PodHostNetworkRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.3.6",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestPodHostPath = AddCatalogEntry(
		"pod-host-path",
		common.AccessControlTestKey,
		`Verifies that the spec.HostPath parameter is not set (not present)`,
		PodHostPathRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.3.6",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestPodHostIPC = AddCatalogEntry(
		"pod-host-ipc",
		common.AccessControlTestKey,
		`Verifies that the spec.HostIpc parameter is set to false`,
		PodHostIPCRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.3.6",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestPodHostPID = AddCatalogEntry(
		"pod-host-pid",
		common.AccessControlTestKey,
		`Verifies that the spec.HostPid parameter is set to false`,
		PodHostPIDRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.3.6",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestContainerIsCertifiedIdentifier = AddCatalogEntry(
		"container-is-certified",
		common.AffiliatedCertTestKey,
		`Tests whether container images listed in the configuration file have passed the Red Hat Container Certification Program (CCP).`,
		ContainerIsCertifiedRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.3.7",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestHugepagesNotManuallyManipulated = AddCatalogEntry(
		"hugepages-config",
		common.PlatformAlterationTestKey,
		`Checks to see that HugePage settings have been configured through MachineConfig, and not manually on the
underlying Node.  This test case applies only to Nodes that are configured with the "worker" MachineConfigSet.  First,
the "worker" MachineConfig is polled, and the Hugepage settings are extracted.  Next, the underlying Nodes are polled
for configured HugePages through inspection of /proc/meminfo.  The results are compared, and the test passes only if
they are the same.`,
		HugepagesNotManuallyManipulatedRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestICMPv6ConnectivityIdentifier = AddCatalogEntry(
		"icmpv6-connectivity",
		common.NetworkingTestKey,
		`Checks that each CNF Container is able to communicate via ICMPv6 on the Default OpenShift network.  This
test case requires the Deployment of the debug daemonset.`,
		ICMPv6ConnectivityRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestICMPv4ConnectivityMultusIdentifier = AddCatalogEntry(
		"icmpv4-connectivity-multus",
		common.NetworkingTestKey,
		`Checks that each CNF Container is able to communicate via ICMPv4 on the Multus network(s).  This
test case requires the Deployment of the debug daemonset.`,
		ICMPv4ConnectivityMultusRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestICMPv6ConnectivityMultusIdentifier = AddCatalogEntry(
		"icmpv6-connectivity-multus",
		common.NetworkingTestKey,
		`Checks that each CNF Container is able to communicate via ICMPv6 on the Multus network(s).  This
test case requires the Deployment of the debug daemonset.`,
		ICMPv6ConnectivityMultusRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestServiceDualStackIdentifier = AddCatalogEntry(
		"dual-stack-service",
		common.NetworkingTestKey,
		`Checks that all services in namespaces under test are either ipv6 single stack or dual stack. This
test case requires the deployment of the debug daemonset.`,
		TestServiceDualStackRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 3.5.7",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagExtended)

	TestNamespaceBestPracticesIdentifier = AddCatalogEntry(
		"namespace",
		common.AccessControlTestKey,
		`Tests that all CNF's resources (PUTs and CRs) belong to valid namespaces. A valid namespace meets
the following conditions: (1) It was declared in the yaml config file under the targetNameSpaces
tag. (2) It doesn't have any of the following prefixes: default, openshift-, istio- and aspenmesh-`,
		NamespaceBestPracticesRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2, 16.3.8 and 16.3.9",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestNonTaintedNodeKernelsIdentifier = AddCatalogEntry(
		"tainted-node-kernel",
		common.PlatformAlterationTestKey,
		`Ensures that the Node(s) hosting CNFs do not utilize tainted kernels. This test case is especially important
to support Highly Available CNFs, since when a CNF is re-instantiated on a backup Node, that Node's kernel may not have
the same hacks.'`,
		NonTaintedNodeKernelsRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.14",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestOperatorInstallStatusSucceededIdentifier = AddCatalogEntry(
		"install-status-succeeded",
		common.OperatorTestKey,
		`Ensures that the target CNF operators report "Succeeded" as their installation status.`,
		OperatorInstallStatusSucceededRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.12 and 5.3.3",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestOperatorNoPrivileges = AddCatalogEntry(
		"install-status-no-privileges",
		common.OperatorTestKey,
		`The operator is not installed with privileged rights. Test passes if clusterPermissions is not present in the CSV manifest or is present
with no resourceNames under its rules.`,
		OperatorNoPrivilegesRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.12 and 5.3.3",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestOperatorIsCertifiedIdentifier = AddCatalogEntry(
		"operator-is-certified",
		common.AffiliatedCertTestKey,
		`Tests whether CNF Operators listed in the configuration file have passed the Red Hat Operator Certification Program (OCP).`,
		OperatorIsCertifiedRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.12 and 5.3.3",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestHelmIsCertifiedIdentifier = AddCatalogEntry(
		"helmchart-is-certified",
		common.AffiliatedCertTestKey,
		`Tests whether helm charts listed in the cluster passed the Red Hat Helm Certification Program.`,
		HelmIsCertifiedRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.12 and 5.3.3",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestOperatorIsInstalledViaOLMIdentifier = AddCatalogEntry(
		"install-source",
		common.OperatorTestKey,
		`Tests whether a CNF Operator is installed via OLM.`,
		OperatorIsInstalledViaOLMRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.12 and 5.3.3",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestPodNodeSelectorAndAffinityBestPractices = AddCatalogEntry(
		"pod-scheduling",
		common.LifecycleTestKey,
		`Ensures that CNF Pods do not specify nodeSelector or nodeAffinity.  In most cases, Pods should allow for
instantiation on any underlying Node.`,
		PodNodeSelectorAndAffinityBestPracticesRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestPodHighAvailabilityBestPractices = AddCatalogEntry(
		"pod-high-availability",
		common.LifecycleTestKey,
		`Ensures that CNF Pods specify podAntiAffinity rules and replica value is set to more than 1.`,
		PodHighAvailabilityBestPracticesRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestPodClusterRoleBindingsBestPracticesIdentifier = AddCatalogEntry(
		"cluster-role-bindings",
		common.AccessControlTestKey,
		`Tests that a Pod does not specify ClusterRoleBindings.`,
		PodClusterRoleBindingsBestPracticesRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.10 and 5.3.6",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestPodDeploymentBestPracticesIdentifier = AddCatalogEntry(
		"pod-owner-type",
		common.LifecycleTestKey,
		`Tests that CNF Pod(s) are deployed as part of a ReplicaSet(s)/StatefulSet(s).`,
		PodDeploymentBestPracticesRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.3.3 and 5.3.8",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestDeploymentScalingIdentifier = AddCatalogEntry(
		"deployment-scaling",
		common.LifecycleTestKey,
		`Tests that CNF deployments support scale in/out operations.
            First, The test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the
            scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s.
            In case of deployments that are managed by HPA the test is changing the min and max value to deployment Replica - 1 during scale-in and the
            original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the deployment/s`,
		DeploymentScalingRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestStateFulSetScalingIdentifier = AddCatalogEntry(
		"statefulset-scaling",
		common.LifecycleTestKey,
		`Tests that CNF statefulsets support scale in/out operations.
            First, The test starts getting the current replicaCount (N) of the statefulset/s with the Pod Under Test. Then, it executes the
            scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the statefulset/s.
            In case of statefulsets that are managed by HPA the test is changing the min and max value to statefulset Replica - 1 during scale-in and the
            original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the statefulset/s`,
		StatefulSetScalingRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestImagePullPolicyIdentifier = AddCatalogEntry(
		"image-pull-policy",
		common.LifecycleTestKey,
		`Ensure that the containers under test are using IfNotPresent as Image Pull Policy.`,
		ImagePullPolicyRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+"  Section 12.6",
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagCommon, TagTelco)

	TestPodRecreationIdentifier = AddCatalogEntry(
		"pod-recreation",
		common.LifecycleTestKey,
		`Tests that a CNF is configured to support High Availability.
            First, this test cordons and drains a Node that hosts the CNF Pod.
            Next, the test ensures that OpenShift can re-instantiate the Pod on another Node,
            and that the actual replica count matches the desired replica count.`,
		PodRecreationRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestPodRoleBindingsBestPracticesIdentifier = AddCatalogEntry(
		"pod-role-bindings",
		common.AccessControlTestKey,
		`Ensures that a CNF does not utilize RoleBinding(s) in a non-CNF Namespace.`,
		PodRoleBindingsBestPracticesRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.3.3 and 5.3.5",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestPodServiceAccountBestPracticesIdentifier = AddCatalogEntry(
		"pod-service-account",
		common.AccessControlTestKey,
		`Tests that each CNF Pod utilizes a valid Service Account.`,
		PodServiceAccountBestPracticesRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.3 and 5.2.7",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestPodAutomountServiceAccountIdentifier = AddCatalogEntry(
		"pod-automount-service-account-token",
		common.AccessControlTestKey,
		`Check that all pods under test have automountServiceAccountToken set to false. Only pods that require access to the kubernetes API server should have automountServiceAccountToken set to true`,
		AutomountServiceTokenRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 12.7",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestServicesDoNotUseNodeportsIdentifier = AddCatalogEntry(
		"service-type",
		common.NetworkingTestKey,
		`Tests that each CNF Service does not utilize NodePort(s).`,
		ServicesDoNotUseNodeportsRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.3.1",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestUnalteredBaseImageIdentifier = AddCatalogEntry(
		"base-image",
		common.PlatformAlterationTestKey,
		`Ensures that the Container Base Image is not altered post-startup.  This test is a heuristic, and ensures
that there are no changes to the following directories:
1) /var/lib/rpm
2) /var/lib/dpkg
3) /bin
4) /sbin
5) /lib
6) /lib64
7) /usr/bin
8) /usr/sbin
9) /usr/lib
10) /usr/lib64`,
		UnalteredBaseImageRemediation,
		NormativeResult,
		UnalteredBaseImageExceptionProcess,
		bestPracticeDocV1dot3URL+" Section 5.1.4",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestUnalteredStartupBootParamsIdentifier = AddCatalogEntry(
		"boot-params",
		common.PlatformAlterationTestKey,
		`Tests that boot parameters are set through the MachineConfigOperator, and not set manually on the Node.`,
		UnalteredStartupBootParamsRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.13 and 5.2.14",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestLoggingIdentifier = AddCatalogEntry(
		"container-logging",
		common.ObservabilityTestKey,
		`Check that all containers under test use standard input output and standard error when logging`,
		LoggingRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 10.1",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestTerminationMessagePolicyIdentifier = AddCatalogEntry(
		"termination-policy",
		common.ObservabilityTestKey,
		`Check that all containers are using terminationMessagePolicy: FallbackToLogsOnError`,
		TerminationMessagePolicyRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 12.1",
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagCommon)

	TestCrdsStatusSubresourceIdentifier = AddCatalogEntry(
		"crd-status",
		common.ObservabilityTestKey,
		`Checks that all CRDs have a status subresource specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties["status"]).`,
		CrdsStatusSubresourceRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestSysctlConfigsIdentifier = AddCatalogEntry(
		"sysctl-config",
		common.PlatformAlterationTestKey,
		`Tests that no one has changed the node's sysctl configs after the node
            was created, the tests works by checking if the sysctl configs are consistent with the
            MachineConfig CR which defines how the node should be configured`,
		SysctlConfigsRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestServiceMeshIdentifier = AddCatalogEntry(
		"service-mesh-usage",
		common.PlatformAlterationTestKey,
		`Checks if the istio namespace ("istio-system") is present. If it is present, checks that the istio sidecar is present in all pods under test.`,
		ServiceMeshRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagCommon)

	TestOCPLifecycleIdentifier = AddCatalogEntry(
		"ocp-lifecycle",
		common.PlatformAlterationTestKey,
		`Tests that the running OCP version is not end of life.`,
		OCPLifecycleRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 7.9",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestNodeOperatingSystemIdentifier = AddCatalogEntry(
		"ocp-node-os-lifecycle",
		common.PlatformAlterationTestKey,
		`Tests that the nodes running in the cluster have operating systems
            that are compatible with the deployed version of OpenShift.`,
		NodeOperatingSystemRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 7.9",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestIsRedHatReleaseIdentifier = AddCatalogEntry(
		"isredhat-release",
		common.PlatformAlterationTestKey,
		`verifies if the container base image is redhat.`,
		IsRedHatReleaseRemediation,
		NormativeResult,
		IsRedHatReleaseExceptionProcess,
		bestPracticeDocV1dot3URL+" Section 5.2",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestIsSELinuxEnforcingIdentifier = AddCatalogEntry(
		"is-selinux-enforcing",
		common.PlatformAlterationTestKey,
		`verifies that all openshift platform/cluster nodes have selinux in "Enforcing" mode.`,
		IsSELinuxEnforcingRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 10.3 Pod Security",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestUndeclaredContainerPortsUsage = AddCatalogEntry(
		"undeclared-container-ports-usage",
		common.NetworkingTestKey,
		`Check that containers do not listen on ports that weren't declared in their specification`,
		UndeclaredContainerPortsRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 16.3.1.1",
		false,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestOCPReservedPortsUsage = AddCatalogEntry(
		"ocp-reserved-ports-usage",
		common.NetworkingTestKey,
		`Check that containers do not listen on ports that are reserved by OpenShift`,
		OCPReservedPortsUsageRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 3.5.9",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon)

	TestLivenessProbeIdentifier = AddCatalogEntry(
		"liveness-probe",
		common.LifecycleTestKey,
		`Check that all containers under test have liveness probe defined`,
		LivenessProbeRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.16, 12.1 and 12.5",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestReadinessProbeIdentifier = AddCatalogEntry(
		"readiness-probe",
		common.LifecycleTestKey,
		`Check that all containers under test have readiness probe defined`,
		ReadinessProbeRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 5.2.16, 12.1 and 12.5",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestStartupProbeIdentifier = AddCatalogEntry(
		"startup-probe",
		common.LifecycleTestKey,
		`Check that all containers under test have startup probe defined`,
		StartupProbeRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 4.6.12", // TODO Change this to v1.4 when available
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestOneProcessPerContainerIdentifier = AddCatalogEntry(
		"one-process-per-container",
		common.AccessControlTestKey,
		`Check that all containers under test have only one process running`,
		OneProcessPerContainerRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 10.8.3",
		false,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagCommon)

	TestSYSNiceRealtimeCapabilityIdentifier = AddCatalogEntry(
		"sys-nice-realtime-capability",
		common.AccessControlTestKey,
		`Check that pods running on nodes with realtime kernel enabled have the SYS_NICE capability enabled in their spec.`,
		SYSNiceRealtimeCapabilityRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 2.7.4",
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestSysPtraceCapabilityIdentifier = AddCatalogEntry(
		"sys-ptrace-capability",
		common.AccessControlTestKey,
		`Check that if process namespace sharing is enabled for a Pod then the SYS_PTRACE capability is allowed`,
		SysPtraceCapabilityRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 2.7.5",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestPodRequestsAndLimitsIdentifier = AddCatalogEntry(
		"requests-and-limits",
		common.AccessControlTestKey,
		`Check that containers have resource requests and limits specified in their spec.`,
		RequestsAndLimitsRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 4.6.11",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestNamespaceResourceQuotaIdentifier = AddCatalogEntry(
		"namespace-resource-quota",
		common.AccessControlTestKey,
		`Checks to see if CNF workload pods are running in namespaces that have resource quotas applied.`,
		NamespaceResourceQuotaRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 4.6.8", // TODO Change this to v1.4 when available
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestPodDisruptionBudgetIdentifier = AddCatalogEntry(
		"pod-disruption-budget",
		common.ObservabilityTestKey,
		`Checks to see if pod disruption budgets have allowed values for minAvailable and maxUnavailable`,
		PodDisruptionBudgetRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 4.6.20", // TODO Change this to v1.4 when available
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestPodTolerationBypassIdentifier = AddCatalogEntry(
		"pod-toleration-bypass",
		common.LifecycleTestKey,
		`Check that pods do not have NoExecute, PreferNoSchedule, or NoSchedule tolerations that have been modified from the default.`,
		PodTolerationBypassRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 10.6",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestPersistentVolumeReclaimPolicyIdentifier = AddCatalogEntry(
		"persistent-volume-reclaim-policy",
		common.LifecycleTestKey,
		`Check that the persistent volumes the CNF pods are using have a reclaim policy of delete.`,
		PersistentVolumeReclaimPolicyRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 3.3.4",
		true,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagCommon, TagTelco)

	TestContainersImageTag = AddCatalogEntry(
		"containers-image-tag",
		common.ManageabilityTestKey,
		`Check that image tag exists on containers.`,
		ContainersImageTagRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 4.6.12",
		false,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optional",
		},
		TagExtended)

	TestNoSSHDaemonsAllowedIdentifier = AddCatalogEntry(
		"ssh-daemons",
		common.AccessControlTestKey,
		`Check that pods do not run SSH daemons.`,
		NoSSHDaemonsAllowedRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot3URL+" Section 4.6.12", // TODO Change this to v1.4 when available
		false,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestCPUIsolationIdentifier = AddCatalogEntry(
		"cpu-isolation",
		common.LifecycleTestKey,
		`CPU isolation requires: For each container within the pod, resource requests and limits must be identical.
        Request and Limits are in the form of whole CPUs. The runTimeClassName must be specified. Annotations required disabling CPU and IRQ load-balancing.`,
		CPUIsolationRemediation,
		InformativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 3.5.5",
		true,
		claim.Categoryclassification{
			FarEdge:   "Mandatory",
			Telco:     "Mandatory",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagCommon, TagTelco)

	TestContainerPortNameFormat = AddCatalogEntry(
		"container-port-name-format",
		common.ManageabilityTestKey,
		`Check that the container's ports name follow the naming conventions.`,
		ContainerPortNameFormatRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 4.6.20",
		false,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Mandatory",
		},
		TagExtended)

	TestCrdScalingIdentifier = AddCatalogEntry(
		"crd-scaling",
		common.LifecycleTestKey,
		`Tests that CNF crd support scale in/out operations.
                First, The test starts getting the current replicaCount (N) of the crd/s with the Pod Under Test. Then, it executes the
                scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the crd/s.
                In case of crd that are managed by HPA the test is changing the min and max value to crd Replica - 1 during scale-in and the
                original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the crd/s`,
		CrdScalingRemediation,
		NormativeResult,
		NoDocumentedProcess,
		bestPracticeDocV1dot4URL+" Section 4.6.20",
		false,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Mandatory",
			NoneTelco: "Mandatory",
			Extended:  "Mandatory",
		},
		TagCommon,
	)

	TestPodDeleteIdentifier = AddCatalogEntry(
		"pod-delete",
		common.ChaosTesting,
		"",
		"",
		"",
		NoDocumentedProcess,
		"",
		false,
		claim.Categoryclassification{
			FarEdge:   "Optional",
			Telco:     "Optional",
			NoneTelco: "Optional",
			Extended:  "Optionalx",
		},
		TagCommon)

	return Catalog
}

var (
	// TestIdToClaimId converts the testcase short ID to the claim identifier
	TestIDToClaimID = map[string]claim.Identifier{}
)

// GetGinkgoTestIDAndLabels transform the claim.Identifier into a test Id that can be used to skip
// specific tests
func GetGinkgoTestIDAndLabels(identifier claim.Identifier) (testID string, tags []string) {
	tags = strings.Split(identifier.Tags, ",")
	tags = append(tags, identifier.Id, identifier.Suite)
	TestIDToClaimID[identifier.Id] = identifier
	return identifier.Id, tags
}

// Catalog is the JUnit testcase catalog of tests.
var Catalog = map[claim.Identifier]claim.TestCaseDescription{}
