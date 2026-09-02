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
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/arrayhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/checks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Cached expensive resources -- computed once, reused across all checks.
var (
	cachedResources         *checks.DiscoveredResources
	cachedResourcesOnce     sync.Once
	cachedCertValidator     checks.CertificationValidator
	cachedCertValidatorOnce sync.Once
	cachedCRInstances       map[string]map[string][]string
	cachedCRInstancesOnce   sync.Once
	cacheMutex              sync.RWMutex
)

// ResetCache clears the cached resources.
func ResetCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cachedResources = nil
	cachedResourcesOnce = sync.Once{}
	cachedCertValidator = nil
	cachedCertValidatorOnce = sync.Once{}
	cachedCRInstances = nil
	cachedCRInstancesOnce = sync.Once{}
}

// ConvertToDiscoveredResources converts provider.TestEnvironment to checks.DiscoveredResources.
// It caches the result until ResetCache is called.
func ConvertToDiscoveredResources(env *provider.TestEnvironment) *checks.DiscoveredResources {
	cacheMutex.RLock()
	if cachedResources != nil {
		defer cacheMutex.RUnlock()
		return cachedResources
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	cachedResourcesOnce.Do(func() {
		cachedResources = &checks.DiscoveredResources{
			Namespaces:    env.Namespaces,
			ProbePods:     env.ProbePods,
			ProbeExecutor: &ProbeExecutorAdapter{},
		}

		convertWorkloads(cachedResources, env)
		convertRBACAndPolicies(cachedResources, env)
		convertClusterResources(cachedResources, env)
		convertCertificationResources(cachedResources, env)

		// Use cached CR instances (expensive dynamic client calls)
		cachedCRInstancesOnce.Do(func() {
			cachedCRInstances = buildCRInstances(cachedResources)
		})
		cachedResources.CRInstances = cachedCRInstances
	})

	return cachedResources
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

	resources.Services = arrayhelper.DerefSlice(env.Services)
	resources.ServiceAccounts = arrayhelper.DerefSlice(env.ServiceAccounts)
	resources.CRDs = arrayhelper.DerefSlice(env.Crds)

	resources.CSVs = arrayhelper.DerefSlice(env.AllCsvs)

	convertScalingConfig(resources, env)
}

func convertScalingConfig(resources *checks.DiscoveredResources, env *provider.TestEnvironment) {
	resources.ManagedDeployments = make([]string, 0, len(env.Config.ManagedDeployments))
	for _, m := range env.Config.ManagedDeployments {
		resources.ManagedDeployments = append(resources.ManagedDeployments, m.Name)
	}
	resources.ManagedStatefulSets = make([]string, 0, len(env.Config.ManagedStatefulsets))
	for _, m := range env.Config.ManagedStatefulsets {
		resources.ManagedStatefulSets = append(resources.ManagedStatefulSets, m.Name)
	}

	resources.SkipScalingDeployments = make([]checks.SkipScalingEntry, 0, len(env.Config.SkipScalingTestDeployments))
	for _, s := range env.Config.SkipScalingTestDeployments {
		resources.SkipScalingDeployments = append(resources.SkipScalingDeployments, checks.SkipScalingEntry{
			Name:      s.Name,
			Namespace: s.Namespace,
		})
	}
	resources.SkipScalingStatefulSets = make([]checks.SkipScalingEntry, 0, len(env.Config.SkipScalingTestStatefulSets))
	for _, s := range env.Config.SkipScalingTestStatefulSets {
		resources.SkipScalingStatefulSets = append(resources.SkipScalingStatefulSets, checks.SkipScalingEntry{
			Name:      s.Name,
			Namespace: s.Namespace,
		})
	}

	resources.CRDFilters = make([]checks.CRDFilter, 0, len(env.Config.CrdFilters))
	for _, f := range env.Config.CrdFilters {
		resources.CRDFilters = append(resources.CRDFilters, checks.CRDFilter{
			NameSuffix: f.NameSuffix,
			Scalable:   f.Scalable,
		})
	}

	resources.HPAs = make([]checks.HPAInfo, 0, len(env.HorizontalScaler))
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
	resources.CatalogSources = arrayhelper.DerefSlice(env.AllCatalogSources)
	resources.PackageManifests = arrayhelper.DerefSlice(env.AllPackageManifests)
	resources.Subscriptions = env.AllSubscriptions

	// Networking
	resources.NetworkAttachmentDefinitions = env.NetworkAttachmentDefinitions
	resources.SriovNetworks = env.AllSriovNetworks
	resources.SriovNetworkNodePolicies = env.AllSriovNetworkNodePolicies

	// Cluster metadata
	resources.K8sVersion = env.K8sVersion
	resources.OpenshiftVersion = env.OpenshiftVersion
	resources.OCPStatus = env.OCPStatus

	resources.TLSSecurityProfile = env.TLSSecurityProfile
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

	type crResult struct {
		crdName string
		nsCRs   map[string][]string
	}
	resChan := make(chan crResult, len(resources.CRDs))
	var wg sync.WaitGroup

	for i := range resources.CRDs {
		wg.Add(1)
		go func(crd *apiextv1.CustomResourceDefinition) {
			defer wg.Done()
			if len(crd.Spec.Versions) == 0 {
				return
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
				return
			}

			nsCRs := make(map[string][]string)
			for k := range list.Items {
				cr := &list.Items[k]
				nsCRs[cr.GetNamespace()] = append(nsCRs[cr.GetNamespace()], cr.GetName())
			}
			if len(nsCRs) > 0 {
				resChan <- crResult{crdName: crd.Name, nsCRs: nsCRs}
			}
		}(&resources.CRDs[i])
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	result := make(map[string]map[string][]string)
	for res := range resChan {
		result[res.crdName] = res.nsCRs
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
// the provider Pod objects.
func buildPodMultusNetworks(pods []*provider.Pod) map[string][]checks.MultusNetwork {
	result := make(map[string][]checks.MultusNetwork)
	for _, pod := range pods {
		networks := pod.ToMultusNetwork()
		if len(networks) > 0 {
			podKey := pod.Namespace + "/" + pod.Name
			result[podKey] = networks
		}
	}
	return result
}

