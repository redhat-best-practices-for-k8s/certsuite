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

// Pod Represents a Kubernetes pod with extended metadata and helper methods
//
// This structure embeds the corev1.Pod type and adds fields that track
// additional information such as service account mappings, container lists,
// network interface data, PCI device references, and flags indicating whether
// the pod is an operator or operand. It also provides boolean indicators for
// skipping certain tests. The struct’s methods offer utilities for examining
// resource guarantees, CPU isolation compliance, affinity requirements,
// SR‑IOV usage, and other security and configuration checks.
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

// NewPod Creates a Pod wrapper with network and container details
//
// The function takes a Kubernetes pod object, extracts its annotations to
// determine Multus network interfaces and PCI addresses, logs missing or empty
// annotations, and handles errors gracefully. It also inspects labels to decide
// whether to skip connectivity tests and populates the list of containers from
// the pod specification. The resulting Pod structure includes the original pod
// pointer, network interface maps, PCI information, container slice, and flags
// controlling test behavior.
func NewPod(aPod *corev1.Pod) (out Pod) {
	var err error
	out.Pod = aPod
	out.MultusNetworkInterfaces = make(map[string]CniNetworkInterface)
	annotations := aPod.GetAnnotations()
	netStatus, exists := annotations[CniNetworksStatusKey]
	if !exists || strings.TrimSpace(netStatus) == "" {
		// Be graceful: log which annotations are present when the expected one is missing/empty
		keys := make([]string, 0, len(annotations))
		for k := range annotations {
			keys = append(keys, k)
		}
		log.Info("Pod %q (namespace %q) missing or empty annotation %q. Present annotations: %v", aPod.Name, aPod.Namespace, CniNetworksStatusKey, keys)
	} else {
		out.MultusNetworkInterfaces, err = GetPodIPsPerNet(netStatus)
		if err != nil {
			log.Error("Could not get IPs for Pod %q (namespace %q), err: %v", aPod.Name, aPod.Namespace, err)
		}

		out.MultusPCIs, err = GetPciPerPod(netStatus)
		if err != nil {
			log.Error("Could not get PCIs for Pod %q (namespace %q), err: %v", aPod.Name, aPod.Namespace, err)
		}
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

// ConvertArrayPods Transforms a slice of core Kubernetes pods into provider-specific pod wrappers
//
// The function iterates over each input pod, creates a new wrapper object with
// the helper constructor, and collects pointers to these wrappers in a result
// slice. Each wrapper contains additional fields such as network interfaces,
// PCI devices, and test skip flags based on pod annotations and labels. The
// returned slice provides an enriched representation suitable for downstream
// connectivity testing.
func ConvertArrayPods(pods []*corev1.Pod) (out []*Pod) {
	for i := range pods {
		aPodWrapper := NewPod(pods[i])
		out = append(out, &aPodWrapper)
	}
	return out
}

// Pod.IsPodGuaranteed Determines if the pod meets guaranteed resource conditions
//
// The method checks whether every container in the pod has defined CPU and
// memory limits that match their requests, indicating a guaranteed QoS class.
// It delegates this logic to AreResourcesIdentical, which verifies consistency
// across all containers. The result is returned as a boolean.
func (p *Pod) IsPodGuaranteed() bool {
	return AreResourcesIdentical(p)
}

// Pod.IsPodGuaranteedWithExclusiveCPUs Determines if a pod’s CPU requests and limits are whole units and match exactly
//
// It checks that each container in the pod specifies CPU resources as whole and
// that the request equals the limit for both CPU and memory. If all containers
// satisfy these conditions, it returns true; otherwise false.
func (p *Pod) IsPodGuaranteedWithExclusiveCPUs() bool {
	return AreCPUResourcesWholeUnits(p) && AreResourcesIdentical(p)
}

// Pod.IsCPUIsolationCompliant Determines whether a pod meets CPU isolation requirements
//
// The method checks that the pod has annotations disabling both CPU and IRQ
// load balancing, and verifies a runtime class name is set. If either condition
// fails it logs a debug message and returns false; otherwise true.
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

// Pod.String Formats pod name and namespace into a readable string
//
// This method constructs a human‑readable representation of a Pod by
// combining its name and namespace. It uses formatting to produce the pattern
// "pod: <name> ns: <namespace>", which is helpful for logging or debugging
// output throughout the provider package.
func (p *Pod) String() string {
	return fmt.Sprintf("pod: %s ns: %s",
		p.Name,
		p.Namespace,
	)
}

// Pod.AffinityRequired Determines if a pod requires affinity based on its labels
//
// The method looks for the key that indicates whether affinity is required in
// the pod's label set. If present, it attempts to interpret the value as a
// boolean string; on parsing failure it logs a warning and returns false. When
// the key is absent or parsing succeeds, it returns the parsed boolean result.
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

// Pod.HasHugepages determines if any container requests or limits hugepage resources
//
// The method scans each container’s resource requests and limits for a name
// containing the substring "hugepage". If such a resource is found, it
// immediately returns true; otherwise, after all containers are checked, it
// returns false.
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

// Pod.CheckResourceHugePagesSize Verifies that all huge page resources match the specified size
//
// The method iterates over each container in a pod, checking both requested and
// limited resources for any huge page entries. If a huge page resource is found
// but its name differs from the supplied size, the function returns false
// immediately. When no mismatches are detected, it returns true.
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

// Pod.IsAffinityCompliant checks whether a pod has required affinity rules
//
// The method examines the pod's specification to determine if it contains any
// affinity configuration. If no affinity is present, or if anti‑affinity
// rules exist, or if neither pod nor node affinity are defined, it returns
// false along with an explanatory error. Otherwise it reports success by
// returning true and a nil error.
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

// Pod.IsShareProcessNamespace determines if a pod shares its process namespace
//
// The method checks the pod specification for the ShareProcessNamespace field.
// If the field exists and is set to true, it returns true; otherwise it returns
// false.
func (p *Pod) IsShareProcessNamespace() bool {
	return p.Spec.ShareProcessNamespace != nil && *p.Spec.ShareProcessNamespace
}

// Pod.ContainsIstioProxy Detects the presence of an Istio side‑car container in a pod
//
// The method scans each container defined in the pod, comparing its name
// against the predefined Istio proxy container identifier. If it finds a match,
// it immediately returns true; otherwise, after examining all containers, it
// returns false.
func (p *Pod) ContainsIstioProxy() bool {
	for _, container := range p.Containers {
		if container.Name == IstioProxyContainerName {
			return true
		}
	}
	return false
}

// Pod.CreatedByDeploymentConfig Determines if a pod originates from an OpenShift DeploymentConfig
//
// This method examines each owner reference of the pod, looking for a
// ReplicationController that itself references a DeploymentConfig. It retrieves
// replication controller objects via the Kubernetes client and checks their
// owners to find a matching deployment config name. The function returns true
// if such a relationship exists, otherwise false, along with any error
// encountered during API calls.
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

// Pod.HasNodeSelector Indicates if the pod specifies a node selector
//
// The method examines the pod's specification for a non‑empty nodeSelector
// map. It returns true when at least one key/value pair is present, meaning the
// pod has constraints on which nodes it can run. If the map is empty or nil,
// the function returns false.
func (p *Pod) HasNodeSelector() bool {
	// Checks whether or not the pod has a nodeSelector or a NodeName supplied
	return len(p.Spec.NodeSelector) != 0
}

// Pod.IsRuntimeClassNameSpecified checks whether a pod has a runtime class specified
//
// The method returns true when the pod’s specification includes a
// runtimeClassName field, indicating that a runtime class has been assigned. If
// the field is nil, it returns false, implying no runtime class is set for the
// pod.
func (p *Pod) IsRuntimeClassNameSpecified() bool {
	return p.Spec.RuntimeClassName != nil
}

// getCNCFNetworksNamesFromPodAnnotation Extracts network names from a pod's CNCF annotation
//
// The function receives the raw value of the k8s.v1.cni.cncf.io/networks
// annotation, which can be either a comma‑separated list or a JSON array of
// objects. It attempts to unmarshal the JSON; if that succeeds it collects the
// "name" field from each object. If unmarshalling fails, it falls back to
// splitting the string on commas and trimming spaces, returning all non‑empty
// names. The result is a slice of strings containing only the network
// identifiers.
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

// isNetworkAttachmentDefinitionSRIOVConfigMTUSet determines whether a SR-IOV plugin specifies an MTU
//
// The function parses the JSON network attachment definition string into a CNI
// configuration structure, verifies that it contains multiple plugins, and then
// iterates over those plugins to find one of type "sriov" with a positive MTU
// value. If such a plugin is found, it returns true; otherwise false. Errors
// are returned for malformed JSON or missing plugin list.
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

// isNetworkAttachmentDefinitionConfigTypeSRIOV checks if a CNI configuration string contains an SR-IOV plugin
//
// The function parses the JSON-formatted CNI config, handling both
// single-plugin and multi-plugin layouts. It looks for a "type" field or
// iterates through the plugins array to find an entry with type "sriov",
// returning true if found. Errors are produced for malformed JSON or unexpected
// structures.
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

// Pod.IsUsingSRIOV determines whether a pod has any SR‑IOV network interfaces
//
// The method inspects the pod’s annotations for CNCF network names, retrieves
// each corresponding NetworkAttachmentDefinition, and checks if its CNI
// configuration type is "sriov". If at least one definition matches, it returns
// true; otherwise false. Errors from annotation parsing or API calls are
// propagated to the caller.
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

// Pod.IsUsingSRIOVWithMTU determines if the pod has any SR-IOV interface configured with an MTU
//
// The method inspects the pod's annotations to find declared CNCF networks,
// then retrieves each corresponding NetworkAttachmentDefinition. For every
// network it checks whether a SriovNetwork and matching SriovNetworkNodePolicy
// exist that specify an MTU value; if so it returns true. If no such
// configuration is found, it returns false without error.
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

// sriovNetworkUsesMTU Checks whether a SriovNetwork has an MTU configured
//
// The function iterates through all provided SriovNetworks and matches one by
// name to the given NetworkAttachmentDefinition. For each match it looks for a
// SriovNetworkNodePolicy in the same namespace that shares the same
// resourceName, then examines its spec for an MTU value greater than zero. If
// such a policy is found, true is returned; otherwise false.
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

// Pod.IsUsingClusterRoleBinding Checks if a pod’s service account is linked to any cluster role binding
//
// The function receives a list of cluster role bindings and logs the pod being
// examined. It iterates through each binding, comparing the pod’s service
// account name and namespace with the subjects in the binding. If a match is
// found, it reports true along with the role reference name; otherwise it
// returns false.
//
//nolint:gocritic
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

// Pod.IsRunAsUserID Checks if the pod runs as a specific user ID
//
// The method inspects the pod's security context, returning false if it is nil
// or if no RunAsUser value is set. If a run-as-user value exists, it compares
// that value to the supplied uid and returns true when they match. This allows
// callers to verify whether the pod will execute with the given user identity.
func (p *Pod) IsRunAsUserID(uid int64) bool {
	if p.Spec.SecurityContext == nil || p.Spec.SecurityContext.RunAsUser == nil {
		return false
	}
	return *p.Spec.SecurityContext.RunAsUser == uid
}

// Pod.GetRunAsNonRootFalseContainers identifies containers violating non-root security policies
//
// This method examines each container in a pod to determine if it inherits or
// sets runAsNonRoot to false or runs as user ID zero, indicating a root
// context. It skips any containers listed in the provided map and aggregates
// those that fail the checks along with explanatory reasons. The function
// returns two slices: one of non-compliant containers and another containing
// the corresponding justification strings.
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

// Pod.GetTopOwner Retrieves the top-level owners of a pod
//
// The method returns a map keyed by owner kind, containing information about
// each top-level resource that owns the pod. It calls an internal helper to
// resolve all owner references, following chains up to the root. The result is
// returned along with any error encountered during resolution.
func (p *Pod) GetTopOwner() (topOwners map[string]podhelper.TopOwner, err error) {
	return podhelper.GetPodTopOwner(p.Namespace, p.OwnerReferences)
}

// Pod.IsAutomountServiceAccountSetOnSA Determines if a pod’s service account has automount enabled
//
// The method inspects the pod’s associated service account to see whether its
// AutomountServiceAccountToken field is set. It first validates that the
// service account map exists and contains an entry for the pod’s namespace
// and name, returning errors otherwise. If found, it returns a pointer to the
// boolean value indicating automount status along with nil error.
func (p *Pod) IsAutomountServiceAccountSetOnSA() (isSet *bool, err error) {
	if p.AllServiceAccountsMap == nil {
		return isSet, fmt.Errorf("AllServiceAccountsMap is not initialized for pod with ns: %s and name %s", p.Namespace, p.Name)
	}
	if _, ok := (*p.AllServiceAccountsMap)[p.Namespace+p.Spec.ServiceAccountName]; !ok {
		return isSet, fmt.Errorf("could not find a service account with ns: %s and name %s", p.Namespace, p.Spec.ServiceAccountName)
	}
	return (*p.AllServiceAccountsMap)[p.Namespace+p.Spec.ServiceAccountName].AutomountServiceAccountToken, nil
}
