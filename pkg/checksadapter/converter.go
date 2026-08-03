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

package checksadapter

import (
	"context"
	"sync"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/checks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Cached expensive resources -- computed once, reused across all checks.
var (
	cachedCertValidator     checks.CertificationValidator
	cachedCertValidatorOnce sync.Once
	cachedCRInstances       map[string]map[string][]string
	cachedCRInstancesOnce   sync.Once
)

// ConvertToDiscoveredResources converts provider.TestEnvironment to checks.DiscoveredResources.
func ConvertToDiscoveredResources(env *provider.TestEnvironment) *checks.DiscoveredResources {
	resources := &checks.DiscoveredResources{
		Namespaces:    env.Namespaces,
		ProbePods:     env.ProbePods,
		ProbeExecutor: &ProbeExecutorAdapter{},
	}

	convertWorkloads(resources, env)
	convertRBACAndPolicies(resources, env)
	convertClusterResources(resources, env)
	convertCertificationResources(resources, env)

	// Use cached CR instances (expensive dynamic client calls)
	cachedCRInstancesOnce.Do(func() {
		cachedCRInstances = buildCRInstances(resources)
	})
	resources.CRInstances = cachedCRInstances

	return resources
}

func convertWorkloads(resources *checks.DiscoveredResources, env *provider.TestEnvironment) {
	resources.Pods = make([]corev1.Pod, len(env.Pods))
	for i, pod := range env.Pods {
		if pod.Pod != nil {
			resources.Pods[i] = *pod.Pod
		}
	}

	// Populate Multus network data from provider pods
	resources.PodMultusNetworks = buildPodMultusNetworks(env.Pods)

	resources.Deployments = make([]appsv1.Deployment, len(env.Deployments))
	for i, dep := range env.Deployments {
		if dep.Deployment != nil {
			resources.Deployments[i] = *dep.Deployment
		}
	}

	resources.StatefulSets = make([]appsv1.StatefulSet, len(env.StatefulSets))
	for i, sts := range env.StatefulSets {
		if sts.StatefulSet != nil {
			resources.StatefulSets[i] = *sts.StatefulSet
		}
	}

	resources.Services = derefSlice(env.Services)
	resources.ServiceAccounts = derefSlice(env.ServiceAccounts)
	resources.CRDs = derefSlice(env.Crds)

	resources.CSVs = derefSlice(env.AllCsvs)

	convertScalingConfig(resources, env)
}

func convertScalingConfig(resources *checks.DiscoveredResources, env *provider.TestEnvironment) {
	// Managed workload names
	for _, m := range env.Config.ManagedDeployments {
		resources.ManagedDeployments = append(resources.ManagedDeployments, m.Name)
	}
	for _, m := range env.Config.ManagedStatefulsets {
		resources.ManagedStatefulSets = append(resources.ManagedStatefulSets, m.Name)
	}

	// Skip lists
	for _, s := range env.Config.SkipScalingTestDeployments {
		resources.SkipScalingDeployments = append(resources.SkipScalingDeployments, checks.SkipScalingEntry{
			Name:      s.Name,
			Namespace: s.Namespace,
		})
	}
	for _, s := range env.Config.SkipScalingTestStatefulSets {
		resources.SkipScalingStatefulSets = append(resources.SkipScalingStatefulSets, checks.SkipScalingEntry{
			Name:      s.Name,
			Namespace: s.Namespace,
		})
	}

	// CRD filters
	for _, f := range env.Config.CrdFilters {
		resources.CRDFilters = append(resources.CRDFilters, checks.CRDFilter{
			NameSuffix: f.NameSuffix,
			Scalable:   f.Scalable,
		})
	}

	// HPAs
	for _, hpa := range env.HorizontalScaler {
		if hpa == nil {
			continue
		}
		resources.HPAs = append(resources.HPAs, checks.HPAInfo{
			Name:       hpa.Name,
			Namespace:  hpa.Namespace,
			TargetKind: hpa.Spec.ScaleTargetRef.Kind,
			TargetName: hpa.Spec.ScaleTargetRef.Name,
		})
	}
}

func convertRBACAndPolicies(resources *checks.DiscoveredResources, env *provider.TestEnvironment) {
	resources.Roles = env.Roles
	resources.RoleBindings = env.RoleBindings
	resources.ClusterRoleBindings = env.ClusterRoleBindings
	resources.NetworkPolicies = env.NetworkPolicies
	resources.ResourceQuotas = env.ResourceQuotas
	resources.PodDisruptionBudgets = env.PodDisruptionBudgets
	resources.StorageClasses = env.StorageClassList
}

func convertClusterResources(resources *checks.DiscoveredResources, env *provider.TestEnvironment) {
	resources.Nodes = convertNodes(env.Nodes)
	resources.PersistentVolumes = env.PersistentVolumes
	resources.PersistentVolumeClaims = env.PersistentVolumeClaims
	resources.ClusterOperators = env.ClusterOperators

	// OLM resources
	resources.CatalogSources = derefSlice(env.AllCatalogSources)
	resources.PackageManifests = derefSlice(env.AllPackageManifests)
	resources.Subscriptions = env.AllSubscriptions

	// Networking
	resources.NetworkAttachmentDefinitions = env.NetworkAttachmentDefinitions
	resources.SriovNetworks = env.AllSriovNetworks
	resources.SriovNetworkNodePolicies = env.AllSriovNetworkNodePolicies

	// Cluster metadata
	resources.K8sVersion = env.K8sVersion
	resources.OpenshiftVersion = env.OpenshiftVersion
	resources.OCPStatus = env.OCPStatus
}

func convertCertificationResources(resources *checks.DiscoveredResources, env *provider.TestEnvironment) {
	resources.HelmChartReleases = convertHelmReleases(env)
	clients := clientsholder.GetClientsHolder()
	resources.K8sClientset = clients.K8sClient
	resources.ScaleClient = clients.ScalingClient

	// Cache the cert validator -- creating it involves HTTP pings and DB loading
	cachedCertValidatorOnce.Do(func() {
		cachedCertValidator = NewCertValidator(env.GetOfflineDBPath())
	})
	resources.CertValidator = cachedCertValidator
}

// derefSlice converts a slice of pointers to a slice of values, skipping nil entries.
func derefSlice[T any](ptrs []*T) []T {
	result := make([]T, 0, len(ptrs))
	for _, p := range ptrs {
		if p != nil {
			result = append(result, *p)
		}
	}
	return result
}

func convertNodes(nodeMap map[string]provider.Node) []corev1.Node {
	result := make([]corev1.Node, 0, len(nodeMap))
	for _, node := range nodeMap {
		if node.Data != nil {
			result = append(result, *node.Data)
		}
	}
	return result
}

// buildCRInstances lists CR instances for each CRD using the dynamic client.
func buildCRInstances(resources *checks.DiscoveredResources) map[string]map[string][]string {
	if len(resources.CRDs) == 0 {
		return nil
	}
	clients := clientsholder.GetClientsHolder()
	if clients.DynamicClient == nil {
		return nil
	}

	result := make(map[string]map[string][]string)
	for i := range resources.CRDs {
		crd := &resources.CRDs[i]
		if len(crd.Spec.Versions) == 0 {
			continue
		}
		version := crd.Spec.Versions[0].Name
		for j := range crd.Spec.Versions {
			if crd.Spec.Versions[j].Served {
				version = crd.Spec.Versions[j].Name
				break
			}
		}

		gvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  version,
			Resource: crd.Spec.Names.Plural,
		}

		list, err := clients.DynamicClient.Resource(gvr).Namespace("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Debug("Failed to list CRs for %s: %v", crd.Name, err)
			continue
		}

		nsCRs := make(map[string][]string)
		for k := range list.Items {
			cr := &list.Items[k]
			nsCRs[cr.GetNamespace()] = append(nsCRs[cr.GetNamespace()], cr.GetName())
		}
		if len(nsCRs) > 0 {
			result[crd.Name] = nsCRs
		}
	}
	return result
}

func convertHelmReleases(env *provider.TestEnvironment) []checks.HelmChartRelease {
	result := make([]checks.HelmChartRelease, 0, len(env.HelmChartReleases))
	for _, rel := range env.HelmChartReleases {
		if rel == nil || rel.Chart == nil || rel.Chart.Metadata == nil {
			continue
		}
		result = append(result, checks.HelmChartRelease{
			Name:      rel.Name,
			Namespace: rel.Namespace,
			Version:   rel.Chart.Metadata.Version,
		})
	}
	return result
}

// buildPodMultusNetworks extracts Multus (secondary) network interface data from
// the provider Pod objects, which have already parsed the
// k8s.v1.cni.cncf.io/network-status annotation.
func buildPodMultusNetworks(pods []*provider.Pod) map[string][]checks.MultusNetwork {
	result := make(map[string][]checks.MultusNetwork)
	for _, pod := range pods {
		if pod.Pod == nil || len(pod.MultusNetworkInterfaces) == 0 {
			continue
		}
		podKey := pod.Namespace + "/" + pod.Name
		var networks []checks.MultusNetwork
		for netName, iface := range pod.MultusNetworkInterfaces {
			networks = append(networks, checks.MultusNetwork{
				Name:          netName,
				InterfaceName: iface.Interface,
				IPs:           iface.IPs,
			})
		}
		if len(networks) > 0 {
			result[podKey] = networks
		}
	}
	return result
}
