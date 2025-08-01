package identifiers

const (
	// Default Strings
	NoDocLinkExtended = "No Doc Link - Extended"
	NoDocLinkFarEdge  = "No Doc Link - Far Edge"
	NoDocLinkTelco    = "No Doc Link - Telco"
	NoDocLink         = "No Doc Link"

	// Networking Suite
	TestICMPv4ConnectivityIdentifierDocLink             = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-ipv4-&-ipv6"
	TestNetworkPolicyDenyAllIdentifierDocLink           = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-vrfs-aka-routing-instances"
	TestReservedExtendedPartnerPortsDocLink             = NoDocLinkExtended
	TestDpdkCPUPinningExecProbeDocLink                  = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cpu-manager-pinning"
	TestRestartOnRebootLabelOnPodsUsingSRIOVDocLink     = NoDocLinkFarEdge
	TestNetworkAttachmentDefinitionSRIOVUsingMTUDocLink = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-multus-sr-iov---macvlan"
	TestLimitedUseOfExecProbesIdentifierDocLink         = NoDocLinkFarEdge
	TestICMPv6ConnectivityIdentifierDocLink             = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-ipv4-&-ipv6"
	TestICMPv4ConnectivityMultusIdentifierDocLink       = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations"
	TestICMPv6ConnectivityMultusIdentifierDocLink       = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations"
	TestServiceDualStackIdentifierDocLink               = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-ipv4-&-ipv6"
	TestUndeclaredContainerPortsUsageDocLink            = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-requirements-cnf-reqs"
	TestOCPReservedPortsUsageDocLink                    = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-ports-reserved-by-openshift"

	// Access Control Suite
	Test1337UIDIdentifierDocLink                             = NoDocLinkExtended
	TestNetAdminIdentifierDocLink                            = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-net_admin"
	TestSysAdminIdentifierDocLink                            = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-avoid-sys_admin"
	TestIpcLockIdentifierDocLink                             = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-ipc_lock"
	TestNetRawIdentifierDocLink                              = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-user-plane-cnfs"
	TestBpfIdentifierDocLink                                 = NoDocLinkTelco
	TestSecConNonRootUserIdentifierDocLink                   = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security"
	TestSecContextIdentifierDocLink                          = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-linux-capabilities"
	TestSecConPrivilegeEscalationDocLink                     = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security"
	TestContainerHostPortDocLink                             = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-avoid-accessing-resource-on-host"
	TestContainerHostNetworkDocLink                          = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security"
	TestPodHostNetworkDocLink                                = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security"
	TestPodHostPathDocLink                                   = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security"
	TestPodHostIPCDocLink                                    = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security"
	TestPodHostPIDDocLink                                    = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security"
	TestNamespaceBestPracticesIdentifierDocLink              = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-requirements-cnf-reqs"
	TestPodClusterRoleBindingsBestPracticesIdentifierDocLink = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-security-and-role-based-access-control"
	TestPodRoleBindingsBestPracticesIdentifierDocLink        = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-security-and-role-based-access-control"
	TestPodServiceAccountBestPracticesIdentifierDocLink      = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-scc-permissions-for-an-application"
	TestPodAutomountServiceAccountIdentifierDocLink          = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-automount-services-for-pods"
	TestServicesDoNotUseNodeportsIdentifierDocLink           = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security"
	TestUnalteredBaseImageIdentifierDocLink                  = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-image-standards"
	TestOneProcessPerContainerIdentifierDocLink              = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-one-process-per-container"
	TestSYSNiceRealtimeCapabilityIdentifierDocLink           = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-sys_nice"
	TestSysPtraceCapabilityIdentifierDocLink                 = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-sys_ptrace"
	TestPodRequestsIdentifierDocLink                         = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-requests-limits"
	TestNamespaceResourceQuotaIdentifierDocLink              = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-memory-allocation"
	TestNoSSHDaemonsAllowedIdentifierDocLink                 = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-pod-interaction-and-configuration"

	// Affiliated Certification Suite
	TestHelmVersionIdentifierDocLink                = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-helm"
	TestContainerIsCertifiedDigestIdentifierDocLink = "https://docs.redhat.com/en/documentation/red_hat_software_certification/2025/html/red_hat_software_certification_workflow_guide/index"
	TestContainerIsCertifiedIdentifierDocLink       = "https://redhat-connect.gitbook.io/partner-guide-for-red-hat-openshift-and-container/certify-your-application/overview"
	TestHelmIsCertifiedIdentifierDocLink            = "https://docs.redhat.com/en/documentation/red_hat_software_certification/2025/html/red_hat_software_certification_workflow_guide/index"

	// Platform Alteration Suite
	TestPodHugePages2MDocLink                       = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-huge-pages"
	TestPodHugePages1GDocLink                       = NoDocLinkFarEdge
	TestHugepagesNotManuallyManipulatedDocLink      = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-huge-pages"
	TestNonTaintedNodeKernelsIdentifierDocLink      = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations"
	TestUnalteredStartupBootParamsIdentifierDocLink = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-host-os"
	TestSysctlConfigsIdentifierDocLink              = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security"
	TestServiceMeshIdentifierDocLink                = NoDocLinkExtended
	TestHyperThreadEnableDocLink                    = NoDocLinkExtended

	TestOCPLifecycleIdentifierDocLink        = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-k8s"
	TestNodeOperatingSystemIdentifierDocLink = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-host-os"
	TestIsRedHatReleaseIdentifierDocLink     = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-security"
	TestIsSELinuxEnforcingIdentifierDocLink  = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-pod-security"
	TestClusterOperatorHealthDocLink         = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements"

	// Lifecycle Suite
	TestAffinityRequiredPodsDocLink                    = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations"
	TestStorageProvisionerDocLink                      = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-local-storage"
	TestContainerPostStartIdentifierDocLink            = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cloud-native-design-best-practices"
	TestContainerPrestopIdentifierDocLink              = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cloud-native-design-best-practices"
	TestPodNodeSelectorAndAffinityBestPracticesDocLink = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations"
	TestPodHighAvailabilityBestPracticesDocLink        = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations"
	TestPodDeploymentBestPracticesIdentifierDocLink    = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-no-naked-pods"
	TestDeploymentScalingIdentifierDocLink             = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations"
	TestStatefulSetScalingIdentifierDocLink            = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations"
	TestImagePullPolicyIdentifierDocLink               = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-use-imagepullpolicy:-ifnotpresent"
	TestPodRecreationIdentifierDocLink                 = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-upgrade-expectations"
	TestLivenessProbeIdentifierDocLink                 = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-liveness-readiness-and-startup-probes"
	TestReadinessProbeIdentifierDocLink                = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-liveness-readiness-and-startup-probes"
	TestStartupProbeIdentifierDocLink                  = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-liveness-readiness-and-startup-probes"
	//nolint:gosec
	TestPodTolerationBypassIdentifierDocLink           = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cpu-manager-pinning"
	TestPersistentVolumeReclaimPolicyIdentifierDocLink = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-csi"
	TestCPUIsolationIdentifierDocLink                  = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cpu-isolation"
	TestCrdScalingIdentifierDocLink                    = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-high-level-cnf-expectations"

	// Performance Test Suite
	TestExclusiveCPUPoolIdentifierDocLink       = NoDocLinkFarEdge
	TestSharedCPUPoolSchedulingPolicyDocLink    = NoDocLinkFarEdge
	TestExclusiveCPUPoolSchedulingPolicyDocLink = NoDocLinkFarEdge
	TestIsolatedCPUPoolSchedulingPolicyDocLink  = NoDocLinkFarEdge
	TestRtAppNoExecProbesDocLink                = NoDocLinkFarEdge

	// Operator Test Suite
	DocOperatorRequirement                                                  = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements"
	TestOperatorInstallStatusSucceededIdentifierDocLink                     = DocOperatorRequirement
	TestOperatorNoPrivilegesDocLink                                         = DocOperatorRequirement
	TestOperatorIsCertifiedIdentifierDocLink                                = DocOperatorRequirement
	TestOperatorIsInstalledViaOLMIdentifierDocLink                          = DocOperatorRequirement
	TestSingleOrMultiNamespacedOperatorInstallationInTenantNamespaceDocLink = DocOperatorRequirement
	TestOperatorHasSemanticVersioningIdentifierDocLink                      = DocOperatorRequirement
	TestOperatorCrdSchemaIdentifierDocLink                                  = DocOperatorRequirement
	TestOperatorCrdVersioningIdentifierDocLink                              = DocOperatorRequirement
	TestOperatorSingleCrdOwnerIdentifierDocLink                             = DocOperatorRequirement
	TestOperatorRunAsUserIDDocLink                                          = DocOperatorRequirement
	TestOperatorRunAsNonRootDocLink                                         = DocOperatorRequirement
	TestOperatorAutomountTokensDocLink                                      = DocOperatorRequirement
	TestOperatorReadOnlyFilesystemDocLink                                   = DocOperatorRequirement
	TestOperatorPodsNoHugepagesDocLink                                      = DocOperatorRequirement
	TestOperatorCatalogSourceBundleCountIdentifierDocLink                   = DocOperatorRequirement
	TestOperatorOlmSkipRangeDocLink                                         = DocOperatorRequirement
	TestMultipleSameOperatorsIdentifierDocLink                              = DocOperatorRequirement

	// Observability Test Suite
	TestLoggingIdentifierDocLink                            = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-logging"
	TestTerminationMessagePolicyIdentifierDocLink           = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-pod-exit-status"
	TestCrdsStatusSubresourceIdentifierDocLink              = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-cnf-operator-requirements"
	TestPodDisruptionBudgetIdentifierDocLink                = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-upgrade-expectations"
	TestAPICompatibilityWithNextOCPReleaseIdentifierDocLink = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-k8s-api-versions"

	// Manageability Test Suite
	TestContainersImageTagDocLink      = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-image-tagging"
	TestContainerPortNameFormatDocLink = "https://redhat-best-practices-for-k8s.github.io/guide/#k8s-best-practices-requirements-cnf-reqs"
)
