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
	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmpackagev1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/checks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// ConvertToDiscoveredResources converts provider.TestEnvironment to checks.DiscoveredResources.
func ConvertToDiscoveredResources(env *provider.TestEnvironment) *checks.DiscoveredResources {
	resources := &checks.DiscoveredResources{
		Namespaces:    env.Namespaces,
		ProbePods:     env.ProbePods,
		ProbeExecutor: &ProbeExecutorAdapter{env: env},
	}

	convertWorkloads(resources, env)
	convertRBACAndPolicies(resources, env)
	convertClusterResources(resources, env)
	convertCertificationResources(resources, env)

	return resources
}

func convertWorkloads(resources *checks.DiscoveredResources, env *provider.TestEnvironment) {
	resources.Pods = make([]corev1.Pod, len(env.Pods))
	for i, pod := range env.Pods {
		if pod.Pod != nil {
			resources.Pods[i] = *pod.Pod
		}
	}

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

	resources.Services = convertServicePointers(env.Services)
	resources.ServiceAccounts = convertServiceAccountPointers(env.ServiceAccounts)
	resources.CRDs = convertCRDPointers(env.Crds)

	resources.CSVs = make([]olmv1Alpha.ClusterServiceVersion, 0, len(env.AllCsvs))
	for _, csv := range env.AllCsvs {
		if csv != nil {
			resources.CSVs = append(resources.CSVs, *csv)
		}
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
	resources.ClusterOperators = env.ClusterOperators

	// OLM resources
	resources.CatalogSources = convertCatalogSourcePointers(env.AllCatalogSources)
	resources.PackageManifests = convertPackageManifestPointers(env.AllPackageManifests)
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
	resources.K8sClientset = clientsholder.GetClientsHolder().K8sClient
	resources.CertValidator = NewCertValidator(env.GetOfflineDBPath())
}

// Helper functions to convert pointer slices to value slices

func convertServicePointers(services []*corev1.Service) []corev1.Service {
	result := make([]corev1.Service, 0, len(services))
	for _, svc := range services {
		if svc != nil {
			result = append(result, *svc)
		}
	}
	return result
}

func convertServiceAccountPointers(accounts []*corev1.ServiceAccount) []corev1.ServiceAccount {
	result := make([]corev1.ServiceAccount, 0, len(accounts))
	for _, sa := range accounts {
		if sa != nil {
			result = append(result, *sa)
		}
	}
	return result
}

func convertCRDPointers(crds []*apiextv1.CustomResourceDefinition) []apiextv1.CustomResourceDefinition {
	result := make([]apiextv1.CustomResourceDefinition, 0, len(crds))
	for _, crd := range crds {
		if crd != nil {
			result = append(result, *crd)
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

func convertCatalogSourcePointers(sources []*olmv1Alpha.CatalogSource) []olmv1Alpha.CatalogSource {
	result := make([]olmv1Alpha.CatalogSource, 0, len(sources))
	for _, cs := range sources {
		if cs != nil {
			result = append(result, *cs)
		}
	}
	return result
}

func convertPackageManifestPointers(manifests []*olmpackagev1.PackageManifest) []olmpackagev1.PackageManifest {
	result := make([]olmpackagev1.PackageManifest, 0, len(manifests))
	for _, pm := range manifests {
		if pm != nil {
			result = append(result, *pm)
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
