// Copyright (C) 2022-2024 Red Hat, Inc.
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

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/podhelper"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	HugePages2Mi            = "hugepages-2Mi"
	HugePages1Gi            = "hugepages-1Gi"
	hugePages               = "hugepages"
	replicationController   = "ReplicationController"
	deploymentConfig        = "DeploymentConfig"
	IstioProxyContainerName = "istio-proxy"
)

// Pod represents a Kubernetes pod with additional metadata used by the certsuite provider.
//
// It embeds the corev1.Pod type and augments it with fields that track
// service account mappings, container lists, operator/operand flags,
// Multus network interfaces, PCI device references, and test‑skipping
// flags. These extras allow the provider to make decisions about
// security, resource allocation, networking, and test coverage when
// evaluating pod compliance.
type Pod struct {
	*corev1.Pod
	AllServiceAccountsMap   *map[string]*corev1.ServiceAccount
	Containers              []*Container
	MultusNetworkInterfaces map[string]CniNetworkInterface
	MultusPCIs              []string
	SkipNetTests            bool
	SkipMultusNetTests      bool
	IsOperator              bool
	IsOperand               bool
}

// NewPod creates a Pod wrapper from a corev1.Pod pointer.
//
// It extracts IP addresses, annotations, labels, PCI devices, and container information
// to populate the custom Pod structure used by the provider. The function returns
// the constructed Pod value for further processing or inspection.
func NewPod(aPod *corev1.Pod) (out Pod) {
	var err error
	out.Pod = aPod
	out.MultusNetworkInterfaces = make(map[string]CniNetworkInterface)
	out.MultusNetworkInterfaces, err = GetPodIPsPerNet(aPod.GetAnnotations()[CniNetworksStatusKey])
	if err != nil {
		log.Error("Could not get IPs for Pod %q (namespace %q), err: %v", aPod.Name, aPod.Namespace, err)
	}

	out.MultusPCIs, err = GetPciPerPod(aPod.GetAnnotations()[CniNetworksStatusKey])
	if err != nil {
		log.Error("Could not get PCIs for Pod %q (namespace %q), err: %v", aPod.Name, aPod.Namespace, err)
	}

	if _, ok := aPod.GetLabels()[skipConnectivityTestsLabel]; ok {
		out.SkipNetTests = true
	}
	if _, ok := aPod.GetLabels()[skipMultusConnectivityTestsLabel]; ok {
		out.SkipMultusNetTests = true
	}
	out.Containers = append(out.Containers, getPodContainers(aPod, false)...)
	return out
}

// ConvertArrayPods converts a slice of corev1.Pod pointers into a slice of Pod structs used by the provider.
//
// It iterates over each Kubernetes pod, creates a new internal Pod representation using NewPod,
// and appends it to the resulting slice. The function returns the populated slice of *Pod.
func ConvertArrayPods(pods []*corev1.Pod) (out []*Pod) {
	for i := range pods {
		aPodWrapper := NewPod(pods[i])
		out = append(out, &aPodWrapper)
	}
	return out
}

// IsPodGuaranteed reports whether the pod is guaranteed.
//
// It returns true if all containers in the pod request and limit the same
// amount of resources, meaning the pod satisfies the Kubernetes Guaranteed
// QoS class rules. The function relies on AreResourcesIdentical to compare
// container resource requests and limits. If any container differs or a
// non‑resource field is missing, it returns false.
func (p *Pod) IsPodGuaranteed() bool {
	return AreResourcesIdentical(p)
}

// IsPodGuaranteedWithExclusiveCPUs reports whether the pod has a guaranteed QoS class and requests whole CPU units exclusively for its containers.
//
// It examines each container's resource requests to determine if they
// are identical across all containers and represent whole CPU units.
// The function returns true when the pod meets these criteria,
// indicating that it is guaranteed with exclusive CPUs.
func (p *Pod) IsPodGuaranteedWithExclusiveCPUs() bool {
	return AreCPUResourcesWholeUnits(p) && AreResourcesIdentical(p)
}

// IsCPUIsolationCompliant checks whether the pod is compliant with CPU isolation requirements.
//
// It verifies that load balancing is disabled and that a runtime class name is specified.
// The function logs debug information during its checks and returns true if all conditions are met, otherwise false.
func (p *Pod) IsCPUIsolationCompliant() bool {
	isCPUIsolated := true

	if !LoadBalancingDisabled(p) {
		log.Debug("Pod %q has been found to not have annotations set correctly for CPU isolation.", p)
		isCPUIsolated = false
	}

	if !p.IsRuntimeClassNameSpecified() {
		log.Debug("Pod %q has been found to not have runtimeClassName specified.", p)
		isCPUIsolated = false
	}

	return isCPUIsolated
}

// String returns a human readable representation of the Pod.
//
// It formats the pod's basic identifying information into a single string,
// typically including its name and namespace, using fmt.Sprintf internally.
func (p *Pod) String() string {
	return fmt.Sprintf("pod: %s ns: %s",
		p.Name,
		p.Namespace,
	)
}

// AffinityRequired determines if pod affinity is required based on annotations.
//
// It checks the pod's annotations for a specific key indicating whether node
// affinity should be enforced. The value is parsed as a boolean; if parsing fails,
// a warning is logged and false is returned.
// This method returns true when affinity is explicitly requested, otherwise false.
func (p *Pod) AffinityRequired() bool {
	if val, ok := p.Labels[AffinityRequiredKey]; ok {
		result, err := strconv.ParseBool(val)
		if err != nil {
			log.Warn("failure to parse bool %v", val)
			return false
		}
		return result
	}
	return false
}

// HasHugepages reports whether any container in the pod requests a hugepage resource.
//
// It iterates over all containers in the Pod and checks each resource request
// key for the substring "hugepage". If at least one match is found it returns true,
// otherwise it returns false. This function does not modify the Pod.
func (p *Pod) HasHugepages() bool {
	for _, cut := range p.Containers {
		for name := range cut.Resources.Requests {
			if strings.Contains(name.String(), hugePages) {
				return true
			}
		}
		for _, name := range cut.Resources.Limits {
			if strings.Contains(name.String(), hugePages) {
				return true
			}
		}
	}
	return false
}

// CheckResourceHugePagesSize verifies that the pod requests match expected huge page sizes.
//
// It receives a string containing comma‑separated requested memory sizes and checks
// each against the predefined huge page constants (1Gi, 2Mi). The function returns true
// if all requested sizes are valid huge page values; otherwise it returns false. This
// helps ensure that pods declare correct huge page allocations for performance
// or compliance requirements.
func (p *Pod) CheckResourceHugePagesSize(size string) bool {
	for _, cut := range p.Containers {
		// Resources must be specified
		if len(cut.Resources.Requests) == 0 || len(cut.Resources.Limits) == 0 {
			continue
		}
		for name := range cut.Resources.Requests {
			if strings.Contains(name.String(), hugePages) && name.String() != size {
				return false
			}
		}
		for name := range cut.Resources.Limits {
			if strings.Contains(name.String(), hugePages) && name.String() != size {
				return false
			}
		}
	}
	return true
}

// IsAffinityCompliant checks if the pod satisfies node affinity constraints.
//
// It evaluates the pod’s nodeSelector, requiredDuringSchedulingIgnoredDuringExecution,
// and preferredDuringSchedulingIgnoredDuringExecution rules against the
// current cluster node labels. The function returns true if all required
// affinity rules match a node in the cluster; otherwise it returns false.
// An error is returned if any rule cannot be parsed or evaluated.
func (p *Pod) IsAffinityCompliant() (bool, error) {
	if p.Spec.Affinity == nil {
		return false, fmt.Errorf("%s has been found with an AffinityRequired flag but is missing corresponding affinity rules", p.String())
	}
	if p.Spec.Affinity.PodAntiAffinity != nil {
		return false, fmt.Errorf("%s has been found with an AffinityRequired flag but has anti-affinity rules", p.String())
	}
	if p.Spec.Affinity.PodAffinity == nil && p.Spec.Affinity.NodeAffinity == nil {
		return false, fmt.Errorf("%s has been found with an AffinityRequired flag but is missing corresponding pod/node affinity rules", p.String())
	}
	return true, nil
}

// IsShareProcessNamespace reports whether the Pod shares its process namespace with containers.
//
// It returns true if the pod is configured to share a single process namespace
// across all of its containers, allowing them to see each other’s processes.
// If the pod does not enable this feature or has no containers,
// it returns false.
func (p *Pod) IsShareProcessNamespace() bool {
	return p.Spec.ShareProcessNamespace != nil && *p.Spec.ShareProcessNamespace
}

// ContainsIstioProxy reports whether the Pod includes an Istio sidecar.
//
// It scans the Pod's containers and returns true if a container named
// according to IstioProxyContainerName is present, indicating that the pod
// has been injected with the Istio proxy sidecar. If no such container exists,
// it returns false.
func (p *Pod) ContainsIstioProxy() bool {
	for _, container := range p.Containers {
		if container.Name == IstioProxyContainerName {
			return true
		}
	}
	return false
}

// CreatedByDeploymentConfig reports whether the pod was created by a DeploymentConfig.
//
// It examines the pod's owner references to determine if any reference is a
// DeploymentConfig object. If found, it returns true with no error; otherwise it
// returns false and an error indicating that the pod has no DeploymentConfig
// owner or an issue occurred while retrieving ownership information.
func (p *Pod) CreatedByDeploymentConfig() (bool, error) {
	oc := clientsholder.GetClientsHolder()
	for _, podOwner := range p.GetOwnerReferences() {
		if podOwner.Kind == replicationController {
			replicationControllers, err := oc.K8sClient.CoreV1().ReplicationControllers(p.Namespace).Get(context.TODO(), podOwner.Name, metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			for _, rcOwner := range replicationControllers.GetOwnerReferences() {
				if rcOwner.Name == podOwner.Name && rcOwner.Kind == deploymentConfig {
					return true, err
				}
			}
		}
	}
	return false, nil
}

// HasNodeSelector reports whether the pod contains a node selector.
//
// It inspects the pod's specification and returns true if any key/value
// pairs are defined in the NodeSelector map, indicating that the pod is
// constrained to run on nodes matching those labels. If the map is nil or
// empty, it returns false.
func (p *Pod) HasNodeSelector() bool {
	// Checks whether or not the pod has a nodeSelector or a NodeName supplied
	return len(p.Spec.NodeSelector) != 0
}

// IsRuntimeClassNameSpecified reports whether the pod has a runtime class name set.
//
// It checks the pod's specification to determine if a RuntimeClassName field is present
// and non-empty, indicating that the pod should run with a specific runtime configuration.
// The method returns true when a runtime class name is specified; otherwise it returns false.
func (p *Pod) IsRuntimeClassNameSpecified() bool {
	return p.Spec.RuntimeClassName != nil
}

// getCNCFNetworksNamesFromPodAnnotation parses the CNCF networks annotation on a pod and returns only the network names.
//
// It accepts a single string argument containing the value of the
// k8s.v1.cni.cncf.io/networks annotation, which may be either a comma‑separated list of names or a JSON array of objects.
// The function extracts each network name, trims whitespace, and returns them as a slice of strings. If the input cannot be parsed,
// an empty slice is returned.
func getCNCFNetworksNamesFromPodAnnotation(networksAnnotation string) []string {
	// Each CNCF network has many more fields, but here we only need to unmarshal the name.
	// See https://github.com/k8snetworkplumbingwg/multus-cni/blob/e692127d19623c8bdfc4d391224ea542658b584c/pkg/types/types.go#L127
	type CNCFNetwork struct {
		Name string `json:"name"`
	}

	networkObjects := []CNCFNetwork{}
	networkNames := []string{}

	// Let's start trying to unmarshal a json array of objects.
	// We will not care about bad-formatted/invalid annotation value. If that's the case,
	// the pod wouldn't have been deployed or wouldn't be in running state.
	if err := json.Unmarshal([]byte(networksAnnotation), &networkObjects); err == nil {
		for _, network := range networkObjects {
			networkNames = append(networkNames, network.Name)
		}
		return networkNames
	}

	// If the previous unmarshalling didn't work, let's try with parsing the comma separated names list.
	networks := strings.TrimSpace(networksAnnotation)

	// First, avoid empty strings (unlikely).
	if networks == "" {
		return []string{}
	}

	for _, networkName := range strings.Split(networks, ",") {
		networkNames = append(networkNames, strings.TrimSpace(networkName))
	}
	return networkNames
}

// isNetworkAttachmentDefinitionSRIOVConfigMTUSet is a helper function to check whether a CNI config
// string has any config for MTU for SRIOV configs only

/*
	{
		"cniVersion": "0.4.0",
		"name": "vlan-100",
		"plugins": [
			{
				"type": "sriov",
				"master": "ext0",
				"mtu": 1500,
				"vlanId": 100,
				"linkInContainer": true,
				"ipam": {"type": "whereabouts", "ipRanges": [{"range": "1.1.1.0/24"}]}
			}
		]
	}
*/

// isNetworkAttachmentDefinitionSRIOVConfigMTUSet checks whether a NetworkAttachmentDefinition's SR-IOV plugin has an MTU set and returns the result or an error.
//
// isNetworkAttachmentDefinitionSRIOVConfigMTUSet determines if the given NetworkAttachmentDefinition JSON string
// contains an SR‑IOV plugin entry with the "mtu" field defined. It unmarshals the JSON, searches for a plugin of type
// "sriov", and verifies that the "mtu" key is present. If found, it returns true; otherwise false.
// The function reports detailed errors if JSON parsing fails or if no SR‑IOV plugin is detected.
func isNetworkAttachmentDefinitionSRIOVConfigMTUSet(nadConfig string) (bool, error) {
	const (
		typeSriov = "sriov"
	)

	type CNIConfig struct {
		CniVersion string  `json:"cniVersion"`
		Name       string  `json:"name"`
		Type       *string `json:"type,omitempty"`
		Plugins    *[]struct {
			Type string `json:"type"`
			MTU  int    `json:"mtu"`
		} `json:"plugins,omitempty"`
	}

	cniConfig := CNIConfig{}
	if err := json.Unmarshal([]byte(nadConfig), &cniConfig); err != nil {
		return false, fmt.Errorf("failed to unmarshal cni config %s: %v", nadConfig, err)
	}

	if cniConfig.Plugins == nil {
		return false, fmt.Errorf("invalid multi-plugins cni config: %s", nadConfig)
	}

	log.Debug("CNI plugins: %+v", *cniConfig.Plugins)
	for i := range *cniConfig.Plugins {
		plugin := (*cniConfig.Plugins)[i]
		if plugin.Type == typeSriov && plugin.MTU > 0 {
			return true, nil
		}
	}

	// No sriov plugin type found.
	return false, nil
}

// isNetworkAttachmentDefinitionConfigTypeSRIOV checks whether a CNI configuration string contains any SR‑I/O V plugin configuration.
//
// It accepts the raw JSON configuration of a CNI network attachment definition and returns
// a boolean indicating if an SR‑I/O V plugin is present, along with an error if the
// configuration cannot be parsed. The function handles both single‑plugin and multi‑plugin
// layouts, inspecting either the top‑level "type" field or each element in the "plugins"
// array for a value of "sriov".
func isNetworkAttachmentDefinitionConfigTypeSRIOV(nadConfig string) (bool, error) {
	const (
		typeSriov = "sriov"
	)

	type CNIConfig struct {
		CniVersion string  `json:"cniVersion"`
		Name       string  `json:"name"`
		Type       *string `json:"type,omitempty"`
		Plugins    *[]struct {
			Type string `json:"type"`
		} `json:"plugins,omitempty"`
	}

	cniConfig := CNIConfig{}
	if err := json.Unmarshal([]byte(nadConfig), &cniConfig); err != nil {
		return false, fmt.Errorf("failed to unmarshal cni config %s: %v", nadConfig, err)
	}

	// If type is found, it's a single plugin CNI config.
	if cniConfig.Type != nil {
		log.Debug("Single plugin config type found: %+v, type=%s", cniConfig, *cniConfig.Type)
		return *cniConfig.Type == typeSriov, nil
	}

	if cniConfig.Plugins == nil {
		return false, fmt.Errorf("invalid multi-plugins cni config: %s", nadConfig)
	}

	log.Debug("CNI plugins: %+v", *cniConfig.Plugins)
	for i := range *cniConfig.Plugins {
		plugin := (*cniConfig.Plugins)[i]
		if plugin.Type == typeSriov {
			return true, nil
		}
	}

	// No sriov plugin type found.
	return false, nil
}

// IsUsingSRIOV determines whether the pod uses any SR-IOV network interfaces.
//
// It retrieves the list of network names from the CNFC annotation on the pod,
// then fetches each corresponding NetworkAttachmentDefinition (NAD) and
// checks if its configuration type is SR-IOV. If at least one interface is
// SR‑IOV, it returns true; otherwise false. An error is returned if any
// step of the lookup or validation fails.
func (p *Pod) IsUsingSRIOV() (bool, error) {
	const (
		cncfNetworksAnnotation = "k8s.v1.cni.cncf.io/networks"
	)

	cncfNetworks, exist := p.Annotations[cncfNetworksAnnotation]
	if !exist {
		return false, nil
	}

	// Get all CNCF network names
	cncfNetworkNames := getCNCFNetworksNamesFromPodAnnotation(cncfNetworks)

	// For each CNCF network, get its network attachment definition and check
	// whether its config's type is "sriov"
	oc := clientsholder.GetClientsHolder()

	for _, networkName := range cncfNetworkNames {
		log.Debug("%s: Reviewing network-attachment definition %q", p, networkName)
		nad, err := oc.CNCFNetworkingClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(p.Namespace).Get(context.TODO(), networkName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to get NetworkAttachment %s: %v", networkName, err)
		}

		isSRIOV, err := isNetworkAttachmentDefinitionConfigTypeSRIOV(nad.Spec.Config)
		if err != nil {
			return false, fmt.Errorf("failed to know if network-attachment %s is sriov: %v", networkName, err)
		}

		log.Debug("%s: NAD config: %s", p, nad.Spec.Config)
		if isSRIOV {
			return true, nil
		}
	}

	return false, nil
}

// IsUsingSRIOVWithMTU reports whether the pod has any SR-IOV network interface configured with an MTU value.
//
// It examines all CNI networks attached to the pod, retrieves their definitions,
// and checks each SR‑IOV attachment for a non‑zero MTU. The function returns true
// if at least one such interface is found, along with any error that occurred while
// querying the network definitions or parsing annotations.
func (p *Pod) IsUsingSRIOVWithMTU() (bool, error) {
	const (
		cncfNetworksAnnotation = "k8s.v1.cni.cncf.io/networks"
	)

	cncfNetworks, exist := p.Annotations[cncfNetworksAnnotation]
	if !exist {
		return false, nil
	}

	// Get all CNCF network names
	cncfNetworkNames := getCNCFNetworksNamesFromPodAnnotation(cncfNetworks)

	// For each CNCF network, get its network attachment definition and check
	// whether its config's type is "sriov"

	oc := clientsholder.GetClientsHolder()
	for _, networkName := range cncfNetworkNames {
		log.Debug("%s: Reviewing network-attachment definition %q", p, networkName)
		nad, err := oc.CNCFNetworkingClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(
			p.Namespace).Get(context.TODO(), networkName, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to get NetworkAttachment %s: %v", networkName, err)
		}

		// If the network-status annotation is not set, let's check the SriovNetwork/SriovNetworkNodePolicy CRs
		// to see if the MTU is set there.
		log.Debug("Number of SriovNetworks: %d", len(env.AllSriovNetworks))
		log.Debug("Number of SriovNetworkNodePolicies: %d", len(env.AllSriovNetworkNodePolicies))
		if sriovNetworkUsesMTU(env.AllSriovNetworks, env.AllSriovNetworkNodePolicies, nad.Name) {
			return true, nil
		}
	}

	return false, nil
}

// sriovNetworkUsesMTU(interfaces []unstructured.Unstructured, networks []unstructured.Unstructured, ifaceName string) bool
//
// sriovNetworkUsesMTU determines if the specified SR‑IOV interface should use a custom MTU.
//
// It receives slices of unstructured Kubernetes objects representing node network interfaces
// and network definitions. The ifaceName parameter identifies the specific interface to inspect.
// The function looks up the interface, retrieves its associated SR‑IOV network name,
// finds that network definition, and checks whether an MTU value is defined for it.
// If a custom MTU is present, the function returns true; otherwise false.
func sriovNetworkUsesMTU(sriovNetworks, sriovNetworkNodePolicies []unstructured.Unstructured, nadName string) bool {
	for _, sriovNetwork := range sriovNetworks {
		networkName := sriovNetwork.GetName()
		log.Debug("Checking SriovNetwork %s", networkName)
		if networkName == nadName {
			log.Debug("SriovNetwork %s found to match the NAD name %s", networkName, nadName)

			// Get the ResourceName from the SriovNetwork spec
			spec, found, err := unstructured.NestedMap(sriovNetwork.Object, "spec")
			if !found || err != nil {
				log.Debug("Failed to get spec from SriovNetwork %s: %v", networkName, err)
				continue
			}

			resourceName, found, err := unstructured.NestedString(spec, "resourceName")
			if !found || err != nil {
				log.Debug("Failed to get resourceName from SriovNetwork %s: %v", networkName, err)
				continue
			}

			for _, nodePolicy := range sriovNetworkNodePolicies {
				policyNamespace := nodePolicy.GetNamespace()
				networkNamespace := sriovNetwork.GetNamespace()

				log.Debug("Checking SriovNetworkNodePolicy in namespace %s", policyNamespace)
				if policyNamespace == networkNamespace {
					// Get the ResourceName and MTU from the SriovNetworkNodePolicy spec
					policySpec, found, err := unstructured.NestedMap(nodePolicy.Object, "spec")
					if !found || err != nil {
						log.Debug("Failed to get spec from SriovNetworkNodePolicy: %v", err)
						continue
					}

					policyResourceName, found, err := unstructured.NestedString(policySpec, "resourceName")
					if !found || err != nil {
						log.Debug("Failed to get resourceName from SriovNetworkNodePolicy: %v", err)
						continue
					}

					if policyResourceName == resourceName {
						mtu, found, err := unstructured.NestedInt64(policySpec, "mtu")
						if found && err == nil && mtu > 0 {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// IsUsingClusterRoleBinding checks whether the pod uses a cluster‑wide role binding.
//
// It examines a slice of ClusterRoleBindings and determines if any of them
// grant permissions to the service account used by the pod. The function
// returns a boolean indicating usage, a string describing the found
// binding (or an empty string), and an error if the check fails.
func (p *Pod) IsUsingClusterRoleBinding(clusterRoleBindings []rbacv1.ClusterRoleBinding,
	logger *log.Logger) (bool, string, error) {
	// This function accepts a list of clusterRoleBindings and checks to see if the pod's service account is
	// tied to any of them.  If it is, then it returns true, otherwise it returns false.
	logger.Info("Pod %q is using service account %q", p, p.Spec.ServiceAccountName)

	// Loop through the service accounts in the namespace, looking for a match between the pod serviceAccountName and
	// the service account name.  If there is a match, check to make sure that the SA is not a 'subject' of the cluster
	// role bindings.
	for crbIndex := range clusterRoleBindings {
		for _, subject := range clusterRoleBindings[crbIndex].Subjects {
			if subject.Kind == rbacv1.ServiceAccountKind &&
				subject.Name == p.Spec.ServiceAccountName && subject.Namespace == p.Namespace {
				logger.Error("Pod %q has service account %q that is tied to cluster role binding %q", p.Name, p.Spec.ServiceAccountName, clusterRoleBindings[crbIndex].Name)
				return true, clusterRoleBindings[crbIndex].RoleRef.Name, nil
			}
		}
	}

	return false, "", nil
}

// IsRunAsUserID reports whether the Pod has a security context that runs as the specified UID.
//
// It examines the Pod's containers, initContainers, and any pod-level security
// contexts to determine if any of them specify a RunAsUser value equal to the
// provided uid. If at least one container or the pod itself uses that uid,
// the function returns true; otherwise it returns false. This check is used
// to enforce UID constraints in tests.
func (p *Pod) IsRunAsUserID(uid int64) bool {
	if p.Spec.SecurityContext == nil || p.Spec.SecurityContext.RunAsUser == nil {
		return false
	}
	return *p.Spec.SecurityContext.RunAsUser == uid
}

// GetRunAsNonRootFalseContainers returns containers with insecure security settings.
//
// It examines the pod’s default runAsNonRoot and runAsUser values,
// then checks each container to see if it has runAsNonRoot set to false
// or runAsUser set to zero (both considered insecure). The function
// returns a slice of pointers to those containers and a slice of their
// names. If no such containers are found, the slices are empty.
func (p *Pod) GetRunAsNonRootFalseContainers(knownContainersToSkip map[string]bool) (nonCompliantContainers []*Container, nonComplianceReasons []string) {
	// Check pod-level security context this will be set by default for containers
	// If not already configured at the container level
	var podRunAsNonRoot *bool
	if p.Spec.SecurityContext != nil && p.Spec.SecurityContext.RunAsNonRoot != nil {
		podRunAsNonRoot = p.Spec.SecurityContext.RunAsNonRoot
	}

	var podRunAsUserID *int64
	if p.Spec.SecurityContext != nil && p.Spec.SecurityContext.RunAsUser != nil {
		podRunAsUserID = p.Spec.SecurityContext.RunAsUser
	}

	// Check each container for the RunAsNonRoot parameter.
	// If it is not present, the pod value applies
	for _, cut := range p.Containers {
		if knownContainersToSkip[cut.Name] {
			continue
		}

		isRunAsNonRoot, isRunAsNonRootReason := cut.IsContainerRunAsNonRoot(podRunAsNonRoot)
		isRunAsNonRootUserID, isRunAsNonRootUserIDReason := cut.IsContainerRunAsNonRootUserID(podRunAsUserID)

		if isRunAsNonRoot || isRunAsNonRootUserID {
			continue
		}

		nonCompliantContainers = append(nonCompliantContainers, cut)
		nonComplianceReasons = append(nonComplianceReasons, isRunAsNonRootReason+", "+isRunAsNonRootUserIDReason)
	}

	return nonCompliantContainers, nonComplianceReasons
}

// GetTopOwner returns a map of pod top owners and an error if any.
//
// It calls the internal GetPodTopOwner function to gather the
// highest‑level owner references for each pod in the provider.
// The returned map keys are owner identifiers, and the values
// contain details such as kind, name, and UID. An error is
// returned only when the underlying retrieval fails.
func (p *Pod) GetTopOwner() (topOwners map[string]podhelper.TopOwner, err error) {
	return podhelper.GetPodTopOwner(p.Namespace, p.OwnerReferences)
}

// IsAutomountServiceAccountSetOnSA checks if the AutomountServiceAccountToken field is set on the pod's ServiceAccount.
//
// It returns a pointer to a bool indicating whether the AutomountServiceAccountToken field is set, and an error if any occurred during the operation.
func (p *Pod) IsAutomountServiceAccountSetOnSA() (isSet *bool, err error) {
	if p.AllServiceAccountsMap == nil {
		return isSet, fmt.Errorf("AllServiceAccountsMap is not initialized for pod with ns: %s and name %s", p.Namespace, p.Name)
	}
	if _, ok := (*p.AllServiceAccountsMap)[p.Namespace+p.Spec.ServiceAccountName]; !ok {
		return isSet, fmt.Errorf("could not find a service account with ns: %s and name %s", p.Namespace, p.Spec.ServiceAccountName)
	}
	return (*p.AllServiceAccountsMap)[p.Namespace+p.Spec.ServiceAccountName].AutomountServiceAccountToken, nil
}
