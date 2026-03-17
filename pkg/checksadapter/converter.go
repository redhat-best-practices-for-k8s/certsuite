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

	// Convert pods (unwrap provider.Pod -> corev1.Pod)
	resources.Pods = make([]corev1.Pod, len(env.Pods))
	for i, pod := range env.Pods {
		if pod.Pod != nil {
			resources.Pods[i] = *pod.Pod
		}
	}

	// Convert deployments
	resources.Deployments = make([]appsv1.Deployment, len(env.Deployments))
	for i, dep := range env.Deployments {
		if dep.Deployment != nil {
			resources.Deployments[i] = *dep.Deployment
		}
	}

	// Convert statefulsets
	resources.StatefulSets = make([]appsv1.StatefulSet, len(env.StatefulSets))
	for i, sts := range env.StatefulSets {
		if sts.StatefulSet != nil {
			resources.StatefulSets[i] = *sts.StatefulSet
		}
	}

	// Convert services
	resources.Services = convertServicePointers(env.Services)

	// Convert service accounts
	resources.ServiceAccounts = convertServiceAccountPointers(env.ServiceAccounts)

	// Direct assignments (already correct type)
	resources.Roles = env.Roles
	resources.RoleBindings = env.RoleBindings
	resources.ClusterRoleBindings = env.ClusterRoleBindings
	resources.NetworkPolicies = env.NetworkPolicies
	resources.ResourceQuotas = env.ResourceQuotas
	resources.PodDisruptionBudgets = env.PodDisruptionBudgets
	resources.StorageClasses = env.StorageClassList

	// Convert CRDs
	resources.CRDs = convertCRDPointers(env.Crds)

	// Convert CSVs (from pointers to values)
	resources.CSVs = make([]olmv1Alpha.ClusterServiceVersion, 0, len(env.AllCsvs))
	for _, csv := range env.AllCsvs {
		if csv != nil {
			resources.CSVs = append(resources.CSVs, *csv)
		}
	}


	// Convert nodes
	resources.Nodes = convertNodes(env.Nodes)

	// Convert persistent volumes
	resources.PersistentVolumes = env.PersistentVolumes

	return resources
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
