// Copyright (C) 2021-2022 Red Hat, Inc.
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
	"fmt"
	"strings"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

const (
	InformativeResult        = "informative"
	NormativeResult          = "normative"
	url                      = "http://test-network-function.com/testcases"
	VersionOne               = "v1.0.0"
	bestPracticeDocV1dot3URL = "https://connect.redhat.com/sites/default/files/2022-05/Cloud%20Native%20Network%20Function%20Requirements%201-3.pdf"
	bestPracticeDocV1dot4URL = "https://TODO" // TODO: Fill in this variable with the new v1.4 document when available.
)

const (
	TagCommon   = "common"
	TagExtended = "extended"
	TagOnline   = "online"
)

// TestCaseDescription describes a JUnit test case.
type TestCaseDescription struct {
	// Identifier is the unique test identifier.
	Identifier claim.Identifier `json:"identifier" yaml:"identifier"`

	// Description is a helpful description of the purpose of the test case.
	Description string `json:"description" yaml:"description"`

	// Remediation is an optional suggested remediation for passing the test.
	Remediation string `json:"remediation,omitempty" yaml:"remediation,omitempty"`

	// Type is the type of the test (i.e., normative).
	Type string `json:"type" yaml:"type"`

	// BestPracticeReference is a helpful best practice references of the test case.
	BestPracticeReference string `json:"BestPracticeReference" yaml:"BestPracticeReference"`

	// ExceptionProcess will show any possible exception processes documented for partners to follow.
	ExceptionProcess string `json:"exceptionProcess,omitempty" yaml:"exceptionProcess,omitempty"`

	// Tags will show all of the ginkgo tags that the test case applies to
	Tags string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func init() {
	InitCatalog()
}

func formTestURL(suite, name string) string {
	return fmt.Sprintf("%s/%s/%s", url, suite, name)
}

func formTestTags(tags ...string) string {
	return strings.Join(tags, ",")
}

func AddCatalogEntry(testID, suiteName, description, remediation, testType, exception, version, reference string, tags ...string) (aID claim.Identifier) {
	// Default Values (if missing)
	if strings.TrimSpace(exception) == "" {
		exception = NoDocumentedProcess
	}
	if strings.TrimSpace(version) == "" {
		version = VersionOne
	}
	if strings.TrimSpace(reference) == "" {
		reference = "No Reference Document Specified"
	}
	if len(tags) == 0 {
		tags = append(tags, TagCommon)
	}

	aID = claim.Identifier{
		Tags:    formTestTags(tags...),
		Url:     formTestURL(suiteName, testID),
		Version: version,
	}

	aTCDescription := TestCaseDescription{}
	aTCDescription.Identifier = aID
	aTCDescription.Type = testType
	aTCDescription.Description = formDescription(aID, description)
	aTCDescription.Remediation = remediation
	aTCDescription.ExceptionProcess = exception
	aTCDescription.BestPracticeReference = reference
	aTCDescription.Tags = strings.Join(tags, ",")

	Catalog[aID] = aTCDescription

	return aID
}

var (
	TestICMPv4ConnectivityIdentifier         claim.Identifier
	TestNetworkPolicyDenyAllIdentifier       claim.Identifier
	Test1337UIDIdentifier                    claim.Identifier
	TestContainerIsCertifiedDigestIdentifier claim.Identifier
	TestPodHugePages2M                       claim.Identifier
	TestReservedExtendedPartnerPorts         claim.Identifier
	TestAffinityRequiredPods                 claim.Identifier
	TestStartupIdentifier                    claim.Identifier
	TestShutdownIdentifier                   claim.Identifier
	TestDpdkCPUPinningExecProbe              claim.Identifier
)

//nolint:funlen
func InitCatalog() map[claim.Identifier]TestCaseDescription {
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
		VersionOne,
		bestPracticeDocV1dot3URL+" Section 5.2",
		TagCommon)

	TestNetworkPolicyDenyAllIdentifier = AddCatalogEntry(
		"network-policy-deny-all",
		common.NetworkingTestKey,
		`Check that network policies attached to namespaces running CNF pods contain a default deny-all rule for both ingress and egress traffic`,
		NetworkPolicyDenyAllRemediation,
		InformativeResult,
		NoDocumentedProcess,
		VersionOne,
		bestPracticeDocV1dot3URL+" Section 10.6",
		TagCommon)

	Test1337UIDIdentifier = AddCatalogEntry(
		"no-1337-uid",
		common.AccessControlTestKey,
		`Checks that all pods are not using the securityContext UID 1337`,
		UID1337Remediation,
		InformativeResult,
		NoDocumentedProcess,
		VersionOne,
		bestPracticeDocV1dot4URL+" Section 4.6.24",
		TagExtended)

	// TestContainerIsCertifiedDigestIdentifier tests whether the container has passed Container Certification.
	TestContainerIsCertifiedDigestIdentifier = AddCatalogEntry(
		"container-is-certified-digest",
		common.AffiliatedCertTestKey,
		`Tests whether container images that are autodiscovered have passed the Red Hat Container Certification Program by their digest(CCP).`,
		ContainerIsCertifiedRemediation,
		NormativeResult,
		NoDocumentedProcess,
		VersionOne,
		bestPracticeDocV1dot4URL+" Section 5.3.7",
		TagExtended)
	TestPodHugePages2M = AddCatalogEntry(
		"hugepages-2m-only",
		common.PlatformAlterationTestKey,
		`Check that pods using hugepages only use 2Mi size`,
		"Modify pod to consume 2Mi hugepages only",
		NormativeResult,
		NoDocumentedProcess,
		VersionOne,
		bestPracticeDocV1dot4URL+" Section 3.5.4",
		TagExtended)

	TestReservedExtendedPartnerPorts = AddCatalogEntry(
		"reserved-partner-ports",
		common.NetworkingTestKey,
		`Checks that pods and containers are not consuming ports designated as reserved by partner`,
		ReservedPartnerPortsRemediation,
		InformativeResult,
		NoDocumentedProcess,
		VersionOne,
		bestPracticeDocV1dot4URL+" Section 4.6.24",
		TagExtended)

	TestAffinityRequiredPods = AddCatalogEntry(
		"affinity-required-pods",
		common.LifecycleTestKey,
		`Checks that affinity rules are in place if AffinityRequired: 'true' labels are set on Pods.`,
		AffinityRequiredRemediation,
		InformativeResult,
		NoDocumentedProcess,
		VersionOne,
		bestPracticeDocV1dot4URL+" Section 4.6.24",
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
		VersionOne,
		bestPracticeDocV1dot3URL+" Section 5.1.3, 12.2 and 12.5",
		TagCommon)

	TestShutdownIdentifier = AddCatalogEntry(
		"container-shutdown",
		common.LifecycleTestKey,
		`Ensure that the containers lifecycle preStop management feature is configured.`,
		`The preStop can be used to gracefully stop the container and clean resources (e.g., DB connection).
For details, see https://www.containiq.com/post/kubernetes-container-lifecycle-events-and-hooks and
https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks`,
		NormativeResult,
		`Identify which pod is not conforming to the process and submit information as to why it cannot use a preStop shutdown specification.`,
		VersionOne,
		bestPracticeDocV1dot3URL+" Section 5.1.3, 12.2 and 12.5",
		TagCommon)

	TestDpdkCPUPinningExecProbe = AddCatalogEntry(
		"dpdk-cpu-pinning-exec-probe",
		common.NetworkingTestKey,
		`If a CNF is doing CPI pinning, exec probes may not be used.`,
		DpdkCPUPinningExecProbeRemediation,
		InformativeResult,
		NoDocumentedProcess,
		VersionOne,
		bestPracticeDocV1dot4URL+" Section 4.6.24",
		TagExtended)

	return Catalog
}

var (
	// TestIdToClaimId converts the testcase short ID to the claim identifier
	TestIDToClaimID = map[string]claim.Identifier{}

	// TestPodDeleteIdentifier tests for delete pod test
	TestPodDeleteIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.ChaosTesting, "pod-delete"),
		Version: VersionOne,
	}

	// BaseDomain for the test cases
	TestIDBaseDomain = url

	// TestSecConCapabilitiesIdentifier tests for non compliant security context capabilities
	TestSecConCapabilitiesIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "security-context-capabilities-check"),
		Version: VersionOne,
	}
	// TestSecConNonRootUserIdentifier tests that pods or containers are not running with root permissions
	TestSecConNonRootUserIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "security-context-non-root-user-check"),
		Version: VersionOne,
	}
	// TestSecContextIdentifier tests that pods or containers are not running with root permissions
	TestSecContextIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "security-context"),
		Version: VersionOne,
	}
	// TestSecPrivilegedEscalation tests that containers are not allowed privilege escalation
	TestSecConPrivilegeEscalation = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "security-context-privilege-escalation"),
		Version: VersionOne,
	}
	// TestSecPrivilegedEscalation tests that containers are not configured with host port privileges
	TestContainerHostPort = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "container-host-port"),
		Version: VersionOne,
	}
	// TestPodHostNetwork tests that pods do not configure hostnetwork to true
	TestPodHostNetwork = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "pod-host-network"),
		Version: VersionOne,
	}
	// TestPodHostPath tests that pods do not configure a hostpath volume
	TestPodHostPath = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "pod-host-path"),
		Version: VersionOne,
	}
	// TestPodHostPath tests that pods do not configure a hostpath volume
	TestPodHostIPC = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "pod-host-ipc"),
		Version: VersionOne,
	}
	// TestPodHostPath tests that pods do not configure a hostpath volume
	TestPodHostPID = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "pod-host-pid"),
		Version: VersionOne,
	}
	// TestContainerIsCertifiedIdentifier tests whether the container has passed Container Certification.
	TestContainerIsCertifiedIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon, TagOnline),
		Url:     formTestURL(common.AffiliatedCertTestKey, "container-is-certified"),
		Version: VersionOne,
	}
	// TestHugepagesNotManuallyManipulated represents the test identifier testing hugepages have not been manipulated.
	TestHugepagesNotManuallyManipulated = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.PlatformAlterationTestKey, "hugepages-config"),
		Version: VersionOne,
	}
	// TestICMPv6ConnectivityIdentifier tests icmpv6 connectivity.
	TestICMPv6ConnectivityIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.NetworkingTestKey, "icmpv6-connectivity"),
		Version: VersionOne,
	}
	// TestICMPv4ConnectivityIdentifier tests icmpv4 Multus connectivity.
	TestICMPv4ConnectivityMultusIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.NetworkingTestKey, "icmpv4-connectivity-multus"),
		Version: VersionOne,
	}
	// TestICMPv6ConnectivityIdentifier tests icmpv6 Multus connectivity.
	TestICMPv6ConnectivityMultusIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.NetworkingTestKey, "icmpv6-connectivity-multus"),
		Version: VersionOne,
	}
	// TestServiceDualStack verifies that all services under test are either ipv6 single stack or dual-stack
	TestServiceDualStackIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.NetworkingTestKey, "dual-stack-service"),
		Version: VersionOne,
	}
	// TestNFTablesIdentifier verifies that there is no nftable configuration in any containers of the CNF
	TestNFTablesIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.NetworkingTestKey, "nftables"),
		Version: VersionOne,
	}
	// TestIPTablesIdentifier verifies that there is no iptables configuration in any containers of the CNF
	TestIPTablesIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.NetworkingTestKey, "iptables"),
		Version: VersionOne,
	}
	// TestNamespaceBestPracticesIdentifier ensures the namespace has followed best namespace practices.
	TestNamespaceBestPracticesIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "namespace"),
		Version: VersionOne,
	}
	// TestNonTaintedNodeKernelsIdentifier is the identifier for the test checking tainted nodes.
	TestNonTaintedNodeKernelsIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.PlatformAlterationTestKey, "tainted-node-kernel"),
		Version: VersionOne,
	}
	// TestOperatorInstallStatusSucceededIdentifier tests Operator best practices.
	TestOperatorInstallStatusSucceededIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.OperatorTestKey, "install-status-succeeded"),
		Version: VersionOne,
	}
	// TestOperatorNoPrivileges tests Operator has no privileges on resources.
	TestOperatorNoPrivileges = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.OperatorTestKey, "install-status-no-privileges"),
		Version: VersionOne,
	}
	// TestOperatorIsCertifiedIdentifier tests that an Operator has passed Operator certification.
	TestOperatorIsCertifiedIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon, TagOnline),
		Url:     formTestURL(common.AffiliatedCertTestKey, "operator-is-certified"),
		Version: VersionOne,
	}
	TestHelmIsCertifiedIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon, TagOnline),
		Url:     formTestURL(common.AffiliatedCertTestKey, "helmchart-is-certified"),
		Version: VersionOne,
	}
	// TestOperatorIsInstalledViaOLMIdentifier tests that an Operator is installed via OLM.
	TestOperatorIsInstalledViaOLMIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.OperatorTestKey, "install-source"),
		Version: VersionOne,
	}
	// TestPodNodeSelectorAndAffinityBestPractices is the test ensuring nodeSelector and nodeAffinity are not used by a
	// Pod.
	TestPodNodeSelectorAndAffinityBestPractices = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "pod-scheduling"),
		Version: VersionOne,
	}
	// TestPodHighAvailabilityBestPractices is the test ensuring podAntiAffinity are used by a
	// Pod when pod replica # are great than 1
	TestPodHighAvailabilityBestPractices = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "pod-high-availability"),
		Version: VersionOne,
	}

	// TestPodClusterRoleBindingsBestPracticesIdentifier ensures Pod crb best practices.
	TestPodClusterRoleBindingsBestPracticesIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "cluster-role-bindings"),
		Version: VersionOne,
	}
	// TestPodDeploymentBestPracticesIdentifier ensures a CNF follows best Deployment practices.
	TestPodDeploymentBestPracticesIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "pod-owner-type"),
		Version: VersionOne,
	}
	// TestDeploymentScalingIdentifier ensures deployment scale in/out operations work correctly.
	TestDeploymentScalingIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "deployment-scaling"),
		Version: VersionOne,
	}
	// TestStateFulSetScalingIdentifier ensures statefulset scale in/out operations work correctly.
	TestStateFulSetScalingIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "statefulset-scaling"),
		Version: VersionOne,
	}
	// TestImagePullPolicyIdentifier ensures represent image pull policy practices.
	TestImagePullPolicyIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "image-pull-policy"),
		Version: VersionOne,
	}
	// TestPodRecreationIdentifier ensures recreation best practices.
	TestPodRecreationIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "pod-recreation"),
		Version: VersionOne,
	}
	// TestPodRoleBindingsBestPracticesIdentifier represents rb best practices.
	TestPodRoleBindingsBestPracticesIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "pod-role-bindings"),
		Version: VersionOne,
	}
	// TestPodServiceAccountBestPracticesIdentifier tests Pod SA best practices.
	TestPodServiceAccountBestPracticesIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "pod-service-account"),
		Version: VersionOne,
	}
	//
	TestPodAutomountServiceAccountIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "pod-automount-service-account-token"),
		Version: VersionOne,
	}
	// TestServicesDoNotUseNodeportsIdentifier ensures Services do not utilize NodePorts.
	TestServicesDoNotUseNodeportsIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.NetworkingTestKey, "service-type"),
		Version: VersionOne,
	}
	// TestUnalteredBaseImageIdentifier ensures the base image is not altered.
	TestUnalteredBaseImageIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.PlatformAlterationTestKey, "base-image"),
		Version: VersionOne,
	}
	// TestUnalteredStartupBootParamsIdentifier ensures startup boot params are not altered.
	TestUnalteredStartupBootParamsIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.PlatformAlterationTestKey, "boot-params"),
		Version: VersionOne,
	}
	// TestLoggingIdentifier ensures stderr/stdout are used
	TestLoggingIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.ObservabilityTestKey, "container-logging"),
		Version: VersionOne,
	}
	// TestTerminationMessagePolicy ensures pods have FallbackToLogsOnError set
	TestTerminationMessagePolicyIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.ObservabilityTestKey, "termination-policy"),
		Version: VersionOne,
	}
	// TestCrdsStatusSubresourceIdentifier ensures all CRDs have a valid status subresource
	TestCrdsStatusSubresourceIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.ObservabilityTestKey, "crd-status"),
		Version: VersionOne,
	}
	// TestSysctlConfigsIdentifier ensures that the node's sysctl configs are consistent with the MachineConfig CR
	TestSysctlConfigsIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.PlatformAlterationTestKey, "sysctl-config"),
		Version: VersionOne,
	}
	// TestServiceMesh checks if service mesh is exist.
	TestServiceMeshIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.PlatformAlterationTestKey, "service-mesh-usage"),
		Version: VersionOne,
	}
	// TestOCPLifecycleIdentifier ensures the OCP version of the cluster is within the valid lifecycle status
	TestOCPLifecycleIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.PlatformAlterationTestKey, "ocp-lifecycle"),
		Version: VersionOne,
	}
	// TestNodeOperatingSystemIdentifier ensures workers in the cluster have an operating system that is compatible with the installed version of OpenShift
	TestNodeOperatingSystemIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.PlatformAlterationTestKey, "ocp-node-os-lifecycle"),
		Version: VersionOne,
	}
	// TestScalingIdentifier ensures deployment scale in/out operations work correctly.
	TestScalingIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "scaling"),
		Version: VersionOne,
	}
	// TestIsRedHatReleaseIdentifier ensures platform is defined
	TestIsRedHatReleaseIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.PlatformAlterationTestKey, "isredhat-release"),
		Version: VersionOne,
	}
	// TestIsSELinuxEnforcingIdentifier ensures selinux is in enforcing mode
	TestIsSELinuxEnforcingIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.PlatformAlterationTestKey, "is-selinux-enforcing"),
		Version: VersionOne,
	}
	TestUndeclaredContainerPortsUsage = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.NetworkingTestKey, "undeclared-container-ports-usage"),
		Version: VersionOne,
	}
	TestOCPReservedPortsUsage = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.NetworkingTestKey, "ocp-reserved-ports-usage"),
		Version: VersionOne,
	}
	TestLivenessProbeIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "liveness-probe"),
		Version: VersionOne,
	}
	TestReadinessProbeIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "readiness-probe"),
		Version: VersionOne,
	}
	TestStartupProbeIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "startup-probe"),
		Version: VersionOne,
	}
	// TestOneProcessPerContainerIdentifier ensures that only one process per container is running
	TestOneProcessPerContainerIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "one-process-per-container"),
		Version: VersionOne,
	}
	TestSYSNiceRealtimeCapabilityIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "sys-nice-realtime-capability"),
		Version: VersionOne,
	}
	// TestSysPtraceCapabilityIdentifier ensures that if process namespace sharing is enabled then the SYS_PTRACE capability is allowed
	TestSysPtraceCapabilityIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "sys-ptrace-capability"),
		Version: VersionOne,
	}
	TestPodRequestsAndLimitsIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "requests-and-limits"),
		Version: VersionOne,
	}
	TestNamespaceResourceQuotaIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "namespace-resource-quota"),
		Version: VersionOne,
	}
	TestPodDisruptionBudgetIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.ObservabilityTestKey, "pod-disruption-budget"),
		Version: VersionOne,
	}
	TestPodTolerationBypassIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "pod-toleration-bypass"),
		Version: VersionOne,
	}
	TestPersistentVolumeReclaimPolicyIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "persistent-volume-reclaim-policy"),
		Version: VersionOne,
	}
	TestContainersImageTag = claim.Identifier{
		Tags:    formTestTags(TagExtended),
		Url:     formTestURL(common.ManageabilityTestKey, "containers-image-tag"),
		Version: VersionOne,
	}
	TestNoSSHDaemonsAllowedIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.AccessControlTestKey, "ssh-daemons"),
		Version: VersionOne,
	}
	TestCPUIsolationIdentifier = claim.Identifier{
		Tags:    formTestTags(TagCommon),
		Url:     formTestURL(common.LifecycleTestKey, "cpu-isolation"),
		Version: VersionOne,
	}
	TestContainerPortNameFormat = claim.Identifier{
		Tags:    formTestTags(TagExtended),
		Url:     formTestURL(common.ManageabilityTestKey, "container-port-name-format"),
		Version: VersionOne,
	}
)

func formDescription(identifier claim.Identifier, description string) string {
	return fmt.Sprintf("%s %s", identifier.Url, description)
}

// GetGinkgoTestIDAndLabels transform the claim.Identifier into a test Id that can be used to skip
// specific tests
func GetGinkgoTestIDAndLabels(identifier claim.Identifier) (testID string, tags []string) {
	suiteName, testName := GetSuiteAndTestFromIdentifier(identifier)
	testID = suiteName + "-" + testName
	tags = strings.Split(identifier.Tags, ",")
	tags = append(tags, suiteName, testID)

	TestIDToClaimID[testID] = identifier

	return testID, tags
}

// It extracts the suite name and test name from a claim.Identifier based
// on the const url which contains a base domain
// From a claim.Identifier.url:
//   http://test-network-function.com/tests-case/SuitName/TestName
// It extracts SuitNAme and TestName

func GetSuiteAndTestFromIdentifier(identifier claim.Identifier) (suiteName, testName string) {
	result := strings.Split(identifier.Url, url+"/")
	const SPLITN = 2
	// len 2, the baseDomain can appear only once in the url
	// so it returns what you have previous and before basedomain
	if len(result) != SPLITN {
		panic(fmt.Sprintf("Invalid Identifier URL: %s", identifier.Url))
	}

	result = strings.Split(result[1], "/")
	suiteName = result[0]
	testName = result[1]

	return
}

// Catalog is the JUnit testcase catalog of tests.
var Catalog = map[claim.Identifier]TestCaseDescription{

	TestNodeOperatingSystemIdentifier: {
		Identifier: TestNodeOperatingSystemIdentifier,
		Type:       NormativeResult,
		Description: formDescription(TestNodeOperatingSystemIdentifier, `Tests that the nodes running in the cluster have operating systems
			that are compatible with the deployed version of OpenShift.`),
		Remediation:           NodeOperatingSystemRemediation,
		ExceptionProcess:      NoDocumentedProcess,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 7.9",
		Tags:                  TestNodeOperatingSystemIdentifier.Tags,
	},

	TestOCPLifecycleIdentifier: {
		Identifier:            TestOCPLifecycleIdentifier,
		Type:                  NormativeResult,
		Description:           formDescription(TestOCPLifecycleIdentifier, `Tests that the running OCP version is not end of life.`),
		Remediation:           OCPLifecycleRemediation,
		ExceptionProcess:      NoDocumentedProcess,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 7.9",
		Tags:                  TestOCPLifecycleIdentifier.Tags,
	},

	TestDeploymentScalingIdentifier: {
		Identifier: TestDeploymentScalingIdentifier,
		Type:       NormativeResult,
		Description: formDescription(TestDeploymentScalingIdentifier,
			`Tests that CNF deployments support scale in/out operations.
			First, The test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the
			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s.
		    In case of deployments that are managed by HPA the test is changing the min and max value to deployment Replica - 1 during scale-in and the
			original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the deployment/s`),
		Remediation:           DeploymentScalingRemediation,
		ExceptionProcess:      NoDocumentedProcess,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		Tags:                  TestDeploymentScalingIdentifier.Tags,
	},
	TestStateFulSetScalingIdentifier: {
		Identifier: TestStateFulSetScalingIdentifier,
		Type:       NormativeResult,
		Description: formDescription(TestStateFulSetScalingIdentifier,
			`Tests that CNF statefulsets support scale in/out operations.
			First, The test starts getting the current replicaCount (N) of the statefulset/s with the Pod Under Test. Then, it executes the
			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the statefulset/s.
			In case of statefulsets that are managed by HPA the test is changing the min and max value to statefulset Replica - 1 during scale-in and the
			original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the statefulset/s`),
		Remediation:           StatefulSetScalingRemediation,
		ExceptionProcess:      NoDocumentedProcess,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		Tags:                  TestStateFulSetScalingIdentifier.Tags,
	},
	TestSecConCapabilitiesIdentifier: {
		Identifier:       TestSecConCapabilitiesIdentifier,
		Type:             NormativeResult,
		Remediation:      SecConCapabilitiesRemediation,
		ExceptionProcess: SecConCapabilitiesExceptionProcess,
		Description: formDescription(TestSecConCapabilitiesIdentifier,
			`Tests that the following capabilities are not granted:
			- NET_ADMIN
			- SYS_ADMIN
			- NET_RAW
			- IPC_LOCK
`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		Tags:                  TestSecConCapabilitiesIdentifier.Tags,
	},
	// TestPodDeleteIdentifier: {
	// 	Identifier:  TestPodDeleteIdentifier,
	// 	Type:        NormativeResult,
	// 	Remediation: `Make sure that the pods can be recreated successfully after deleting them`,
	// 	Description: formDescription(TestPodDeleteIdentifier,
	// 		`Using the litmus chaos operator, this test checks that pods are recreated successfully after deleting them.`),
	// 	BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
	// },
	TestSecConNonRootUserIdentifier: {
		Identifier:       TestSecConNonRootUserIdentifier,
		Type:             NormativeResult,
		Remediation:      SecConNonRootUserRemediation,
		ExceptionProcess: SecConNonRootUserExceptionProcess,
		Description: formDescription(TestSecConNonRootUserIdentifier,
			`Checks the security context runAsUser parameter in pods and containers to make sure it is not set to uid root(0)`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		Tags:                  TestSecConNonRootUserIdentifier.Tags,
	},
	TestSecContextIdentifier: {
		Identifier:       TestSecContextIdentifier,
		Type:             NormativeResult,
		Remediation:      SecConNonRootUserRemediation,
		ExceptionProcess: SecConNonRootUserExceptionProcess,
		Description: formDescription(TestSecContextIdentifier,
			`Checks the security context is match one of the 4 categories`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		Tags:                  TestSecContextIdentifier.Tags,
	},
	TestSecConPrivilegeEscalation: {
		Identifier:       TestSecConPrivilegeEscalation,
		Type:             NormativeResult,
		Remediation:      SecConPrivilegeRemediation,
		ExceptionProcess: NoDocumentedProcess,
		Description: formDescription(TestSecConPrivilegeEscalation,
			`Checks if privileged escalation is enabled (AllowPrivilegeEscalation=true)`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		Tags:                  TestSecConPrivilegeEscalation.Tags,
	},
	TestContainerIsCertifiedIdentifier: {
		Identifier:  TestContainerIsCertifiedIdentifier,
		Type:        NormativeResult,
		Remediation: ContainerIsCertifiedRemediation,
		Description: formDescription(TestContainerIsCertifiedIdentifier,
			`Tests whether container images listed in the configuration file have passed the Red Hat Container Certification Program (CCP).`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.3.7",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestContainerIsCertifiedDigestIdentifier.Tags,
	},
	TestContainerHostPort: {
		Identifier:  TestContainerHostPort,
		Type:        InformativeResult,
		Remediation: ContainerHostPortRemediation,
		Description: formDescription(TestContainerHostPort,
			`Verifies if containers define a hostPort.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.3.6",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestContainerHostPort.Tags,
	},
	TestPodHostNetwork: {
		Identifier:  TestPodHostNetwork,
		Type:        InformativeResult,
		Remediation: PodHostNetworkRemediation,
		Description: formDescription(TestPodHostNetwork,
			`Verifies that the spec.HostNetwork parameter is set to false`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.3.6",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodHostNetwork.Tags,
	},
	TestPodHostPath: {
		Identifier:  TestPodHostPath,
		Type:        InformativeResult,
		Remediation: PodHostNetworkRemediation,
		Description: formDescription(TestPodHostPath,
			`Verifies that the spec.HostPath parameter is not set (not present)`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.3.6",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodHostPath.Tags,
	},
	TestPodHostIPC: {
		Identifier:  TestPodHostIPC,
		Type:        InformativeResult,
		Remediation: PodHostIPCRemediation,
		Description: formDescription(TestPodHostIPC,
			`Verifies that the spec.HostIpc parameter is set to false`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.3.6",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodHostIPC.Tags,
	},
	TestPodHostPID: {
		Identifier:  TestPodHostPID,
		Type:        InformativeResult,
		Remediation: PodHostPIDRemediation,
		Description: formDescription(TestPodHostPID,
			`Verifies that the spec.HostPid parameter is set to false`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.3.6",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodHostIPC.Tags,
	},
	TestHugepagesNotManuallyManipulated: {
		Identifier:  TestHugepagesNotManuallyManipulated,
		Type:        NormativeResult,
		Remediation: HugepagesNotManuallyManipulatedRemediation,
		Description: formDescription(TestHugepagesNotManuallyManipulated,
			`Checks to see that HugePage settings have been configured through MachineConfig, and not manually on the
underlying Node.  This test case applies only to Nodes that are configured with the "worker" MachineConfigSet.  First,
the "worker" MachineConfig is polled, and the Hugepage settings are extracted.  Next, the underlying Nodes are polled
for configured HugePages through inspection of /proc/meminfo.  The results are compared, and the test passes only if
they are the same.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestHugepagesNotManuallyManipulated.Tags,
	},
	TestICMPv6ConnectivityIdentifier: {
		Identifier:  TestICMPv6ConnectivityIdentifier,
		Type:        NormativeResult,
		Remediation: ICMPv6ConnectivityRemediation,
		Description: formDescription(TestICMPv6ConnectivityIdentifier,
			`Checks that each CNF Container is able to communicate via ICMPv6 on the Default OpenShift network.  This
test case requires the Deployment of the debug daemonset.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestICMPv6ConnectivityIdentifier.Tags,
	},

	TestICMPv4ConnectivityMultusIdentifier: {
		Identifier:  TestICMPv4ConnectivityMultusIdentifier,
		Type:        NormativeResult,
		Remediation: ICMPv4ConnectivityMultusRemediation,
		Description: formDescription(TestICMPv4ConnectivityMultusIdentifier,
			`Checks that each CNF Container is able to communicate via ICMPv4 on the Multus network(s).  This
test case requires the Deployment of the debug daemonset.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestICMPv4ConnectivityIdentifier.Tags,
	},

	TestICMPv6ConnectivityMultusIdentifier: {
		Identifier:  TestICMPv6ConnectivityMultusIdentifier,
		Type:        NormativeResult,
		Remediation: ICMPv6ConnectivityMultusRemediation,
		Description: formDescription(TestICMPv6ConnectivityMultusIdentifier,
			`Checks that each CNF Container is able to communicate via ICMPv6 on the Multus network(s).  This
test case requires the Deployment of the debug daemonset.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestICMPv6ConnectivityIdentifier.Tags,
	},

	TestServiceDualStackIdentifier: {
		Identifier:  TestServiceDualStackIdentifier,
		Type:        NormativeResult,
		Remediation: TestServiceDualStackRemediation,
		Description: formDescription(TestServiceDualStackIdentifier,
			`Checks that all services in namespaces under test are either ipv6 single stack or dual stack. This
test case requires the deployment of the debug daemonset.`),
		BestPracticeReference: bestPracticeDocV1dot4URL + " Section 3.5.7",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestServiceDualStackIdentifier.Tags,
	},

	TestNFTablesIdentifier: {
		Identifier:  TestNFTablesIdentifier,
		Type:        NormativeResult,
		Remediation: TestNFTablesRemediation,
		Description: formDescription(TestNFTablesIdentifier,
			`Checks that the output of "nft list ruleset" is empty, e.g. there is no nftables configuration on any CNF containers.`),
		BestPracticeReference: bestPracticeDocV1dot4URL + " Section 4.6.23",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestNFTablesIdentifier.Tags,
	},

	TestIPTablesIdentifier: {
		Identifier:  TestIPTablesIdentifier,
		Type:        NormativeResult,
		Remediation: TestIPTablesRemediation,
		Description: formDescription(TestIPTablesIdentifier,
			`Checks that the output of "iptables-save" is empty, e.g. there is no iptables configuration on any CNF containers.`),
		BestPracticeReference: bestPracticeDocV1dot4URL + " Section 4.6.23",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestIPTablesIdentifier.Tags,
	},

	TestNamespaceBestPracticesIdentifier: {
		Identifier:  TestNamespaceBestPracticesIdentifier,
		Type:        NormativeResult,
		Remediation: NamespaceBestPracticesRemediation,
		Description: formDescription(TestNamespaceBestPracticesIdentifier,
			`Tests that all CNF's resources (PUTs and CRs) belong to valid namespaces. A valid namespace meets
the following conditions: (1) It was declared in the yaml config file under the targetNameSpaces
tag. (2) It doesn't have any of the following prefixes: default, openshift-, istio- and aspenmesh-`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2, 16.3.8 and 16.3.9",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestNamespaceBestPracticesIdentifier.Tags,
	},

	TestNonTaintedNodeKernelsIdentifier: {
		Identifier:  TestNonTaintedNodeKernelsIdentifier,
		Type:        NormativeResult,
		Remediation: NonTaintedNodeKernelsRemediation,
		Description: formDescription(TestNonTaintedNodeKernelsIdentifier,
			`Ensures that the Node(s) hosting CNFs do not utilize tainted kernels. This test case is especially important
to support Highly Available CNFs, since when a CNF is re-instantiated on a backup Node, that Node's kernel may not have
the same hacks.'`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.14",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestNonTaintedNodeKernelsIdentifier.Tags,
	},

	TestOperatorInstallStatusSucceededIdentifier: {
		Identifier:  TestOperatorInstallStatusSucceededIdentifier,
		Type:        NormativeResult,
		Remediation: OperatorInstallStatusSucceededRemediation,
		Description: formDescription(TestOperatorInstallStatusSucceededIdentifier,
			`Ensures that the target CNF operators report "Succeeded" as their installation status.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.12 and 5.3.3",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestOperatorInstallStatusSucceededIdentifier.Tags,
	},

	TestOperatorNoPrivileges: {
		Identifier:  TestOperatorNoPrivileges,
		Type:        NormativeResult,
		Remediation: OperatorNoPrivilegesRemediation,
		Description: formDescription(TestOperatorNoPrivileges,
			`The operator is not installed with privileged rights. Test passes if clusterPermissions is not present in the CSV manifest or is present
with no resourceNames under its rules.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.12 and 5.3.3",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestOperatorNoPrivileges.Tags,
	},

	TestOperatorIsCertifiedIdentifier: {
		Identifier:  TestOperatorIsCertifiedIdentifier,
		Type:        NormativeResult,
		Remediation: OperatorIsCertifiedRemediation,
		Description: formDescription(TestOperatorIsCertifiedIdentifier,
			`Tests whether CNF Operators listed in the configuration file have passed the Red Hat Operator Certification Program (OCP).`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.12 and 5.3.3",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestOperatorIsCertifiedIdentifier.Tags,
	},

	TestHelmIsCertifiedIdentifier: {
		Identifier:  TestHelmIsCertifiedIdentifier,
		Type:        NormativeResult,
		Remediation: HelmIsCertifiedRemediation,
		Description: formDescription(TestHelmIsCertifiedIdentifier,
			`Tests whether helm charts listed in the cluster passed the Red Hat Helm Certification Program.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.12 and 5.3.3",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestHelmIsCertifiedIdentifier.Tags,
	},

	TestOperatorIsInstalledViaOLMIdentifier: {
		Identifier:  TestOperatorIsInstalledViaOLMIdentifier,
		Type:        NormativeResult,
		Remediation: OperatorIsInstalledViaOLMRemediation,
		Description: formDescription(TestOperatorIsInstalledViaOLMIdentifier,
			`Tests whether a CNF Operator is installed via OLM.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.12 and 5.3.3",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestOperatorIsInstalledViaOLMIdentifier.Tags,
	},

	TestPodNodeSelectorAndAffinityBestPractices: {
		Identifier:  TestPodNodeSelectorAndAffinityBestPractices,
		Type:        InformativeResult,
		Remediation: PodNodeSelectorAndAffinityBestPracticesRemediation,
		Description: formDescription(TestPodNodeSelectorAndAffinityBestPractices,
			`Ensures that CNF Pods do not specify nodeSelector or nodeAffinity.  In most cases, Pods should allow for
instantiation on any underlying Node.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodNodeSelectorAndAffinityBestPractices.Tags,
	},

	TestPodHighAvailabilityBestPractices: {
		Identifier:  TestPodHighAvailabilityBestPractices,
		Type:        InformativeResult,
		Remediation: PodHighAvailabilityBestPracticesRemediation,
		Description: formDescription(TestPodHighAvailabilityBestPractices,
			`Ensures that CNF Pods specify podAntiAffinity rules and replica value is set to more than 1.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodHighAvailabilityBestPractices.Tags,
	},

	TestPodClusterRoleBindingsBestPracticesIdentifier: {
		Identifier:  TestPodClusterRoleBindingsBestPracticesIdentifier,
		Type:        NormativeResult,
		Remediation: PodClusterRoleBindingsBestPracticesRemediation,
		Description: formDescription(TestPodClusterRoleBindingsBestPracticesIdentifier,
			`Tests that a Pod does not specify ClusterRoleBindings.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.10 and 5.3.6",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodClusterRoleBindingsBestPracticesIdentifier.Tags,
	},

	TestPodDeploymentBestPracticesIdentifier: {
		Identifier:  TestPodDeploymentBestPracticesIdentifier,
		Type:        NormativeResult,
		Remediation: PodDeploymentBestPracticesRemediation,
		Description: formDescription(TestPodDeploymentBestPracticesIdentifier,
			`Tests that CNF Pod(s) are deployed as part of a ReplicaSet(s)/StatefulSet(s).`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.3.3 and 5.3.8",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodDeploymentBestPracticesIdentifier.Tags,
	},
	TestImagePullPolicyIdentifier: {
		Identifier:  TestImagePullPolicyIdentifier,
		Type:        NormativeResult,
		Remediation: ImagePullPolicyRemediation,
		Description: formDescription(TestImagePullPolicyIdentifier,
			`Ensure that the containers under test are using IfNotPresent as Image Pull Policy..`),
		BestPracticeReference: bestPracticeDocV1dot3URL + "  Section 12.6",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestImagePullPolicyIdentifier.Tags,
	},

	TestPodRoleBindingsBestPracticesIdentifier: {
		Identifier:  TestPodRoleBindingsBestPracticesIdentifier,
		Type:        NormativeResult,
		Remediation: PodRoleBindingsBestPracticesRemediation,
		Description: formDescription(TestPodRoleBindingsBestPracticesIdentifier,
			`Ensures that a CNF does not utilize RoleBinding(s) in a non-CNF Namespace.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.3.3 and 5.3.5",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodRoleBindingsBestPracticesIdentifier.Tags,
	},

	TestPodServiceAccountBestPracticesIdentifier: {
		Identifier:  TestPodServiceAccountBestPracticesIdentifier,
		Type:        NormativeResult,
		Remediation: PodServiceAccountBestPracticesRemediation,
		Description: formDescription(TestPodServiceAccountBestPracticesIdentifier,
			`Tests that each CNF Pod utilizes a valid Service Account.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.3 and 5.2.7",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodServiceAccountBestPracticesIdentifier.Tags,
	},

	TestServicesDoNotUseNodeportsIdentifier: {
		Identifier:  TestServicesDoNotUseNodeportsIdentifier,
		Type:        NormativeResult,
		Remediation: ServicesDoNotUseNodeportsRemediation,
		Description: formDescription(TestServicesDoNotUseNodeportsIdentifier,
			`Tests that each CNF Service does not utilize NodePort(s).`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.3.1",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestServicesDoNotUseNodeportsIdentifier.Tags,
	},

	TestUnalteredBaseImageIdentifier: {
		Identifier:  TestUnalteredBaseImageIdentifier,
		Type:        NormativeResult,
		Remediation: UnalteredBaseImageRemediation,
		Description: formDescription(TestUnalteredBaseImageIdentifier,
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
10) /usr/lib64`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.1.4",
		ExceptionProcess:      UnalteredBaseImageExceptionProcess,
		Tags:                  TestUnalteredBaseImageIdentifier.Tags,
	},

	TestUnalteredStartupBootParamsIdentifier: {
		Identifier:  TestUnalteredStartupBootParamsIdentifier,
		Type:        NormativeResult,
		Remediation: UnalteredStartupBootParamsRemediation,
		Description: formDescription(TestUnalteredStartupBootParamsIdentifier,
			`Tests that boot parameters are set through the MachineConfigOperator, and not set manually on the Node.`),
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.13 and 5.2.14",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestUnalteredStartupBootParamsIdentifier.Tags,
	},
	TestPodRecreationIdentifier: {
		Identifier: TestPodRecreationIdentifier,
		Type:       NormativeResult,
		Description: formDescription(TestPodRecreationIdentifier,
			`Tests that a CNF is configured to support High Availability.
			First, this test cordons and drains a Node that hosts the CNF Pod.
			Next, the test ensures that OpenShift can re-instantiate the Pod on another Node,
			and that the actual replica count matches the desired replica count.`),
		Remediation:           PodRecreationRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodRecreationIdentifier.Tags,
	},
	TestSysctlConfigsIdentifier: {
		Identifier: TestSysctlConfigsIdentifier,
		Type:       NormativeResult,
		Description: formDescription(TestPodRecreationIdentifier,
			`Tests that no one has changed the node's sysctl configs after the node
			was created, the tests works by checking if the sysctl configs are consistent with the
			MachineConfig CR which defines how the node should be configured`),
		Remediation:           SysctlConfigsRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestSysctlConfigsIdentifier.Tags,
	},
	TestServiceMeshIdentifier: {
		Identifier: TestServiceMeshIdentifier,
		Type:       NormativeResult,
		Description: formDescription(TestServiceMeshIdentifier,
			`verifies whether, if available, service mesh is actually being used by the CNF pods`),
		Remediation:           ServiceMeshRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestServiceMeshIdentifier.Tags,
	},
	TestScalingIdentifier: {
		Identifier: TestScalingIdentifier,
		Type:       NormativeResult,
		Description: formDescription(TestScalingIdentifier,
			`Tests that CNF deployments support scale in/out operations.
			First, The test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the
			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s.`),
		Remediation:           ScalingRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestScalingIdentifier.Tags,
	},
	TestIsRedHatReleaseIdentifier: {
		Identifier: TestIsRedHatReleaseIdentifier,
		Type:       NormativeResult,
		Description: formDescription(TestIsRedHatReleaseIdentifier,
			`verifies if the container base image is redhat.`),
		Remediation:           IsRedHatReleaseRemediation,
		ExceptionProcess:      IsRedHatReleaseExceptionProcess,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		Tags:                  TestIsRedHatReleaseIdentifier.Tags,
	},
	TestIsSELinuxEnforcingIdentifier: {
		Identifier: TestIsSELinuxEnforcingIdentifier,
		Type:       NormativeResult,
		Description: formDescription(TestIsSELinuxEnforcingIdentifier,
			`verifies that all openshift platform/cluster nodes have selinux in "Enforcing" mode.`),
		Remediation:           IsSELinuxEnforcingRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 10.3 Pod Security",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestIsSELinuxEnforcingIdentifier.Tags,
	},
	TestUndeclaredContainerPortsUsage: {
		Identifier: TestUndeclaredContainerPortsUsage,
		Type:       NormativeResult,
		Description: formDescription(TestUndeclaredContainerPortsUsage,
			`Check that containers do not listen on ports that weren't declared in their specification`),
		Remediation:           UndeclaredContainerPortsRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 16.3.1.1",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestUndeclaredContainerPortsUsage.Tags,
	},
	TestOCPReservedPortsUsage: {
		Identifier: TestOCPReservedPortsUsage,
		Type:       NormativeResult,
		Description: formDescription(TestOCPReservedPortsUsage,
			`Check that containers do not listen on ports that are reserved by OpenShift`),
		Remediation:           OCPReservedPortsUsageRemediation,
		BestPracticeReference: bestPracticeDocV1dot4URL + " Section 3.5.9",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestOCPReservedPortsUsage.Tags,
	},
	TestCrdsStatusSubresourceIdentifier: {
		Identifier: TestCrdsStatusSubresourceIdentifier,
		Type:       InformativeResult,
		Description: formDescription(TestCrdsStatusSubresourceIdentifier,
			`Checks that all CRDs have a status subresource specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties["status"]).`),
		Remediation:           CrdsStatusSubresourceRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestCrdsStatusSubresourceIdentifier.Tags,
	},
	TestLoggingIdentifier: {
		Identifier: TestLoggingIdentifier,
		Type:       InformativeResult,
		Description: formDescription(TestLoggingIdentifier,
			`Check that all containers under test use standard input output and standard error when logging`),
		Remediation:           LoggingRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 10.1",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestLoggingIdentifier.Tags,
	},
	TestTerminationMessagePolicyIdentifier: {
		Identifier: TestTerminationMessagePolicyIdentifier,
		Type:       InformativeResult,
		Description: formDescription(TestTerminationMessagePolicyIdentifier,
			`Check that all containers are using terminationMessagePolicy: FallbackToLogsOnError`),
		Remediation:           TerminationMessagePolicyRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 12.1",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestTerminationMessagePolicyIdentifier.Tags,
	},
	TestPodAutomountServiceAccountIdentifier: {
		Identifier: TestPodAutomountServiceAccountIdentifier,
		Type:       NormativeResult,
		Description: formDescription(TestPodAutomountServiceAccountIdentifier,
			`Check that all pods under test have automountServiceAccountToken set to false`),
		Remediation:           AutomountServiceTokenRemediation,
		ExceptionProcess:      AutomountServiceTokenExceptionProcess,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 12.7",
		Tags:                  TestPodAutomountServiceAccountIdentifier.Tags,
	},
	TestLivenessProbeIdentifier: {
		Identifier:            TestLivenessProbeIdentifier,
		Type:                  NormativeResult,
		Description:           formDescription(TestLivenessProbeIdentifier, `Check that all containers under test a have liveness probe defined`),
		Remediation:           LivenessProbeRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.16, 12.1 and 12.5",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestLivenessProbeIdentifier.Tags,
	},
	TestReadinessProbeIdentifier: {
		Identifier:            TestReadinessProbeIdentifier,
		Type:                  NormativeResult,
		Description:           formDescription(TestReadinessProbeIdentifier, `Check that all containers under test a have readiness probe defined`),
		Remediation:           ReadinessProbeRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 5.2.16, 12.1 and 12.5",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestReadinessProbeIdentifier.Tags,
	},
	TestStartupProbeIdentifier: {
		Identifier:            TestStartupProbeIdentifier,
		Type:                  NormativeResult,
		Description:           formDescription(TestStartupProbeIdentifier, `Check that all containers under test a have startup probe defined`),
		Remediation:           StartupProbeRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 4.6.12", // TODO Change this to v1.4 when available
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestStartupProbeIdentifier.Tags,
	},
	TestOneProcessPerContainerIdentifier: {
		Identifier:            TestOneProcessPerContainerIdentifier,
		Type:                  InformativeResult,
		Description:           formDescription(TestOneProcessPerContainerIdentifier, `Check that all containers under test have only one process running`),
		Remediation:           OneProcessPerContainerRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 10.8.3",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestOneProcessPerContainerIdentifier.Tags,
	},
	TestSYSNiceRealtimeCapabilityIdentifier: {
		Identifier:            TestSYSNiceRealtimeCapabilityIdentifier,
		Type:                  InformativeResult,
		Description:           formDescription(TestSYSNiceRealtimeCapabilityIdentifier, `Check that pods running on nodes with realtime kernel enabled have the SYS_NICE capability enabled in their spec.`),
		Remediation:           SYSNiceRealtimeCapabilityRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 2.7.4",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestSYSNiceRealtimeCapabilityIdentifier.Tags,
	},
	TestSysPtraceCapabilityIdentifier: {
		Identifier:            TestSysPtraceCapabilityIdentifier,
		Type:                  InformativeResult,
		Description:           formDescription(TestSysPtraceCapabilityIdentifier, `Check that if process namespace sharing is enabled for a Pod then the SYS_PTRACE capability is allowed`),
		Remediation:           SysPtraceCapabilityRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 2.7.5",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestSysPtraceCapabilityIdentifier.Tags,
	},
	TestPodRequestsAndLimitsIdentifier: {
		Identifier:            TestPodRequestsAndLimitsIdentifier,
		Type:                  InformativeResult,
		Description:           formDescription(TestPodRequestsAndLimitsIdentifier, `Check that containers have resource requests and limits specified in their spec.`),
		Remediation:           RequestsAndLimitsRemediation,
		BestPracticeReference: bestPracticeDocV1dot4URL + " Section 4.6.11",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodRequestsAndLimitsIdentifier.Tags,
	},
	TestPersistentVolumeReclaimPolicyIdentifier: {
		Identifier:            TestPersistentVolumeReclaimPolicyIdentifier,
		Type:                  InformativeResult,
		Description:           formDescription(TestPersistentVolumeReclaimPolicyIdentifier, `Check that the persistent volumes the CNF pods are using have a reclaim policy of delete.`),
		Remediation:           PersistentVolumeReclaimPolicyRemediation,
		BestPracticeReference: bestPracticeDocV1dot4URL + " Section 3.3.4",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPersistentVolumeReclaimPolicyIdentifier.Tags,
	},
	TestContainersImageTag: {
		Identifier:            TestContainersImageTag,
		Type:                  InformativeResult,
		Description:           formDescription(TestContainersImageTag, `Check that image tag exists on containers.`),
		Remediation:           ContainersImageTag,
		BestPracticeReference: bestPracticeDocV1dot4URL + " Section 4.6.12",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestContainersImageTag.Tags,
	},
	TestNamespaceResourceQuotaIdentifier: {
		Identifier:            TestNamespaceResourceQuotaIdentifier,
		Type:                  InformativeResult,
		Description:           formDescription(TestNamespaceResourceQuotaIdentifier, `Checks to see if CNF workload pods are running in namespaces that have resource quotas applied.`),
		Remediation:           NamespaceResourceQuotaRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 4.6.8", // TODO Change this to v1.4 when available
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestNamespaceResourceQuotaIdentifier.Tags,
	},
	TestPodDisruptionBudgetIdentifier: {
		Identifier:            TestPodDisruptionBudgetIdentifier,
		Type:                  NormativeResult,
		Description:           formDescription(TestPodDisruptionBudgetIdentifier, `Checks to see if pod disruption budgets have allowed values for minAvailable and maxUnavailable`),
		Remediation:           PodDisruptionBudgetRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 4.6.20", // TODO Change this to v1.4 when available
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodDisruptionBudgetIdentifier.Tags,
	},
	TestPodTolerationBypassIdentifier: {
		Identifier:            TestPodTolerationBypassIdentifier,
		Type:                  InformativeResult,
		Description:           formDescription(TestPodTolerationBypassIdentifier, `Check that pods do not have NoExecute, PreferNoSchedule, or NoSchedule tolerations that have been modified from the default.`),
		Remediation:           PodTolerationBypassRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 10.6",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestPodTolerationBypassIdentifier.Tags,
	},
	TestNoSSHDaemonsAllowedIdentifier: {
		Identifier:            TestNoSSHDaemonsAllowedIdentifier,
		Type:                  NormativeResult,
		Description:           formDescription(TestNoSSHDaemonsAllowedIdentifier, `Check that pods do not run SSH daemons.`),
		Remediation:           NoSSHDaemonsAllowedRemediation,
		BestPracticeReference: bestPracticeDocV1dot3URL + " Section 4.6.12", // TODO Change this to v1.4 when available
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestNoSSHDaemonsAllowedIdentifier.Tags,
	},
	TestCPUIsolationIdentifier: {
		Identifier: TestCPUIsolationIdentifier,
		Type:       InformativeResult,
		Description: formDescription(TestCPUIsolationIdentifier, `CPU isolation requires: For each container within the pod, resource requests and limits must be identical.
		Request and Limits are in the form of whole CPUs. The runTimeClassName must be specified. Annotations required disabling CPU and IRQ load-balancing.`),
		Remediation:           CPUIsolationRemediation,
		BestPracticeReference: bestPracticeDocV1dot4URL + " Section 3.5.5",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestCPUIsolationIdentifier.Tags,
	},
	TestContainerPortNameFormat: {
		Identifier:            TestContainerPortNameFormat,
		Type:                  NormativeResult,
		Description:           formDescription(TestContainerPortNameFormat, `Check that the container's ports name follow the naming conventions.`),
		Remediation:           ContainerPortNameFormatRemediation,
		BestPracticeReference: bestPracticeDocV1dot4URL + " Section 4.6.20",
		ExceptionProcess:      NoDocumentedProcess,
		Tags:                  TestContainerPortNameFormat.Tags,
	},
}
