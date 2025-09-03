// Copyright (C) 2021-2024 Red Hat, Inc.
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

/*
	Use this file to store the impact statements for each test in our suite.
	Impact statements explain the consequences of not passing the test - what could go wrong
	in production or what risks are introduced when the test fails.
*/

const (
	// Networking Suite Impact Statements
	TestICMPv4ConnectivityIdentifierImpact             = `Failure indicates potential network isolation issues that could prevent workload components from communicating, leading to service degradation or complete application failure.`
	TestNetworkPolicyDenyAllIdentifierImpact           = `Without default deny-all network policies, workloads are exposed to lateral movement attacks and unauthorized network access, compromising security posture and potentially enabling data breaches.`
	TestReservedExtendedPartnerPortsImpact             = `Using reserved ports can cause port conflicts with essential platform services, leading to service startup failures and unpredictable application behavior.`
	TestDpdkCPUPinningExecProbeImpact                  = `Exec probes on CPU-pinned DPDK workloads can cause performance degradation, interrupt real-time operations, and potentially crash applications due to resource contention.`
	TestRestartOnRebootLabelOnPodsUsingSRIOVImpact     = `Without restart-on-reboot labels, SRIOV-enabled pods may fail to recover from a race condition between kubernetes services startup and SR-IOV device plugin configuration on StarlingX AIO systems, causing SR-IOV devices to disappear from running pods when FPGA devices are reset.`
	TestNetworkAttachmentDefinitionSRIOVUsingMTUImpact = `Incorrect MTU settings can cause packet fragmentation, network performance issues, and connectivity failures in high-performance networking scenarios.`
	TestLimitedUseOfExecProbesIdentifierImpact         = `Excessive exec probes can overwhelm system resources, degrade performance, and interfere with critical application operations in resource-constrained environments.`
	TestICMPv6ConnectivityIdentifierImpact             = `IPv6 connectivity failures can prevent dual-stack applications from functioning properly and limit future network architecture flexibility.`
	TestICMPv4ConnectivityMultusIdentifierImpact       = `Multus network connectivity issues can isolate workloads from secondary networks, breaking multi-network applications and reducing network redundancy.`
	TestICMPv6ConnectivityMultusIdentifierImpact       = `IPv6 Multus connectivity problems can prevent dual-stack multi-network scenarios from working, limiting network scalability and future-proofing.`
	TestServiceDualStackIdentifierImpact               = `Single-stack IPv4 services limit network architecture flexibility and prevent migration to modern dual-stack infrastructures.`
	TestUndeclaredContainerPortsUsageImpact            = `Undeclared ports can be blocked by security policies, causing unexpected connectivity issues and making troubleshooting difficult.`
	TestOCPReservedPortsUsageImpact                    = `Using OpenShift-reserved ports can cause critical platform services to fail, potentially destabilizing the entire cluster.`

	// Access Control Suite Impact Statements
	Test1337UIDIdentifierImpact                             = `UID 1337 is reserved for use by Istio service mesh components; using it for applications can cause conflicts with Istio sidecars and break service mesh functionality.`
	TestNetAdminIdentifierImpact                            = `NET_ADMIN capability allows network configuration changes that can compromise cluster networking, enable privilege escalation, and bypass network security controls.`
	TestSysAdminIdentifierImpact                            = `SYS_ADMIN capability provides extensive privileges that can compromise container isolation, enable host system access, and create serious security vulnerabilities.`
	TestIpcLockIdentifierImpact                             = `IPC_LOCK capability can be exploited to lock system memory, potentially causing denial of service and affecting other workloads on the same node.`
	TestNetRawIdentifierImpact                              = `NET_RAW capability enables packet manipulation and network sniffing, which can be used for attacks against other workloads and compromise network security.`
	TestBpfIdentifierImpact                                 = `BPF capability allows kernel-level programming that can bypass security controls, monitor other processes, and potentially compromise the entire host system.`
	TestSecConNonRootUserIdentifierImpact                   = `Running containers as root increases the blast radius of security vulnerabilities and can lead to full host compromise if containers are breached.`
	TestSecContextIdentifierImpact                          = `Incorrect security context configurations can weaken container isolation, enable privilege escalation, and create exploitable attack vectors.`
	TestSecConPrivilegeEscalationImpact                     = `Allowing privilege escalation can lead to containers gaining root access, compromising the security boundary between containers and hosts.`
	TestContainerHostPortImpact                             = `Host port usage can create port conflicts with host services and expose containers directly to the host network, bypassing network security controls.`
	TestPodHostNetworkImpact                                = `Host network access removes network isolation, exposes containers to host network interfaces, and can compromise cluster networking security.`
	TestPodHostPathImpact                                   = `Host path mounts can expose sensitive host files to containers, enable container escape attacks, and compromise host system integrity.`
	TestPodHostIPCImpact                                    = `Host IPC access allows containers to communicate with host processes, potentially exposing sensitive information and enabling privilege escalation.`
	TestPodHostPIDImpact                                    = `Host PID access allows containers to see and interact with all host processes, creating opportunities for privilege escalation and information disclosure.`
	TestNamespaceBestPracticesIdentifierImpact              = `Using inappropriate namespaces can lead to resource conflicts, security boundary violations, and administrative complexity in multi-tenant environments.`
	TestPodClusterRoleBindingsBestPracticesIdentifierImpact = `Cluster-wide role bindings grant excessive privileges that can be exploited for lateral movement and privilege escalation across the entire cluster.`
	TestPodRoleBindingsBestPracticesIdentifierImpact        = `Cross-namespace role bindings can violate tenant isolation and create unintended privilege escalation paths.`
	TestPodServiceAccountBestPracticesIdentifierImpact      = `Default service accounts often have excessive privileges; improper usage can lead to unauthorized API access and security violations.`
	TestPodAutomountServiceAccountIdentifierImpact          = `Auto-mounted service account tokens expose Kubernetes API credentials to application code, creating potential attack vectors if applications are compromised.`
	TestServicesDoNotUseNodeportsIdentifierImpact           = `NodePort services expose applications directly on host ports, creating security risks and potential port conflicts with host services.`
	TestOneProcessPerContainerIdentifierImpact              = `Multiple processes per container complicate monitoring, debugging, and security assessment, and can lead to zombie processes and resource leaks.`
	TestSYSNiceRealtimeCapabilityIdentifierImpact           = `Missing SYS_NICE capability on real-time nodes prevents applications from setting appropriate scheduling priorities, causing performance degradation.`
	TestSysPtraceCapabilityIdentifierImpact                 = `Missing SYS_PTRACE capability when using shared process namespaces prevents inter-container process communication, breaking application functionality.`
	TestPodRequestsIdentifierImpact                         = `Missing resource requests can lead to resource contention, node instability, and unpredictable application performance.`
	TestNamespaceResourceQuotaIdentifierImpact              = `Without resource quotas, workloads can consume excessive cluster resources, causing performance issues and potential denial of service for other applications.`
	TestSecConReadOnlyFilesystemImpact                      = `Writable root filesystems increase the attack surface and can be exploited to modify container behavior or persist malware.`
	TestNoSSHDaemonsAllowedIdentifierImpact                 = `SSH daemons in containers create additional attack surfaces, violate immutable infrastructure principles, and can be exploited for unauthorized access.`

	// Affiliated Certification Suite Impact Statements
	TestHelmVersionIdentifierImpact                = `Helm v2 has known security vulnerabilities and lacks proper RBAC controls, creating significant security risks in production environments.`
	TestContainerIsCertifiedDigestIdentifierImpact = `Uncertified containers may contain security vulnerabilities, lack enterprise support, and fail to meet compliance requirements.`
	TestOperatorIsCertifiedIdentifierImpact        = `Uncertified operators may have security flaws, compatibility issues, and lack enterprise support, creating operational risks.`
	TestHelmIsCertifiedIdentifierImpact            = `Uncertified helm charts may contain security vulnerabilities, configuration errors, and lack proper testing, leading to deployment failures.`

	// Platform Alteration Suite Impact Statements
	TestPodHugePages2MImpact                       = `Using inappropriate hugepage sizes can cause memory allocation failures and reduce overall system performance and stability.`
	TestPodHugePages1GImpact                       = `Incorrect hugepage configuration can lead to memory fragmentation and application startup failures in memory-constrained environments.`
	TestHugepagesNotManuallyManipulatedImpact      = `Manual hugepage configuration bypasses cluster management, can cause node instability, and creates configuration drift issues.`
	TestNonTaintedNodeKernelsIdentifierImpact      = `Tainted kernels indicate unauthorized modifications that can introduce instability, security vulnerabilities, and support issues.`
	TestUnalteredBaseImageIdentifierImpact         = `Modified base images can introduce security vulnerabilities, create inconsistent behavior, and violate immutable infrastructure principles.`
	TestUnalteredStartupBootParamsIdentifierImpact = `Manual boot parameter changes bypass cluster configuration management and can cause node instability and configuration drift.`
	TestSysctlConfigsIdentifierImpact              = `Manual sysctl modifications can cause system instability, security vulnerabilities, and unpredictable kernel behavior.`
	TestServiceMeshIdentifierImpact                = `Inconsistent service mesh configuration can create security gaps, monitoring blind spots, and traffic management issues.`
	TestOCPLifecycleIdentifierImpact               = `End-of-life OpenShift versions lack security updates and support, creating significant security and operational risks.`
	TestNodeOperatingSystemIdentifierImpact        = `Incompatible node operating systems can cause stability issues, security vulnerabilities, and lack of vendor support.`
	TestIsRedHatReleaseIdentifierImpact            = `Non-Red Hat base images may lack security updates, enterprise support, and compliance certifications required for production use.`
	TestClusterOperatorHealthImpact                = `Unhealthy cluster operators can cause platform instability, feature failures, and degraded cluster functionality.`
	TestIsSELinuxEnforcingIdentifierImpact         = `Non-enforcing SELinux reduces security isolation and can allow privilege escalation attacks and unauthorized resource access.`
	TestHyperThreadEnableImpact                    = `Disabled hyperthreading reduces CPU performance and can affect workload scheduling and resource utilization efficiency.`

	// Lifecycle Suite Impact Statements
	TestAffinityRequiredPodsImpact                    = `Missing affinity rules can cause incorrect pod placement, leading to performance issues and failure to meet co-location requirements.`
	TestStorageProvisionerImpact                      = `Inappropriate storage provisioners can cause data persistence issues, performance problems, and storage failures.`
	TestContainerPostStartIdentifierImpact            = `Missing PostStart hooks can cause containers to start serving traffic before proper initialization, leading to application errors.`
	TestContainerPrestopIdentifierImpact              = `Missing PreStop hooks can cause ungraceful shutdowns, data loss, and connection drops during container termination.`
	TestPodNodeSelectorAndAffinityBestPracticesImpact = `Node selectors can create scheduling constraints that reduce cluster flexibility and cause deployment failures when nodes are unavailable.`
	TestPodHighAvailabilityBestPracticesImpact        = `Missing anti-affinity rules can cause all pod replicas to be scheduled on the same node, creating single points of failure.`
	TestPodDeploymentBestPracticesIdentifierImpact    = `Naked pods and DaemonSets lack proper lifecycle management, making updates, scaling, and recovery operations difficult or impossible.`
	TestDeploymentScalingIdentifierImpact             = `Deployment scaling failures prevent horizontal scaling operations, limiting application elasticity and availability during high load.`
	TestStatefulSetScalingIdentifierImpact            = `StatefulSet scaling issues can prevent proper data persistence and ordered deployment of stateful applications.`
	TestImagePullPolicyIdentifierImpact               = `Incorrect image pull policies can cause deployment failures when image registries are unavailable or during network issues.`
	TestPodRecreationIdentifierImpact                 = `Failed pod recreation indicates poor high availability configuration, leading to potential service outages during node failures.`
	TestLivenessProbeIdentifierImpact                 = `Missing liveness probes prevent Kubernetes from detecting and recovering from application deadlocks and hangs.`
	TestReadinessProbeIdentifierImpact                = `Missing readiness probes can cause traffic to be routed to non-ready pods, resulting in failed requests and poor user experience.`
	TestStartupProbeIdentifierImpact                  = `Missing startup probes can cause slow-starting applications to be killed prematurely, preventing successful application startup.`
	TestPodTolerationBypassIdentifierImpact           = `Modified tolerations can allow pods to be scheduled on inappropriate nodes, violating scheduling policies and causing performance issues.`
	TestPersistentVolumeReclaimPolicyIdentifierImpact = `Incorrect reclaim policies can lead to data persistence after application removal, causing storage waste and potential data security issues.`
	TestCPUIsolationIdentifierImpact                  = `Improper CPU isolation can cause performance interference between workloads and fail to provide guaranteed compute resources.`
	TestCrdScalingIdentifierImpact                    = `CRD scaling failures can prevent operator-managed applications from scaling properly, limiting application availability and performance.`

	// Performance Test Suite Impact Statements
	TestExclusiveCPUPoolIdentifierImpact       = `Inconsistent CPU pool selection can cause performance interference and unpredictable latency in real-time applications.`
	TestSharedCPUPoolSchedulingPolicyImpact    = `Incorrect scheduling policies in shared CPU pools can cause performance interference and unfair resource distribution.`
	TestExclusiveCPUPoolSchedulingPolicyImpact = `Wrong scheduling policies in exclusive CPU pools can prevent real-time applications from meeting latency requirements.`
	TestIsolatedCPUPoolSchedulingPolicyImpact  = `Incorrect scheduling policies in isolated CPU pools can cause performance degradation and violate real-time guarantees.`
	TestRtAppNoExecProbesImpact                = `Exec probes on real-time applications can cause latency spikes and interrupt time-critical operations.`

	// Operator Test Suite Impact Statements
	TestOperatorInstallStatusSucceededIdentifierImpact                     = `Failed operator installations can leave applications in incomplete states, causing functionality gaps and operational issues.`
	TestOperatorNoSCCAccessImpact                                          = `Operators with SCC access have elevated privileges that can compromise cluster security and violate security policies.`
	TestOperatorIsInstalledViaOLMIdentifierImpact                          = `Non-OLM operators bypass lifecycle management and dependency resolution, creating operational complexity and update issues.`
	TestSingleOrMultiNamespacedOperatorInstallationInTenantNamespaceImpact = `Improperly scoped operators can violate tenant isolation and create unauthorized cross-namespace access.`
	TestOperatorHasSemanticVersioningIdentifierImpact                      = `Invalid semantic versioning prevents proper upgrade paths and dependency management, causing operational issues.`
	TestOperatorOlmSkipRangeImpact                                         = `Invalid skip ranges can prevent proper operator upgrades and cause version compatibility issues.`
	TestOperatorCrdVersioningIdentifierImpact                              = `Invalid CRD versioning can cause API compatibility issues and prevent proper schema evolution.`
	TestOperatorCrdSchemaIdentifierImpact                                  = `Missing OpenAPI schemas prevent proper validation and can lead to configuration errors and runtime failures.`
	TestOperatorSingleCrdOwnerIdentifierImpact                             = `Multiple CRD owners can cause conflicts, inconsistent behavior, and management complexity.`
	TestOperatorPodsNoHugepagesImpact                                      = `Hugepage usage by operators can interfere with application hugepage allocation and cause resource contention.`
	TestOperatorCatalogSourceBundleCountIdentifierImpact                   = `Large catalog sources can cause performance issues, slow operator resolution, and increase cluster resource usage.`
	TestMultipleSameOperatorsIdentifierImpact                              = `Multiple operator instances can cause conflicts, resource contention, and unpredictable behavior.`

	// Observability Test Suite Impact Statements
	TestLoggingIdentifierImpact                            = `Improper logging configuration prevents log aggregation and monitoring, making troubleshooting and debugging difficult.`
	TestTerminationMessagePolicyIdentifierImpact           = `Incorrect termination message policies can prevent proper error reporting and make failure diagnosis difficult.`
	TestCrdsStatusSubresourceIdentifierImpact              = `Missing status subresources prevent proper monitoring and automation based on custom resource states.`
	TestPodDisruptionBudgetIdentifierImpact                = `Improper disruption budgets can prevent necessary maintenance operations or allow too many pods to be disrupted simultaneously.`
	TestAPICompatibilityWithNextOCPReleaseIdentifierImpact = `Deprecated API usage can cause applications to break during OpenShift upgrades, requiring emergency fixes.`

	// Manageability Test Suite Impact Statements
	TestContainersImageTagImpact      = `Missing image tags make it difficult to track versions, perform rollbacks, and maintain deployment consistency.`
	TestContainerPortNameFormatImpact = `Incorrect port naming conventions can cause service discovery issues and configuration management problems.`

	// Preflight Test Suite Impact Statements
	PreflightAllImageRefsInRelatedImagesImpact                  = `Missing or incorrect image references in related images can cause deployment failures and broken operator functionality.`
	PreflightBasedOnUbiImpact                                   = `Non-UBI base images may lack security updates, enterprise support, and compliance certifications required for production use.`
	PreflightBundleImageRefsAreCertifiedImpact                  = `Uncertified bundle image references can introduce security vulnerabilities and compatibility issues in production deployments.`
	PreflightDeployableByOLMImpact                              = `Operators not deployable by OLM cannot be properly managed, updated, or integrated into OpenShift lifecycle management.`
	PreflightFollowsRestrictedNetworkEnablementGuidelinesImpact = `Non-compliance with restricted network guidelines can prevent deployment in air-gapped environments and violate security policies.`
	PreflightHasLicenseImpact                                   = `Missing license information can create legal compliance issues and prevent proper software asset management.`
	PreflightHasModifiedFilesImpact                             = `Modified files in containers can introduce security vulnerabilities, create inconsistent behavior, and violate immutable infrastructure principles.`
	PreflightHasNoProhibitedLabelsImpact                        = `Misuse of Red Hat trademarks in name, vendor, or maintainer labels creates legal and compliance risks that can block certification and publication.`
	PreflightHasNoProhibitedPackagesImpact                      = `Prohibited packages can introduce security vulnerabilities, licensing issues, and compliance violations.`
	PreflightHasProhibitedContainerNameImpact                   = `Prohibited container names can cause conflicts with system components and violate naming conventions.`
	PreflightHasRequiredLabelImpact                             = `Missing required labels prevent proper metadata management and can cause deployment and management issues.`
	PreflightHasUniqueTagImpact                                 = `Non-unique tags can cause version conflicts and deployment inconsistencies, making rollbacks and troubleshooting difficult.`
	PreflightLayerCountAcceptableImpact                         = `Excessive image layers can cause poor performance, increased storage usage, and longer deployment times.`
	PreflightRequiredAnnotationsImpact                          = `Missing required annotations can prevent proper operator lifecycle management and cause deployment failures.`
	PreflightRunAsNonRootImpact                                 = `Running containers as root increases the blast radius of security vulnerabilities and can lead to full host compromise if containers are breached.`
	PreflightScorecardBasicSpecCheckImpact                      = `Failing basic scorecard checks indicates fundamental operator implementation issues that can cause runtime failures.`
	PreflightScorecardOlmSuiteCheckImpact                       = `Failing OLM suite checks indicates operator lifecycle management issues that can prevent proper installation and updates.`
	PreflightSecurityContextConstraintsInCSVImpact              = `Incorrect SCC definitions in CSV can cause security policy violations and deployment failures.`
	PreflightValidateOperatorBundleImpact                       = `Invalid operator bundles can cause deployment failures, update issues, and operational instability.`
)

// ImpactMap maps test IDs to their impact statements
var ImpactMap = map[string]string{
	// Networking Suite
	"networking-icmpv4-connectivity":                     TestICMPv4ConnectivityIdentifierImpact,
	"networking-network-policy-deny-all":                 TestNetworkPolicyDenyAllIdentifierImpact,
	"networking-reserved-partner-ports":                  TestReservedExtendedPartnerPortsImpact,
	"networking-dpdk-cpu-pinning-exec-probe":             TestDpdkCPUPinningExecProbeImpact,
	"networking-restart-on-reboot-sriov-pod":             TestRestartOnRebootLabelOnPodsUsingSRIOVImpact,
	"networking-network-attachment-definition-sriov-mtu": TestNetworkAttachmentDefinitionSRIOVUsingMTUImpact,
	"performance-max-resources-exec-probes":              TestLimitedUseOfExecProbesIdentifierImpact,
	"networking-icmpv6-connectivity":                     TestICMPv6ConnectivityIdentifierImpact,
	"networking-icmpv4-connectivity-multus":              TestICMPv4ConnectivityMultusIdentifierImpact,
	"networking-icmpv6-connectivity-multus":              TestICMPv6ConnectivityMultusIdentifierImpact,
	"networking-dual-stack-service":                      TestServiceDualStackIdentifierImpact,
	"networking-undeclared-container-ports-usage":        TestUndeclaredContainerPortsUsageImpact,
	"networking-ocp-reserved-ports-usage":                TestOCPReservedPortsUsageImpact,

	// Access Control Suite
	"access-control-no-1337-uid":                             Test1337UIDIdentifierImpact,
	"access-control-net-admin-capability-check":              TestNetAdminIdentifierImpact,
	"access-control-sys-admin-capability-check":              TestSysAdminIdentifierImpact,
	"access-control-ipc-lock-capability-check":               TestIpcLockIdentifierImpact,
	"access-control-net-raw-capability-check":                TestNetRawIdentifierImpact,
	"access-control-bpf-capability-check":                    TestBpfIdentifierImpact,
	"access-control-security-context-non-root-user-id-check": TestSecConNonRootUserIdentifierImpact,
	"access-control-security-context":                        TestSecContextIdentifierImpact,
	"access-control-security-context-privilege-escalation":   TestSecConPrivilegeEscalationImpact,
	"access-control-container-host-port":                     TestContainerHostPortImpact,
	"access-control-pod-host-network":                        TestPodHostNetworkImpact,
	"access-control-pod-host-path":                           TestPodHostPathImpact,
	"access-control-pod-host-ipc":                            TestPodHostIPCImpact,
	"access-control-pod-host-pid":                            TestPodHostPIDImpact,
	"access-control-namespace":                               TestNamespaceBestPracticesIdentifierImpact,
	"access-control-cluster-role-bindings":                   TestPodClusterRoleBindingsBestPracticesIdentifierImpact,
	"access-control-pod-role-bindings":                       TestPodRoleBindingsBestPracticesIdentifierImpact,
	"access-control-pod-service-account":                     TestPodServiceAccountBestPracticesIdentifierImpact,
	"access-control-pod-automount-service-account-token":     TestPodAutomountServiceAccountIdentifierImpact,
	"access-control-service-type":                            TestServicesDoNotUseNodeportsIdentifierImpact,
	"access-control-one-process-per-container":               TestOneProcessPerContainerIdentifierImpact,
	"access-control-sys-nice-realtime-capability":            TestSYSNiceRealtimeCapabilityIdentifierImpact,
	"access-control-sys-ptrace-capability":                   TestSysPtraceCapabilityIdentifierImpact,
	"access-control-requests":                                TestPodRequestsIdentifierImpact,
	"access-control-namespace-resource-quota":                TestNamespaceResourceQuotaIdentifierImpact,
	"access-control-security-context-read-only-file-system":  TestSecConReadOnlyFilesystemImpact,
	"access-control-ssh-daemons":                             TestNoSSHDaemonsAllowedIdentifierImpact,

	// Affiliated Certification Suite
	"affiliated-certification-helm-version":                  TestHelmVersionIdentifierImpact,
	"affiliated-certification-container-is-certified-digest": TestContainerIsCertifiedDigestIdentifierImpact,
	"affiliated-certification-operator-is-certified":         TestOperatorIsCertifiedIdentifierImpact,
	"affiliated-certification-helmchart-is-certified":        TestHelmIsCertifiedIdentifierImpact,

	// Platform Alteration Suite
	"platform-alteration-hugepages-2m-only":       TestPodHugePages2MImpact,
	"platform-alteration-hugepages-1g-only":       TestPodHugePages1GImpact,
	"platform-alteration-hugepages-config":        TestHugepagesNotManuallyManipulatedImpact,
	"platform-alteration-tainted-node-kernel":     TestNonTaintedNodeKernelsIdentifierImpact,
	"platform-alteration-base-image":              TestUnalteredBaseImageIdentifierImpact,
	"platform-alteration-boot-params":             TestUnalteredStartupBootParamsIdentifierImpact,
	"platform-alteration-sysctl-config":           TestSysctlConfigsIdentifierImpact,
	"platform-alteration-service-mesh-usage":      TestServiceMeshIdentifierImpact,
	"platform-alteration-ocp-lifecycle":           TestOCPLifecycleIdentifierImpact,
	"platform-alteration-ocp-node-os-lifecycle":   TestNodeOperatingSystemIdentifierImpact,
	"platform-alteration-isredhat-release":        TestIsRedHatReleaseIdentifierImpact,
	"platform-alteration-cluster-operator-health": TestClusterOperatorHealthImpact,
	"platform-alteration-is-selinux-enforcing":    TestIsSELinuxEnforcingIdentifierImpact,
	"platform-alteration-hyperthread-enable":      TestHyperThreadEnableImpact,

	// Lifecycle Suite
	"lifecycle-affinity-required-pods":           TestAffinityRequiredPodsImpact,
	"lifecycle-storage-provisioner":              TestStorageProvisionerImpact,
	"lifecycle-container-poststart":              TestContainerPostStartIdentifierImpact,
	"lifecycle-container-prestop":                TestContainerPrestopIdentifierImpact,
	"lifecycle-pod-scheduling":                   TestPodNodeSelectorAndAffinityBestPracticesImpact,
	"lifecycle-pod-high-availability":            TestPodHighAvailabilityBestPracticesImpact,
	"lifecycle-pod-owner-type":                   TestPodDeploymentBestPracticesIdentifierImpact,
	"lifecycle-deployment-scaling":               TestDeploymentScalingIdentifierImpact,
	"lifecycle-statefulset-scaling":              TestStatefulSetScalingIdentifierImpact,
	"lifecycle-image-pull-policy":                TestImagePullPolicyIdentifierImpact,
	"lifecycle-pod-recreation":                   TestPodRecreationIdentifierImpact,
	"lifecycle-liveness-probe":                   TestLivenessProbeIdentifierImpact,
	"lifecycle-readiness-probe":                  TestReadinessProbeIdentifierImpact,
	"lifecycle-startup-probe":                    TestStartupProbeIdentifierImpact,
	"lifecycle-pod-toleration-bypass":            TestPodTolerationBypassIdentifierImpact,
	"lifecycle-persistent-volume-reclaim-policy": TestPersistentVolumeReclaimPolicyIdentifierImpact,
	"lifecycle-cpu-isolation":                    TestCPUIsolationIdentifierImpact,
	"lifecycle-crd-scaling":                      TestCrdScalingIdentifierImpact,

	// Performance Test Suite
	"performance-exclusive-cpu-pool":                       TestExclusiveCPUPoolIdentifierImpact,
	"performance-shared-cpu-pool-non-rt-scheduling-policy": TestSharedCPUPoolSchedulingPolicyImpact,
	"performance-exclusive-cpu-pool-rt-scheduling-policy":  TestExclusiveCPUPoolSchedulingPolicyImpact,
	"performance-isolated-cpu-pool-rt-scheduling-policy":   TestIsolatedCPUPoolSchedulingPolicyImpact,
	"performance-rt-apps-no-exec-probes":                   TestRtAppNoExecProbesImpact,

	// Operator Test Suite
	"operator-install-status-succeeded":                                TestOperatorInstallStatusSucceededIdentifierImpact,
	"operator-install-status-no-privileges":                            TestOperatorNoSCCAccessImpact,
	"operator-install-source":                                          TestOperatorIsInstalledViaOLMIdentifierImpact,
	"operator-single-or-multi-namespaced-allowed-in-tenant-namespaces": TestSingleOrMultiNamespacedOperatorInstallationInTenantNamespaceImpact,
	"operator-semantic-versioning":                                     TestOperatorHasSemanticVersioningIdentifierImpact,
	"operator-olm-skip-range":                                          TestOperatorOlmSkipRangeImpact,
	"operator-crd-versioning":                                          TestOperatorCrdVersioningIdentifierImpact,
	"operator-crd-openapi-schema":                                      TestOperatorCrdSchemaIdentifierImpact,
	"operator-single-crd-owner":                                        TestOperatorSingleCrdOwnerIdentifierImpact,
	"operator-pods-no-hugepages":                                       TestOperatorPodsNoHugepagesImpact,
	"operator-catalogsource-bundle-count":                              TestOperatorCatalogSourceBundleCountIdentifierImpact,
	"operator-multiple-same-operators":                                 TestMultipleSameOperatorsIdentifierImpact,

	// Observability Test Suite
	"observability-container-logging":                   TestLoggingIdentifierImpact,
	"observability-termination-policy":                  TestTerminationMessagePolicyIdentifierImpact,
	"observability-crd-status":                          TestCrdsStatusSubresourceIdentifierImpact,
	"observability-pod-disruption-budget":               TestPodDisruptionBudgetIdentifierImpact,
	"observability-compatibility-with-next-ocp-release": TestAPICompatibilityWithNextOCPReleaseIdentifierImpact,

	// Manageability Test Suite
	"manageability-containers-image-tag":       TestContainersImageTagImpact,
	"manageability-container-port-name-format": TestContainerPortNameFormatImpact,

	// Access Control Suite (additional)
	"access-control-crd-roles": `Improper CRD role configurations can grant excessive privileges, violate least-privilege principles, and create security vulnerabilities in custom resource access control.`,

	// Preflight Test Suite
	"preflight-AllImageRefsInRelatedImages":                  PreflightAllImageRefsInRelatedImagesImpact,
	"preflight-BasedOnUbi":                                   PreflightBasedOnUbiImpact,
	"preflight-BundleImageRefsAreCertified":                  PreflightBundleImageRefsAreCertifiedImpact,
	"preflight-DeployableByOLM":                              PreflightDeployableByOLMImpact,
	"preflight-FollowsRestrictedNetworkEnablementGuidelines": PreflightFollowsRestrictedNetworkEnablementGuidelinesImpact,
	"preflight-HasLicense":                                   PreflightHasLicenseImpact,
	"preflight-HasModifiedFiles":                             PreflightHasModifiedFilesImpact,
	"preflight-HasNoProhibitedLabels":                        PreflightHasNoProhibitedLabelsImpact,
	"preflight-HasNoProhibitedPackages":                      PreflightHasNoProhibitedPackagesImpact,
	"preflight-HasProhibitedContainerName":                   PreflightHasProhibitedContainerNameImpact,
	"preflight-HasRequiredLabel":                             PreflightHasRequiredLabelImpact,
	"preflight-HasUniqueTag":                                 PreflightHasUniqueTagImpact,
	"preflight-LayerCountAcceptable":                         PreflightLayerCountAcceptableImpact,
	"preflight-RequiredAnnotations":                          PreflightRequiredAnnotationsImpact,
	"preflight-RunAsNonRoot":                                 PreflightRunAsNonRootImpact,
	"preflight-ScorecardBasicSpecCheck":                      PreflightScorecardBasicSpecCheckImpact,
	"preflight-ScorecardOlmSuiteCheck":                       PreflightScorecardOlmSuiteCheckImpact,
	"preflight-SecurityContextConstraintsInCSV":              PreflightSecurityContextConstraintsInCSVImpact,
	"preflight-ValidateOperatorBundle":                       PreflightValidateOperatorBundleImpact,
}
