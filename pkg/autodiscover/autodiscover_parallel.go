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

package autodiscover

import (
	"context"
	"sync"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/compatibility"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DoAutoDiscoverParallel finds objects under test using parallel API calls
// This is an optimized version of DoAutoDiscover that runs independent API calls concurrently
//
//nolint:funlen,gocyclo
func DoAutoDiscoverParallel(config *configuration.TestConfiguration) DiscoveredTestData {
	start := time.Now()
	oc := clientsholder.GetClientsHolder()

	// Prepare labels early (no API calls)
	podsUnderTestLabelsObjects := CreateLabels(config.PodsUnderTestLabels)
	operatorsUnderTestLabelsObjects := CreateLabels(config.OperatorsUnderTestLabels)
	data.Namespaces = namespacesListToStringList(config.TargetNameSpaces)

	log.Debug("Pods under test labels: %+v", podsUnderTestLabelsObjects)
	log.Debug("Operators under test labels: %+v", operatorsUnderTestLabelsObjects)

	// ============================================================
	// PHASE 1: Independent cluster-wide queries (run in parallel)
	// ============================================================
	var wg1 sync.WaitGroup
	var mu sync.Mutex
	var fatalErrors []string

	// Helper to record fatal errors thread-safely
	recordFatal := func(msg string) {
		mu.Lock()
		fatalErrors = append(fatalErrors, msg)
		mu.Unlock()
	}

	log.Info("Starting Phase 1: Parallel cluster-wide discovery...")
	phase1Start := time.Now()

	// Storage Classes
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getAllStorageClasses(oc.K8sClient.StorageV1())
		if err != nil {
			recordFatal("Failed to retrieve storageClasses - err: " + err.Error())
			return
		}
		mu.Lock()
		data.StorageClasses = result
		mu.Unlock()
	}()

	// All Namespaces
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getAllNamespaces(oc.K8sClient.CoreV1())
		if err != nil {
			recordFatal("Cannot get namespaces, err: " + err.Error())
			return
		}
		mu.Lock()
		data.AllNamespaces = result
		mu.Unlock()
	}()

	// All Subscriptions
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result := findSubscriptions(oc.OlmClient.OperatorsV1alpha1(), []string{""})
		mu.Lock()
		data.AllSubscriptions = result
		mu.Unlock()
	}()

	// All CSVs (Operators)
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getAllOperators(oc.OlmClient.OperatorsV1alpha1())
		if err != nil {
			log.Error("Cannot get operators, err: %v", err)
		}
		mu.Lock()
		data.AllCsvs = result
		mu.Unlock()
	}()

	// All Install Plans
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result := getAllInstallPlans(oc.OlmClient.OperatorsV1alpha1())
		mu.Lock()
		data.AllInstallPlans = result
		mu.Unlock()
	}()

	// All Catalog Sources
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result := getAllCatalogSources(oc.OlmClient.OperatorsV1alpha1())
		mu.Lock()
		data.AllCatalogSources = result
		log.Info("Collected %d catalog sources during autodiscovery", len(data.AllCatalogSources))
		mu.Unlock()
	}()

	// All Package Manifests
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result := getAllPackageManifests(oc.OlmPkgClient.PackageManifests(""))
		mu.Lock()
		data.AllPackageManifests = result
		mu.Unlock()
	}()

	// All CRDs
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getClusterCrdNames()
		if err != nil {
			recordFatal("Cannot get cluster CRD names, err: " + err.Error())
			return
		}
		mu.Lock()
		data.AllCrds = result
		mu.Unlock()
	}()

	// Cluster Operators
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := findClusterOperators(oc.OcpClient.ClusterOperators())
		if err != nil {
			recordFatal("Failed to get cluster operators, err: " + err.Error())
			return
		}
		mu.Lock()
		data.ClusterOperators = result
		mu.Unlock()
	}()

	// Cluster Role Bindings
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getClusterRoleBindings(oc.K8sClient.RbacV1())
		if err != nil {
			recordFatal("Cannot get cluster role bindings, err: " + err.Error())
			return
		}
		mu.Lock()
		data.ClusterRoleBindings = result
		mu.Unlock()
	}()

	// Role Bindings (all namespaces)
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getRoleBindings(oc.K8sClient.RbacV1())
		if err != nil {
			recordFatal("Cannot get role bindings, error: " + err.Error())
			return
		}
		mu.Lock()
		data.RoleBindings = result
		mu.Unlock()
	}()

	// Roles (all namespaces)
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getRoles(oc.K8sClient.RbacV1())
		if err != nil {
			recordFatal("Cannot get roles, err: " + err.Error())
			return
		}
		mu.Lock()
		data.Roles = result
		mu.Unlock()
	}()

	// Nodes
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := oc.K8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			recordFatal("Cannot get list of nodes, err: " + err.Error())
			return
		}
		mu.Lock()
		data.Nodes = result
		mu.Unlock()
	}()

	// Persistent Volumes
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getPersistentVolumes(oc.K8sClient.CoreV1())
		if err != nil {
			recordFatal("Cannot get list of persistent volumes, error: " + err.Error())
			return
		}
		mu.Lock()
		data.PersistentVolumes = result
		mu.Unlock()
	}()

	// Persistent Volume Claims
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getPersistentVolumeClaims(oc.K8sClient.CoreV1())
		if err != nil {
			recordFatal("Cannot get list of persistent volume claims, err: " + err.Error())
			return
		}
		mu.Lock()
		data.PersistentVolumeClaims = result
		mu.Unlock()
	}()

	// OpenShift Version
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getOpenshiftVersion(oc.OcpClient)
		if err != nil {
			recordFatal("Failed to get the OpenShift version, err: " + err.Error())
			return
		}
		mu.Lock()
		data.OpenshiftVersion = result
		mu.Unlock()
	}()

	// K8s Version
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := oc.K8sClient.Discovery().ServerVersion()
		if err != nil {
			recordFatal("Cannot get the K8s version, err: " + err.Error())
			return
		}
		mu.Lock()
		data.K8sVersion = result.GitVersion
		mu.Unlock()
	}()

	// Resource Quotas
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getResourceQuotas(oc.K8sClient.CoreV1())
		if err != nil {
			recordFatal("Cannot get resource quotas, err: " + err.Error())
			return
		}
		mu.Lock()
		data.ResourceQuotaItems = result
		mu.Unlock()
	}()

	// Network Policies
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getNetworkPolicies(oc.K8sNetworkingClient)
		if err != nil {
			recordFatal("Cannot get network policies, err: " + err.Error())
			return
		}
		mu.Lock()
		data.NetworkPolicies = result
		mu.Unlock()
	}()

	// All Service Accounts
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		result, err := getServiceAccounts(oc.K8sClient.CoreV1(), []string{metav1.NamespaceAll})
		if err != nil {
			recordFatal("Cannot get list of all service accounts, err: " + err.Error())
			return
		}
		mu.Lock()
		data.AllServiceAccounts = result
		mu.Unlock()
	}()

	// Wait for Phase 1 to complete
	wg1.Wait()
	log.Info("Phase 1 completed in %v", time.Since(phase1Start))

	// Check for fatal errors
	if len(fatalErrors) > 0 {
		for _, e := range fatalErrors {
			log.Fatal("%s", e)
		}
	}

	// Set OCP status after version is available
	data.OCPStatus = compatibility.DetermineOCPStatus(data.OpenshiftVersion, time.Now())
	data.ValidProtocolNames = config.ValidProtocolNames
	data.ServicesIgnoreList = config.ServicesIgnoreList

	// Filter CRDs based on config
	data.Crds = FindTestCrdNames(data.AllCrds, config.CrdFilters)

	// ============================================================
	// PHASE 2: Namespace-scoped queries (run in parallel)
	// These depend on data.Namespaces or data.AllNamespaces
	// ============================================================
	var wg2 sync.WaitGroup
	log.Info("Starting Phase 2: Parallel namespace-scoped discovery...")
	phase2Start := time.Now()

	// Pods under test
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		pods, allPods := FindPodsByLabels(oc.K8sClient.CoreV1(), podsUnderTestLabelsObjects, data.Namespaces)
		mu.Lock()
		data.Pods = pods
		data.AllPods = allPods
		data.PodStates.BeforeExecution = CountPodsByStatus(data.AllPods)
		mu.Unlock()
	}()

	// Abnormal Events
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result := findAbnormalEvents(oc.K8sClient.CoreV1(), data.Namespaces)
		mu.Lock()
		data.AbnormalEvents = result
		mu.Unlock()
	}()

	// Probe Pods
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		probeLabels := []labelObject{{LabelKey: probeHelperPodsLabelName, LabelValue: probeHelperPodsLabelValue}}
		probeNS := []string{config.ProbeDaemonSetNamespace}
		result, _ := FindPodsByLabels(oc.K8sClient.CoreV1(), probeLabels, probeNS)
		mu.Lock()
		data.ProbePods = result
		mu.Unlock()
	}()

	// Pod Disruption Budgets
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result, err := getPodDisruptionBudgets(oc.K8sClient.PolicyV1(), data.Namespaces)
		if err != nil {
			recordFatal("Cannot get pod disruption budgets, err: " + err.Error())
			return
		}
		mu.Lock()
		data.PodDisruptionBudgets = result
		mu.Unlock()
	}()

	// Operators by labels
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result := findOperatorsByLabels(oc.OlmClient.OperatorsV1alpha1(), operatorsUnderTestLabelsObjects, config.TargetNameSpaces)
		mu.Lock()
		data.Csvs = result
		mu.Unlock()
	}()

	// Subscriptions in target namespaces
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result := findSubscriptions(oc.OlmClient.OperatorsV1alpha1(), data.Namespaces)
		mu.Lock()
		data.Subscriptions = result
		mu.Unlock()
	}()

	// Helm Chart Releases (this can be slow - consider parallelizing internally)
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result := getHelmList(oc.RestConfig, data.Namespaces)
		mu.Lock()
		data.HelmChartReleases = result
		mu.Unlock()
	}()

	// Deployments
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result := findDeploymentsByLabels(oc.K8sClient.AppsV1(), podsUnderTestLabelsObjects, data.Namespaces)
		mu.Lock()
		data.Deployments = result
		mu.Unlock()
	}()

	// StatefulSets
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result := findStatefulSetsByLabels(oc.K8sClient.AppsV1(), podsUnderTestLabelsObjects, data.Namespaces)
		mu.Lock()
		data.StatefulSet = result
		mu.Unlock()
	}()

	// Istio Service Mesh check
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result := isIstioServiceMeshInstalled(oc.K8sClient.AppsV1(), data.AllNamespaces)
		mu.Lock()
		data.IstioServiceMeshFound = result
		mu.Unlock()
	}()

	// HPAs
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result := findHpaControllers(oc.K8sClient, data.Namespaces)
		mu.Lock()
		data.Hpas = result
		mu.Unlock()
	}()

	// Services in target namespaces
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result, err := getServices(oc.K8sClient.CoreV1(), data.Namespaces, data.ServicesIgnoreList)
		if err != nil {
			recordFatal("Cannot get list of services, err: " + err.Error())
			return
		}
		mu.Lock()
		data.Services = result
		mu.Unlock()
	}()

	// All Services
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result, err := getServices(oc.K8sClient.CoreV1(), data.AllNamespaces, data.ServicesIgnoreList)
		if err != nil {
			recordFatal("Cannot get list of all services, err: " + err.Error())
			return
		}
		mu.Lock()
		data.AllServices = result
		mu.Unlock()
	}()

	// Service Accounts in target namespaces
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result, err := getServiceAccounts(oc.K8sClient.CoreV1(), data.Namespaces)
		if err != nil {
			recordFatal("Cannot get list of service accounts under test, err: " + err.Error())
			return
		}
		mu.Lock()
		data.ServiceAccounts = result
		mu.Unlock()
	}()

	// SRIOV Networks in target namespaces
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result, err := getSriovNetworks(oc, data.Namespaces)
		if err != nil {
			recordFatal("Cannot get list of sriov networks, err: " + err.Error())
			return
		}
		mu.Lock()
		data.SriovNetworks = result
		mu.Unlock()
	}()

	// SRIOV Network Node Policies in target namespaces
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result, err := getSriovNetworkNodePolicies(oc, data.Namespaces)
		if err != nil {
			recordFatal("Cannot get list of sriov network node policies, err: " + err.Error())
			return
		}
		mu.Lock()
		data.SriovNetworkNodePolicies = result
		mu.Unlock()
	}()

	// All SRIOV Networks
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result, err := getSriovNetworks(oc, data.AllNamespaces)
		if err != nil {
			recordFatal("Cannot get list of all sriov networks, err: " + err.Error())
			return
		}
		mu.Lock()
		data.AllSriovNetworks = result
		mu.Unlock()
	}()

	// All SRIOV Network Node Policies
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result, err := getSriovNetworkNodePolicies(oc, data.AllNamespaces)
		if err != nil {
			recordFatal("Cannot get list of all sriov network node policies, err: " + err.Error())
			return
		}
		mu.Lock()
		data.AllSriovNetworkNodePolicies = result
		mu.Unlock()
	}()

	// Network Attachment Definitions
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		result, err := getNetworkAttachmentDefinitions(oc, data.Namespaces)
		if err != nil {
			recordFatal("Cannot get list of network attachment definitions, err: " + err.Error())
			return
		}
		mu.Lock()
		data.NetworkAttachmentDefinitions = result
		mu.Unlock()
	}()

	// Wait for Phase 2 to complete
	wg2.Wait()
	log.Info("Phase 2 completed in %v", time.Since(phase2Start))

	// Check for fatal errors
	if len(fatalErrors) > 0 {
		for _, e := range fatalErrors {
			log.Fatal("%s", e)
		}
	}

	// ============================================================
	// PHASE 3: Queries that depend on CSVs or other Phase 2 results
	// ============================================================
	log.Info("Starting Phase 3: Dependent discovery...")
	phase3Start := time.Now()

	// Scale CR under test (depends on Namespaces and Crds)
	data.ScaleCrUnderTest = GetScaleCrUnderTest(data.Namespaces, data.Crds)

	// Get all operator pods (depends on data.Csvs)
	var err error
	data.CSVToPodListMap, err = getOperatorCsvPods(data.Csvs)
	if err != nil {
		log.Fatal("Failed to get the operator pods, err: %v", err)
	}

	// Best effort mode autodiscovery for operand (running-only) pods
	pods, _ := FindPodsByLabels(oc.K8sClient.CoreV1(), nil, data.Namespaces)
	data.OperandPods, err = getOperandPodsFromTestCsvs(data.Csvs, pods)
	if err != nil {
		log.Fatal("Failed to get operand pods, err: %v", err)
	}

	log.Info("Phase 3 completed in %v", time.Since(phase3Start))

	// Set remaining config values
	data.ExecutedBy = config.ExecutedBy
	data.PartnerName = config.PartnerName
	data.CollectorAppPassword = config.CollectorAppPassword
	data.CollectorAppEndpoint = config.CollectorAppEndpoint
	data.ConnectAPIKey = config.ConnectAPIConfig.APIKey
	data.ConnectAPIBaseURL = config.ConnectAPIConfig.BaseURL
	data.ConnectProjectID = config.ConnectAPIConfig.ProjectID
	data.ConnectAPIProxyURL = config.ConnectAPIConfig.ProxyURL
	data.ConnectAPIProxyPort = config.ConnectAPIConfig.ProxyPort

	log.Info("Total autodiscovery completed in %v", time.Since(start))

	return data
}
