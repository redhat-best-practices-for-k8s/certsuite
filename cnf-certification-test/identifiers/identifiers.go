// Copyright (C) 2021 Red Hat, Inc.
// Copyright (C) 2021 Red Hat, Inc.
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
	bestPracticeDocV1dot2URL = "[CNF Best Practice V1.2](https://connect.redhat.com/sites/default/files/2021-03/Cloud%20Native%20Network%20Function%20Requirements.pdf)"
	informativeResult        = "informative"
	normativeResult          = "normative"
	url                      = "http://test-network-function.com/testcases"
	versionOne               = "v1.0.0"
	bestPracticeDocV1dot3URL = "https://docs.google.com/document/d/1wRHMk1ZYUSVmgp_4kxvqjVOKwolsZ5hDXjr5MLy-wbg/edit#"
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
}

func formTestURL(suite, name string) string {
	return fmt.Sprintf("%s/%s/%s", url, suite, name)
}

var (
	// TestIdToClaimId converts the testcase short ID to the claim identifier
	TestIDToClaimID = map[string]claim.Identifier{}

	// TestPodDeleteIdentifier tests for delete pod test
	TestPodDeleteIdentifier = claim.Identifier{
		Url:     formTestURL(common.ChaosTesting, "pod-delete"),
		Version: versionOne,
	}
	// TestSecConCapabilitiesIdentifier tests for non compliant security context capabilities
	TestSecConCapabilitiesIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "security-context-capabilities-check"),
		Version: versionOne,
	}
	// TestSecConNonRootUserIdentifier tests that pods or containers are not running with root permissions
	TestSecConNonRootUserIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "security-context-non-root-user-check"),
		Version: versionOne,
	}
	// TestSecPrivilegedEscalation tests that containers are not allowed privilege escalation
	TestSecConPrivilegeEscalation = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "security-context-privilege-escalation"),
		Version: versionOne,
	}
	// TestSecPrivilegedEscalation tests that containers are not configured with host port privileges
	TestContainerHostPort = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "container-host-port"),
		Version: versionOne,
	}
	// TestPodHostNetwork tests that pods do not configure hostnetwork to true
	TestPodHostNetwork = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "pod-host-network"),
		Version: versionOne,
	}
	// TestPodHostPath tests that pods do not configure an hostpath volume
	TestPodHostPath = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "pod-host-path"),
		Version: versionOne,
	}
	// TestPodHostPath tests that pods do not configure an hostpath volume
	TestPodHostIPC = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "pod-host-ipc"),
		Version: versionOne,
	}
	// TestPodHostPath tests that pods do not configure an hostpath volume
	TestPodHostPID = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "pod-host-pid"),
		Version: versionOne,
	}
	// TestContainerIsCertifiedIdentifier tests whether the container has passed Container Certification.
	TestContainerIsCertifiedIdentifier = claim.Identifier{
		Url:     formTestURL(common.AffiliatedCertTestKey, "container-is-certified"),
		Version: versionOne,
	}
	// TestHugepagesNotManuallyManipulated represents the test identifier testing hugepages have not been manipulated.
	TestHugepagesNotManuallyManipulated = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "hugepages-config"),
		Version: versionOne,
	}
	// TestICMPv4ConnectivityIdentifier tests icmpv4 connectivity.
	TestICMPv4ConnectivityIdentifier = claim.Identifier{
		Url:     formTestURL(common.NetworkingTestKey, "icmpv4-connectivity"),
		Version: versionOne,
	}
	// TestICMPv6ConnectivityIdentifier tests icmpv6 connectivity.
	TestICMPv6ConnectivityIdentifier = claim.Identifier{
		Url:     formTestURL(common.NetworkingTestKey, "icmpv6-connectivity"),
		Version: versionOne,
	}
	// TestICMPv4ConnectivityIdentifier tests icmpv4 Multus connectivity.
	TestICMPv4ConnectivityMultusIdentifier = claim.Identifier{
		Url:     formTestURL(common.NetworkingTestKey, "icmpv4-connectivity-multus"),
		Version: versionOne,
	}
	// TestICMPv6ConnectivityIdentifier tests icmpv6 Multus connectivity.
	TestICMPv6ConnectivityMultusIdentifier = claim.Identifier{
		Url:     formTestURL(common.NetworkingTestKey, "icmpv6-connectivity-multus"),
		Version: versionOne,
	}
	// TestNamespaceBestPracticesIdentifier ensures the namespace has followed best namespace practices.
	TestNamespaceBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "namespace"),
		Version: versionOne,
	}
	// TestNonTaintedNodeKernelsIdentifier is the identifier for the test checking tainted nodes.
	TestNonTaintedNodeKernelsIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "tainted-node-kernel"),
		Version: versionOne,
	}
	// TestOperatorInstallStatusSucceededIdentifier tests Operator best practices.
	TestOperatorInstallStatusSucceededIdentifier = claim.Identifier{
		Url:     formTestURL(common.OperatorTestKey, "install-status-succeeded"),
		Version: versionOne,
	}
	// TestOperatorNoPrivileges tests Operator has no privileges on resources.
	TestOperatorNoPrivileges = claim.Identifier{
		Url:     formTestURL(common.OperatorTestKey, "install-status-no-privileges"),
		Version: versionOne,
	}
	// TestOperatorIsCertifiedIdentifier tests that an Operator has passed Operator certification.
	TestOperatorIsCertifiedIdentifier = claim.Identifier{
		Url:     formTestURL(common.AffiliatedCertTestKey, "operator-is-certified"),
		Version: versionOne,
	}
	TestHelmIsCertifiedIdentifier = claim.Identifier{
		Url:     formTestURL(common.AffiliatedCertTestKey, "helmchart-is-certified"),
		Version: versionOne,
	}
	// TestOperatorIsInstalledViaOLMIdentifier tests that an Operator is installed via OLM.
	TestOperatorIsInstalledViaOLMIdentifier = claim.Identifier{
		Url:     formTestURL(common.OperatorTestKey, "install-source"),
		Version: versionOne,
	}
	// TestPodNodeSelectorAndAffinityBestPractices is the test ensuring nodeSelector and nodeAffinity are not used by a
	// Pod.
	TestPodNodeSelectorAndAffinityBestPractices = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "pod-scheduling"),
		Version: versionOne,
	}
	// TestPodHighAvailabilityBestPractices is the test ensuring podAntiAffinity are used by a
	// Pod when pod replica # are great than 1
	TestPodHighAvailabilityBestPractices = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "pod-high-availability"),
		Version: versionOne,
	}

	// TestPodClusterRoleBindingsBestPracticesIdentifier ensures Pod crb best practices.
	TestPodClusterRoleBindingsBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "cluster-role-bindings"),
		Version: versionOne,
	}
	// TestPodDeploymentBestPracticesIdentifier ensures a CNF follows best Deployment practices.
	TestPodDeploymentBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "pod-owner-type"),
		Version: versionOne,
	}
	// TestDeploymentScalingIdentifier ensures deployment scale in/out operations work correctly.
	TestDeploymentScalingIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "deployment-scaling"),
		Version: versionOne,
	}
	// TestStateFulSetScalingIdentifier ensures statefulset scale in/out operations work correctly.
	TestStateFulSetScalingIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "statefulset-scaling"),
		Version: versionOne,
	}
	// TestImagePullPolicyIdentifier ensures represent image pull policy practices.
	TestImagePullPolicyIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "image-pull-policy"),
		Version: versionOne,
	}
	// TestPodRecreationIdentifier ensures recreation best practices.
	TestPodRecreationIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "pod-recreation"),
		Version: versionOne,
	}
	// TestPodRoleBindingsBestPracticesIdentifier represents rb best practices.
	TestPodRoleBindingsBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "pod-role-bindings"),
		Version: versionOne,
	}
	// TestPodServiceAccountBestPracticesIdentifier tests Pod SA best practices.
	TestPodServiceAccountBestPracticesIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "pod-service-account"),
		Version: versionOne,
	}
	//
	TestPodAutomountServiceAccountIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "pod-automount-service-account-token"),
		Version: versionOne,
	}
	// TestServicesDoNotUseNodeportsIdentifier ensures Services don't utilize NodePorts.
	TestServicesDoNotUseNodeportsIdentifier = claim.Identifier{
		Url:     formTestURL(common.NetworkingTestKey, "service-type"),
		Version: versionOne,
	}
	// TestUnalteredBaseImageIdentifier ensures the base image is not altered.
	TestUnalteredBaseImageIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "base-image"),
		Version: versionOne,
	}
	// TestUnalteredStartupBootParamsIdentifier ensures startup boot params are not altered.
	TestUnalteredStartupBootParamsIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "boot-params"),
		Version: versionOne,
	}
	// TestLoggingIdentifier ensures stderr/stdout are used
	TestLoggingIdentifier = claim.Identifier{
		Url:     formTestURL(common.ObservabilityTestKey, "container-logging"),
		Version: versionOne,
	}
	// TestTerminationMessagePolicy ensures pods have FallbackToLogsOnError set
	TestTerminationMessagePolicyIdentifier = claim.Identifier{
		Url:     formTestURL(common.ObservabilityTestKey, "termination-policy"),
		Version: versionOne,
	}
	// TestCrdsStatusSubresourceIdentifier ensures all CRDs have a valid status subresource
	TestCrdsStatusSubresourceIdentifier = claim.Identifier{
		Url:     formTestURL(common.ObservabilityTestKey, "crd-status"),
		Version: versionOne,
	}
	// TestShudtownIdentifier ensures pre-stop lifecycle is defined
	TestShudtownIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "container-shutdown"),
		Version: versionOne,
	}
	// TestSysctlConfigsIdentifier ensures that the node's sysctl configs are consistent with the MachineConfig CR
	TestSysctlConfigsIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "sysctl-config"),
		Version: versionOne,
	}
	// TestScalingIdentifier ensures deployment scale in/out operations work correctly.
	TestScalingIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "scaling"),
		Version: versionOne,
	}
	// TestIsRedHatReleaseIdentifier ensures platform is defined
	TestIsRedHatReleaseIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "isredhat-release"),
		Version: versionOne,
	}
	// TestIsSELinuxEnforcingIdentifier ensures selinux is in enforcing mode
	TestIsSELinuxEnforcingIdentifier = claim.Identifier{
		Url:     formTestURL(common.PlatformAlterationTestKey, "is-selinux-enforcing"),
		Version: versionOne,
	}
	TestUndeclaredContainerPortsUsage = claim.Identifier{
		Url:     formTestURL(common.NetworkingTestKey, "undeclared-container-ports-usage"),
		Version: versionOne,
	}
	TestLivenessProbeIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "liveness"),
		Version: versionOne,
	}
	TestReadinessProbeIdentifier = claim.Identifier{
		Url:     formTestURL(common.LifecycleTestKey, "readiness"),
		Version: versionOne,
	}
	// TestOneProcessPerContainerIdentifier ensures that only one process per container is running
	TestOneProcessPerContainerIdentifier = claim.Identifier{
		Url:     formTestURL(common.AccessControlTestKey, "one-process-per-container"),
		Version: versionOne,
	}
)

func formDescription(identifier claim.Identifier, description string) string {
	return fmt.Sprintf("%s %s", identifier.Url, description)
}

// XformToGinkgoItIdentifier transform the claim.Identifier into a test Id that can be used to skip
// specific tests
func XformToGinkgoItIdentifier(identifier claim.Identifier) string {
	return XformToGinkgoItIdentifierExtended(identifier, "")
}

// XformToGinkgoItIdentifierExtended transform the claim.Identifier into a test Id that can be used to skip
// specific tests
func XformToGinkgoItIdentifierExtended(identifier claim.Identifier, extra string) string {
	itID := strings.ReplaceAll(strings.TrimPrefix(identifier.Url, url+"/"), "/", "-")
	var key string
	if extra != "" {
		key = itID + "-" + extra
	} else {
		key = itID
	}
	TestIDToClaimID[key] = identifier
	return key
}

// Catalog is the JUnit testcase catalog of tests.
var Catalog = map[claim.Identifier]TestCaseDescription{

	TestDeploymentScalingIdentifier: {
		Identifier: TestDeploymentScalingIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestDeploymentScalingIdentifier,
			`tests that CNF deployments support scale in/out operations. 
			First, The test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the 
			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s.
		    In case of deployments that are managed by HPA the test is changing the min and max value to deployment Replica - 1 during scale-in and the 
			original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the deployment/s`),
		Remediation:           `Make sure CNF deployments/replica sets can scale in/out successfully.`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestStateFulSetScalingIdentifier: {
		Identifier: TestStateFulSetScalingIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestStateFulSetScalingIdentifier,
			`tests that CNF statefulsets support scale in/out operations. 
			First, The test starts getting the current replicaCount (N) of the statefulset/s with the Pod Under Test. Then, it executes the 
			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the statefulset/s.
			In case of statefulsets that are managed by HPA the test is changing the min and max value to statefulset Replica - 1 during scale-in and the 
			original replicaCount again for both min/max during the scale-out stage. lastly its restoring the original min/max replica of the statefulset/s`),
		Remediation:           `Make sure CNF statefulsets/replica sets can scale in/out successfully.`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestSecConCapabilitiesIdentifier: {
		Identifier:  TestSecConCapabilitiesIdentifier,
		Type:        normativeResult,
		Remediation: `Remove the following capabilities from the container/pod definitions: NET_ADMIN SCC, SYS_ADMIN SCC, NET_RAW SCC, IPC_LOCK SCC `,
		Description: formDescription(TestSecConCapabilitiesIdentifier,
			`Tests that the following capabilities are not granted:
			- NET_ADMIN
			- SYS_ADMIN 
			- NET_RAW
			- IPC_LOCK
`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestPodDeleteIdentifier: {
		Identifier:  TestPodDeleteIdentifier,
		Type:        normativeResult,
		Remediation: `Make sure that the pods can be recreated succesfully after deleting them`,
		Description: formDescription(TestPodDeleteIdentifier,
			`Using the litmus chaos operator, this test checks that pods are recreated successfully after deleting them.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestSecConNonRootUserIdentifier: {
		Identifier:  TestSecConNonRootUserIdentifier,
		Type:        normativeResult,
		Remediation: `Change the pod and containers "runAsUser" uid to something other than root(0)`,
		Description: formDescription(TestSecConNonRootUserIdentifier,
			`Checks the security context runAsUser parameter in pods and containers to make sure it is not set to uid root(0)`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestSecConPrivilegeEscalation: {
		Identifier:  TestSecConPrivilegeEscalation,
		Type:        normativeResult,
		Remediation: `Configure privilege escalation to false`,
		Description: formDescription(TestSecConPrivilegeEscalation,
			`Checks if privileged escalation is enabled (AllowPrivilegeEscalation=true)`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestContainerIsCertifiedIdentifier: {
		Identifier:  TestContainerIsCertifiedIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that your container has passed the Red Hat Container Certification Program (CCP).`,
		Description: formDescription(TestContainerIsCertifiedIdentifier,
			`tests whether container images listed in the configuration file have passed the Red Hat Container Certification Program (CCP).`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.3.7",
	},
	TestContainerHostPort: {
		Identifier:  TestContainerHostPort,
		Type:        informativeResult,
		Remediation: `Remove hostPort configuration from the container`,
		Description: formDescription(TestContainerHostPort,
			`Verifies if containers define a hostPort.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.3.6",
	},
	TestPodHostNetwork: {
		Identifier:  TestPodHostNetwork,
		Type:        informativeResult,
		Remediation: `Set the spec.HostNetwork parameter to false in the pod configuration`,
		Description: formDescription(TestPodHostNetwork,
			`Verifies that the spec.HostNetwork parameter is set to false`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.3.6",
	},
	TestPodHostPath: {
		Identifier:  TestPodHostPath,
		Type:        informativeResult,
		Remediation: `Set the spec.HostPath parameter to false in the pod configuration`,
		Description: formDescription(TestPodHostPath,
			`Verifies that the spec.HostPath parameter is not set (not present)`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.3.6",
	},
	TestPodHostIPC: {
		Identifier:  TestPodHostIPC,
		Type:        informativeResult,
		Remediation: `Set the spec.HostIpc parameter to false in the pod configuration`,
		Description: formDescription(TestPodHostIPC,
			`Verifies that the spec.HostIpc parameter is set to false`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.3.6",
	},
	TestPodHostPID: {
		Identifier:  TestPodHostPID,
		Type:        informativeResult,
		Remediation: `Set the spec.HostPid parameter to false in the pod configuration`,
		Description: formDescription(TestPodHostPID,
			`Verifies that the spec.HostPid parameter is set to false`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.3.6",
	},
	TestHugepagesNotManuallyManipulated: {
		Identifier: TestHugepagesNotManuallyManipulated,
		Type:       normativeResult,
		Remediation: `HugePage settings should be configured either directly through the MachineConfigOperator or indirectly using the
PerformanceAddonOperator.  This ensures that OpenShift is aware of the special MachineConfig requirements, and can
provision your CNF on a Node that is part of the corresponding MachineConfigSet.  Avoid making changes directly to an
underlying Node, and let OpenShift handle the heavy lifting of configuring advanced settings.`,
		Description: formDescription(TestHugepagesNotManuallyManipulated,
			`checks to see that HugePage settings have been configured through MachineConfig, and not manually on the
underlying Node.  This test case applies only to Nodes that are configured with the "worker" MachineConfigSet.  First,
the "worker" MachineConfig is polled, and the Hugepage settings are extracted.  Next, the underlying Nodes are polled
for configured HugePages through inspection of /proc/meminfo.  The results are compared, and the test passes only if
they are the same.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},

	TestICMPv4ConnectivityIdentifier: {
		Identifier: TestICMPv4ConnectivityIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that the CNF is able to communicate via the Default OpenShift network. In some rare cases,
CNFs may require routing table changes in order to communicate over the Default network. To exclude a particular pod
from ICMPv4 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is not important, only its presence.`,
		Description: formDescription(TestICMPv4ConnectivityIdentifier,
			`checks that each CNF Container is able to communicate via ICMPv4 on the Default OpenShift network.  This
test case requires the Deployment of the debug daemonset.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},

	TestICMPv6ConnectivityIdentifier: {
		Identifier: TestICMPv6ConnectivityIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that the CNF is able to communicate via the Default OpenShift network. In some rare cases,
CNFs may require routing table changes in order to communicate over the Default network. To exclude a particular pod
from ICMPv6 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is not important, only its presence.`,
		Description: formDescription(TestICMPv6ConnectivityIdentifier,
			`checks that each CNF Container is able to communicate via ICMPv6 on the Default OpenShift network.  This
test case requires the Deployment of the debug daemonset.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},

	TestICMPv4ConnectivityMultusIdentifier: {
		Identifier: TestICMPv4ConnectivityMultusIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that the CNF is able to communicate via the Multus network(s). In some rare cases,
CNFs may require routing table changes in order to communicate over the Multus network(s). To exclude a particular pod
from ICMPv4 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it. The label value is not important, only its presence.`,
		Description: formDescription(TestICMPv4ConnectivityMultusIdentifier,
			`checks that each CNF Container is able to communicate via ICMPv4 on the Multus network(s).  This
test case requires the Deployment of the debug daemonset.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},

	TestICMPv6ConnectivityMultusIdentifier: {
		Identifier: TestICMPv6ConnectivityMultusIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that the CNF is able to communicate via the Multus network(s). In some rare cases,
CNFs may require routing table changes in order to communicate over the Multus network(s). To exclude a particular pod
from ICMPv6 connectivity tests, add the test-network-function.com/skip_connectivity_tests label to it.The label value is not important, only its presence.
`,
		Description: formDescription(TestICMPv6ConnectivityMultusIdentifier,
			`checks that each CNF Container is able to communicate via ICMPv6 on the Multus network(s).  This
test case requires the Deployment of the debug daemonset.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},

	TestNamespaceBestPracticesIdentifier: {
		Identifier: TestNamespaceBestPracticesIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that your CNF utilizes namespaces declared in the yaml config file. Additionally,
the namespaces should not start with "default, openshift-, istio- or aspenmesh-", except in rare cases.`,
		Description: formDescription(TestNamespaceBestPracticesIdentifier,
			`tests that all CNF's resources (PUTs and CRs) belong to valid namespaces. A valid namespace meets
the following conditions: (1) It was declared in the yaml config file under the targetNameSpaces
tag. (2) It doesn't have any of the following prefixes: default, openshift-, istio- and aspenmesh-`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2, 16.3.8 & 16.3.9",
	},

	TestNonTaintedNodeKernelsIdentifier: {
		Identifier: TestNonTaintedNodeKernelsIdentifier,
		Type:       normativeResult,
		Remediation: `Test failure indicates that the underlying Node's' kernel is tainted.  Ensure that you have not altered underlying
Node(s) kernels in order to run the CNF.`,
		Description: formDescription(TestNonTaintedNodeKernelsIdentifier,
			`ensures that the Node(s) hosting CNFs do not utilize tainted kernels. This test case is especially important
to support Highly Available CNFs, since when a CNF is re-instantiated on a backup Node, that Node's kernel may not have
the same hacks.'`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2.14",
	},

	TestOperatorInstallStatusSucceededIdentifier: {
		Identifier:  TestOperatorInstallStatusSucceededIdentifier,
		Type:        normativeResult,
		Remediation: `Make sure all the CNF operators have been successfully installed by OLM.`,
		Description: formDescription(TestOperatorInstallStatusSucceededIdentifier,
			`Ensures that the target CNF operators report "Succeeded" as their installation status.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2.12 and Section 6.3.3",
	},

	TestOperatorNoPrivileges: {
		Identifier:  TestOperatorNoPrivileges,
		Type:        normativeResult,
		Remediation: `Make sure all the CNF operators have no privileges on cluster resources.`,
		Description: formDescription(TestOperatorNoPrivileges,
			`The operator is not installed with privileged rights. Test passes if clusterPermissions is not present in the CSV manifest or is present 
with no resourceNames under its rules.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2.12 and Section 6.3.3",
	},

	TestOperatorIsCertifiedIdentifier: {
		Identifier:  TestOperatorIsCertifiedIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that your Operator has passed Red Hat's Operator Certification Program (OCP).`,
		Description: formDescription(TestOperatorIsCertifiedIdentifier,
			`tests whether CNF Operators listed in the configuration file have passed the Red Hat Operator Certification Program (OCP).`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2.12 and Section 6.3.3",
	},

	TestHelmIsCertifiedIdentifier: {
		Identifier:  TestHelmIsCertifiedIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that the helm charts under test passed the Red Hat's helm Certification Program (e.g. listed in https://charts.openshift.io/index.yaml).`,
		Description: formDescription(TestHelmIsCertifiedIdentifier,
			`tests whether helm charts listed in the cluster passed the Red Hat Helm Certification Program.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2.12 and Section 6.3.3",
	},

	TestOperatorIsInstalledViaOLMIdentifier: {
		Identifier:  TestOperatorIsInstalledViaOLMIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that your Operator is installed via OLM.`,
		Description: formDescription(TestOperatorIsInstalledViaOLMIdentifier,
			`tests whether a CNF Operator is installed via OLM.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2.12 and Section 6.3.3",
	},

	TestPodNodeSelectorAndAffinityBestPractices: {
		Identifier: TestPodNodeSelectorAndAffinityBestPractices,
		Type:       informativeResult,
		Remediation: `In most cases, Pod's should not specify their host Nodes through nodeSelector or nodeAffinity.  However, there are
cases in which CNFs require specialized hardware specific to a particular class of Node.  As such, this test is purely
informative, and will not prevent a CNF from being certified. However, one should have an appropriate justification as
to why nodeSelector and/or nodeAffinity is utilized by a CNF.`,
		Description: formDescription(TestPodNodeSelectorAndAffinityBestPractices,
			`ensures that CNF Pods do not specify nodeSelector or nodeAffinity.  In most cases, Pods should allow for
instantiation on any underlying Node.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},

	TestPodHighAvailabilityBestPractices: {
		Identifier:  TestPodHighAvailabilityBestPractices,
		Type:        informativeResult,
		Remediation: `In high availability cases, Pod replicas value should be set to more than 1 .`,
		Description: formDescription(TestPodHighAvailabilityBestPractices,
			`ensures that CNF Pods replicas value is set to more than 1.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},

	TestPodClusterRoleBindingsBestPracticesIdentifier: {
		Identifier: TestPodClusterRoleBindingsBestPracticesIdentifier,
		Type:       normativeResult,
		Remediation: `In most cases, Pod's should not have ClusterRoleBindings.  The suggested remediation is to remove the need for
ClusterRoleBindings, if possible.`,
		Description: formDescription(TestPodClusterRoleBindingsBestPracticesIdentifier,
			`tests that a Pod does not specify ClusterRoleBindings.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2.10 and 6.3.6",
	},

	TestPodDeploymentBestPracticesIdentifier: {
		Identifier:  TestPodDeploymentBestPracticesIdentifier,
		Type:        normativeResult,
		Remediation: `Deploy the CNF using ReplicaSet/StatefulSet.`,
		Description: formDescription(TestPodDeploymentBestPracticesIdentifier,
			`tests that CNF Pod(s) are deployed as part of a ReplicaSet(s)/StatefulSet(s).`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.3.3 and 6.3.8",
	},
	TestImagePullPolicyIdentifier: {
		Identifier:  TestImagePullPolicyIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that the containers under test are using IfNotPresent as Image Pull Policy.`,
		Description: formDescription(TestImagePullPolicyIdentifier,
			`Ensure that the containers under test are using IfNotPresent as Image Pull Policy..`),
		BestPracticeReference: bestPracticeDocV1dot3URL + "  Section 15.6",
	},

	TestPodRoleBindingsBestPracticesIdentifier: {
		Identifier:  TestPodRoleBindingsBestPracticesIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure the CNF is not configured to use RoleBinding(s) in a non-CNF Namespace.`,
		Description: formDescription(TestPodRoleBindingsBestPracticesIdentifier,
			`ensures that a CNF does not utilize RoleBinding(s) in a non-CNF Namespace.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.3.3 and 6.3.5",
	},

	TestPodServiceAccountBestPracticesIdentifier: {
		Identifier:  TestPodServiceAccountBestPracticesIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure that the each CNF Pod is configured to use a valid Service Account`,
		Description: formDescription(TestPodServiceAccountBestPracticesIdentifier,
			`tests that each CNF Pod utilizes a valid Service Account.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2.3 and 6.2.7",
	},

	TestServicesDoNotUseNodeportsIdentifier: {
		Identifier:  TestServicesDoNotUseNodeportsIdentifier,
		Type:        normativeResult,
		Remediation: `Ensure Services are not configured to use NodePort(s).`,
		Description: formDescription(TestServicesDoNotUseNodeportsIdentifier,
			`tests that each CNF Service does not utilize NodePort(s).`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.3.1",
	},

	TestUnalteredBaseImageIdentifier: {
		Identifier: TestUnalteredBaseImageIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that Container applications do not modify the Container Base Image.  In particular, ensure that the following
directories are not modified:
1) /var/lib/rpm
2) /var/lib/dpkg
3) /bin
4) /sbin
5) /lib
6) /lib64
7) /usr/bin
8) /usr/sbin
9) /usr/lib
10) /usr/lib64
Ensure that all required binaries are built directly into the container image, and are not installed post startup.`,
		Description: formDescription(TestUnalteredBaseImageIdentifier,
			`ensures that the Container Base Image is not altered post-startup.  This test is a heuristic, and ensures
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
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2.2",
	},

	TestUnalteredStartupBootParamsIdentifier: {
		Identifier: TestUnalteredStartupBootParamsIdentifier,
		Type:       normativeResult,
		Remediation: `Ensure that boot parameters are set directly through the MachineConfigOperator, or indirectly through the PerformanceAddonOperator.  Boot parameters should not be changed directly through the Node, as OpenShift should manage
the changes for you.`,
		Description: formDescription(TestUnalteredStartupBootParamsIdentifier,
			`tests that boot parameters are set through the MachineConfigOperator, and not set manually on the Node.`),
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2.13 and 6.2.14",
	},
	TestShudtownIdentifier: {
		Identifier: TestShudtownIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestShudtownIdentifier,
			`Ensure that the containers lifecycle pre-stop management feature is configured.`),
		Remediation: `
		It's considered best-practices to define prestop for proper management of container lifecycle.
		The prestop can be used to gracefully stop the container and clean resources (e.g., DB connection).
		
		The prestop can be configured using :
		 1) Exec : executes the supplied command inside the container
		 2) HTTP : executes HTTP request against the specified endpoint.
		
		When defined. K8s will handle shutdown of the container using the following:
		1) K8s first execute the preStop hook inside the container.
		2) K8s will wait for a grace period.
		3) K8s will clean the remaining processes using KILL signal.		
			`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestPodRecreationIdentifier: {
		Identifier: TestPodRecreationIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestPodRecreationIdentifier,
			`tests that a CNF is configured to support High Availability.  
			First, this test cordons and drains a Node that hosts the CNF Pod.  
			Next, the test ensures that OpenShift can re-instantiate the Pod on another Node, 
			and that the actual replica count matches the desired replica count.`),
		Remediation: `Ensure that CNF Pod(s) utilize a configuration that supports High Availability.  
			Additionally, ensure that there are available Nodes in the OpenShift cluster that can be utilized in the event that a host Node fails.`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestSysctlConfigsIdentifier: {
		Identifier: TestSysctlConfigsIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestPodRecreationIdentifier,
			`tests that no one has changed the node's sysctl configs after the node
			was created, the tests works by checking if the sysctl configs are consistent with the
			MachineConfig CR which defines how the node should be configured`),
		Remediation:           `You should recreate the node or change the sysctls, recreating is recommended because there might be other unknown changes`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestScalingIdentifier: {
		Identifier: TestScalingIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestScalingIdentifier,
			`tests that CNF deployments support scale in/out operations. 
			First, The test starts getting the current replicaCount (N) of the deployment/s with the Pod Under Test. Then, it executes the 
			scale-in oc command for (N-1) replicas. Lastly, it executes the scale-out oc command, restoring the original replicaCount of the deployment/s.`),
		Remediation:           `Make sure CNF deployments/replica sets can scale in/out successfully.`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestIsRedHatReleaseIdentifier: {
		Identifier: TestIsRedHatReleaseIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestIsRedHatReleaseIdentifier,
			`verifies if the container base image is redhat.`),
		Remediation:           `build a new docker image that's based on UBI (redhat universal base image).`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestIsSELinuxEnforcingIdentifier: {
		Identifier: TestIsSELinuxEnforcingIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestIsSELinuxEnforcingIdentifier,
			`verifies that all openshift platform/cluster nodes have selinux in "Enforcing" mode.`),
		Remediation:           `configure selinux and enable enforcing mode.`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 11.3 Pod Security",
	},
	TestUndeclaredContainerPortsUsage: {
		Identifier: TestUndeclaredContainerPortsUsage,
		Type:       normativeResult,
		Description: formDescription(TestUndeclaredContainerPortsUsage,
			`check that containers don't listen on ports that weren't declared in their specification`),
		Remediation:           `ensure the CNF apps don't listen on undeclared containers' ports`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 16.3.1.1",
	},
	TestCrdsStatusSubresourceIdentifier: {
		Identifier: TestCrdsStatusSubresourceIdentifier,
		Type:       informativeResult,
		Description: formDescription(TestCrdsStatusSubresourceIdentifier,
			`checks that all CRDs have a status subresource specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties["status"]).`),
		Remediation:           `make sure that all the CRDs have a meaningful status specification (Spec.versions[].Schema.OpenAPIV3Schema.Properties["status"]).`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 6.2",
	},
	TestLoggingIdentifier: {
		Identifier: TestLoggingIdentifier,
		Type:       informativeResult,
		Description: formDescription(TestLoggingIdentifier,
			`check that all containers under test use standard input output and standard error when logging`),
		Remediation:           `make sure containers are not redirecting stdout/stderr`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 11.1",
	},
	TestTerminationMessagePolicyIdentifier: {
		Identifier: TestTerminationMessagePolicyIdentifier,
		Type:       informativeResult,
		Description: formDescription(TestTerminationMessagePolicyIdentifier,
			`check that all containers are using terminationMessagePolicy: FallbackToLogsOnError`),
		Remediation:           `make sure containers are all using FallbackToLogsOnError in terminationMessagePolicy`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 15.1",
	},
	TestPodAutomountServiceAccountIdentifier: {
		Identifier: TestPodAutomountServiceAccountIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestPodAutomountServiceAccountIdentifier,
			`check that all pods under test have automountServiceAccountToken set to false`),
		Remediation: `check that pod has automountServiceAccountToken set to false or pod is attached to service account which has automountServiceAccountToken set to false`,
	},
	TestLivenessProbeIdentifier: {
		Identifier: TestLivenessProbeIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestLivenessProbeIdentifier, `check that all containers under test
		have liveness probe defined`),
		Remediation: `add liveness probe to deployed containers`,
	},
	TestReadinessProbeIdentifier: {
		Identifier: TestReadinessProbeIdentifier,
		Type:       normativeResult,
		Description: formDescription(TestReadinessProbeIdentifier, `check that all containers under test
		have readiness probe defined`),
		Remediation: `add readiness probe to deployed containers`,
	},
	TestOneProcessPerContainerIdentifier: {
		Identifier: TestOneProcessPerContainerIdentifier,
		Type:       informativeResult,
		Description: formDescription(TestOneProcessPerContainerIdentifier, `check that all containers under test
		have only one process running`),
		Remediation:           `launch only one process per container`,
		BestPracticeReference: bestPracticeDocV1dot2URL + " Section 11.8.3",
	},
}
