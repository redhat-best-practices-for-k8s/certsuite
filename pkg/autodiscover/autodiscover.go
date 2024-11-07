// Copyright (C) 2020-2024 Red Hat, Inc.
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
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	nadClient "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	sriovNetworkOp "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	configv1 "github.com/openshift/api/config/v1"
	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmPkgv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/compatibility"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/podhelper"
	"helm.sh/helm/v3/pkg/release"
	appsv1 "k8s.io/api/apps/v1"
	scalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// NonOpenshiftClusterVersion is a fake version number for non openshift clusters (kind/minikube)
	NonOpenshiftClusterVersion = "0.0.0"
	tnfCsvTargetLabelName      = "operator"
	tnfCsvTargetLabelValue     = ""
	tnfLabelPrefix             = "redhat-best-practices-for-k8s.com"
	labelTemplate              = "%s/%s"
)

type DiscoveredTestData struct {
	Env                          configuration.TestParameters
	Pods                         []corev1.Pod
	AllPods                      []corev1.Pod
	ProbePods                    []corev1.Pod
	CSVToPodListMap              map[types.NamespacedName][]*corev1.Pod
	OperandPods                  []*corev1.Pod
	ResourceQuotaItems           []corev1.ResourceQuota
	PodDisruptionBudgets         []policyv1.PodDisruptionBudget
	NetworkPolicies              []networkingv1.NetworkPolicy
	Crds                         []*apiextv1.CustomResourceDefinition
	Namespaces                   []string
	AllNamespaces                []string
	AbnormalEvents               []corev1.Event
	Csvs                         []*olmv1Alpha.ClusterServiceVersion
	AllCrds                      []*apiextv1.CustomResourceDefinition
	AllCsvs                      []*olmv1Alpha.ClusterServiceVersion
	AllInstallPlans              []*olmv1Alpha.InstallPlan
	AllCatalogSources            []*olmv1Alpha.CatalogSource
	AllPackageManifests          []*olmPkgv1.PackageManifest
	SriovNetworks                []sriovNetworkOp.SriovNetwork
	SriovNetworkNodePolicies     []sriovNetworkOp.SriovNetworkNodePolicy
	AllSriovNetworks             []sriovNetworkOp.SriovNetwork
	AllSriovNetworkNodePolicies  []sriovNetworkOp.SriovNetworkNodePolicy
	NetworkAttachmentDefinitions []nadClient.NetworkAttachmentDefinition
	Deployments                  []appsv1.Deployment
	StatefulSet                  []appsv1.StatefulSet
	PersistentVolumes            []corev1.PersistentVolume
	PersistentVolumeClaims       []corev1.PersistentVolumeClaim
	ClusterRoleBindings          []rbacv1.ClusterRoleBinding
	RoleBindings                 []rbacv1.RoleBinding // Contains all rolebindings from all namespaces
	Roles                        []rbacv1.Role        // Contains all roles from all namespaces
	Services                     []*corev1.Service
	ServiceAccounts              []*corev1.ServiceAccount
	AllServiceAccounts           []*corev1.ServiceAccount
	Hpas                         []*scalingv1.HorizontalPodAutoscaler
	Subscriptions                []olmv1Alpha.Subscription
	AllSubscriptions             []olmv1Alpha.Subscription
	HelmChartReleases            map[string][]*release.Release
	K8sVersion                   string
	OpenshiftVersion             string
	OCPStatus                    string
	Nodes                        *corev1.NodeList
	IstioServiceMeshFound        bool
	ValidProtocolNames           []string
	StorageClasses               []storagev1.StorageClass
	ServicesIgnoreList           []string
	ScaleCrUnderTest             []ScaleObject
	ExecutedBy                   string
	PartnerName                  string
	CollectorAppPassword         string
	CollectorAppEndpoint         string
}

type labelObject struct {
	LabelKey   string
	LabelValue string
}

var data = DiscoveredTestData{}

const labelRegex = `(\S*)\s*:\s*(\S*)`
const labelRegexMatches = 3

func createLabels(labelStrings []string) (labelObjects []labelObject) {
	for _, label := range labelStrings {
		r := regexp.MustCompile(labelRegex)

		values := r.FindStringSubmatch(label)
		if len(values) != labelRegexMatches {
			log.Error("Failed to parse label %q. It will not be used!, ", label)
			continue
		}
		var aLabel labelObject
		aLabel.LabelKey = values[1]
		aLabel.LabelValue = values[2]
		labelObjects = append(labelObjects, aLabel)
	}
	return labelObjects
}

// DoAutoDiscover finds objects under test
//
//nolint:funlen,gocyclo
func DoAutoDiscover(config *configuration.TestConfiguration) DiscoveredTestData {
	oc := clientsholder.GetClientsHolder()

	var err error
	data.StorageClasses, err = getAllStorageClasses(oc.K8sClient.StorageV1())
	if err != nil {
		log.Fatal("Failed to retrieve storageClasses - err: %v", err)
	}

	podsUnderTestLabelsObjects := createLabels(config.PodsUnderTestLabels)
	operatorsUnderTestLabelsObjects := createLabels(config.OperatorsUnderTestLabels)

	log.Debug("Pods under test labels: %+v", podsUnderTestLabelsObjects)
	log.Debug("Operators under test labels: %+v", operatorsUnderTestLabelsObjects)

	data.AllNamespaces, err = getAllNamespaces(oc.K8sClient.CoreV1())
	if err != nil {
		log.Fatal("Cannot get namespaces, err: %v", err)
	}
	data.AllSubscriptions = findSubscriptions(oc.OlmClient, []string{""})
	data.AllCsvs, err = getAllOperators(oc.OlmClient)
	if err != nil {
		log.Error("Cannot get operators, err: %v", err)
	}
	data.AllInstallPlans = getAllInstallPlans(oc.OlmClient)
	data.AllCatalogSources = getAllCatalogSources(oc.OlmClient)
	data.AllPackageManifests = getAllPackageManifests(oc.OlmPkgClient)
	data.Namespaces = namespacesListToStringList(config.TargetNameSpaces)
	data.Pods, data.AllPods = findPodsByLabels(oc.K8sClient.CoreV1(), podsUnderTestLabelsObjects, data.Namespaces)
	data.AbnormalEvents = findAbnormalEvents(oc.K8sClient.CoreV1(), data.Namespaces)
	probeLabels := []labelObject{{LabelKey: probeHelperPodsLabelName, LabelValue: probeHelperPodsLabelValue}}
	probeNS := []string{config.ProbeDaemonSetNamespace}
	data.ProbePods, _ = findPodsByLabels(oc.K8sClient.CoreV1(), probeLabels, probeNS)
	data.ResourceQuotaItems, err = getResourceQuotas(oc.K8sClient.CoreV1())
	if err != nil {
		log.Fatal("Cannot get resource quotas, err: %v", err)
	}
	data.PodDisruptionBudgets, err = getPodDisruptionBudgets(oc.K8sClient.PolicyV1(), data.Namespaces)
	if err != nil {
		log.Fatal("Cannot get pod disruption budgets, err: %v", err)
	}
	data.NetworkPolicies, err = getNetworkPolicies(oc.K8sNetworkingClient)
	if err != nil {
		log.Fatal("Cannot get network policies, err: %v", err)
	}

	// Get cluster crds
	data.AllCrds, err = getClusterCrdNames()
	if err != nil {
		log.Fatal("Cannot get cluster CRD names, err: %v", err)
	}
	data.Crds = FindTestCrdNames(data.AllCrds, config.CrdFilters)

	data.ScaleCrUnderTest = GetScaleCrUnderTest(data.Namespaces, data.Crds)
	data.Csvs = findOperatorsByLabels(oc.OlmClient, operatorsUnderTestLabelsObjects, config.TargetNameSpaces)
	data.Subscriptions = findSubscriptions(oc.OlmClient, data.Namespaces)
	data.HelmChartReleases = getHelmList(oc.RestConfig, data.Namespaces)

	// Get all operator pods
	data.CSVToPodListMap, err = getOperatorCsvPods(data.Csvs)
	if err != nil {
		log.Fatal("Failed to get the operator pods, err: %v", err)
	}

	// Best effort mode autodiscovery for operand (running-only) pods.
	pods, _ := findPodsByLabels(oc.K8sClient.CoreV1(), nil, data.Namespaces)
	if err != nil {
		log.Fatal("Failed to get running pods, err: %v", err)
	}

	data.OperandPods, err = getOperandPodsFromTestCsvs(data.Csvs, pods)
	if err != nil {
		log.Fatal("Failed to get operand pods, err: %v", err)
	}

	openshiftVersion, err := getOpenshiftVersion(oc.OcpClient)
	if err != nil {
		log.Fatal("Failed to get the OpenShift version, err: %v", err)
	}

	data.OpenshiftVersion = openshiftVersion
	k8sVersion, err := oc.K8sClient.Discovery().ServerVersion()
	if err != nil {
		log.Fatal("Cannot get the K8s version, err: %v", err)
	}
	data.ValidProtocolNames = config.ValidProtocolNames
	data.ServicesIgnoreList = config.ServicesIgnoreList

	// Find the status of the OCP version (pre-ga, end-of-life, maintenance, or generally available)
	data.OCPStatus = compatibility.DetermineOCPStatus(openshiftVersion, time.Now())

	data.K8sVersion = k8sVersion.GitVersion
	data.Deployments = findDeploymentsByLabels(oc.K8sClient.AppsV1(), podsUnderTestLabelsObjects, data.Namespaces)
	data.StatefulSet = findStatefulSetsByLabels(oc.K8sClient.AppsV1(), podsUnderTestLabelsObjects, data.Namespaces)

	// Check if the Istio Service Mesh is present
	data.IstioServiceMeshFound = isIstioServiceMeshInstalled(oc.K8sClient.AppsV1(), data.AllNamespaces)

	// Find ClusterRoleBindings
	clusterRoleBindings, err := getClusterRoleBindings(oc.K8sClient.RbacV1())
	if err != nil {
		log.Fatal("Cannot get cluster role bindings, err: %v", err)
	}
	data.ClusterRoleBindings = clusterRoleBindings
	// Find RoleBindings
	roleBindings, err := getRoleBindings(oc.K8sClient.RbacV1())
	if err != nil {
		log.Fatal("Cannot get role bindings, error: %v", err)
	}
	data.RoleBindings = roleBindings
	// find roles
	roles, err := getRoles(oc.K8sClient.RbacV1())
	if err != nil {
		log.Fatal("Cannot get roles, err: %v", err)
	}
	data.Roles = roles
	data.Hpas = findHpaControllers(oc.K8sClient, data.Namespaces)
	data.Nodes, err = oc.K8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal("Cannot get list of nodes, err: %v", err)
	}
	data.PersistentVolumes, err = getPersistentVolumes(oc.K8sClient.CoreV1())
	if err != nil {
		log.Fatal("Cannot get list of persistent volumes, error: %v", err)
	}
	data.PersistentVolumeClaims, err = getPersistentVolumeClaims(oc.K8sClient.CoreV1())
	if err != nil {
		log.Fatal("Cannot get list of persistent volume claims, err: %v", err)
	}
	data.Services, err = getServices(oc.K8sClient.CoreV1(), data.Namespaces, data.ServicesIgnoreList)
	if err != nil {
		log.Fatal("Cannot get list of services, err: %v", err)
	}
	data.ServiceAccounts, err = getServiceAccounts(oc.K8sClient.CoreV1(), data.Namespaces)
	if err != nil {
		log.Fatal("Cannot get list of service accounts under test, err: %v", err)
	}
	data.AllServiceAccounts, err = getServiceAccounts(oc.K8sClient.CoreV1(), []string{metav1.NamespaceAll})
	if err != nil {
		log.Fatal("Cannot get list of all service accounts, err: %v", err)
	}

	data.SriovNetworks, err = getSriovNetworks(oc, data.Namespaces)
	if err != nil {
		log.Fatal("Cannot get list of sriov networks, err: %v", err)
	}

	data.SriovNetworkNodePolicies, err = getSriovNetworkNodePolicies(oc, data.Namespaces)
	if err != nil {
		log.Fatal("Cannot get list of sriov network node policies, err: %v", err)
	}

	data.AllSriovNetworks, err = getSriovNetworks(oc, data.AllNamespaces)
	if err != nil {
		log.Fatal("Cannot get list of sriov networks, err: %v", err)
	}

	data.AllSriovNetworkNodePolicies, err = getSriovNetworkNodePolicies(oc, data.AllNamespaces)
	if err != nil {
		log.Fatal("Cannot get list of sriov network node policies, err: %v", err)
	}

	data.NetworkAttachmentDefinitions, err = getNetworkAttachmentDefinitions(oc, data.Namespaces)
	if err != nil {
		log.Fatal("Cannot get list of network attachment definitions, err: %v", err)
	}

	data.ExecutedBy = config.ExecutedBy
	data.PartnerName = config.PartnerName
	data.CollectorAppPassword = config.CollectorAppPassword
	data.CollectorAppEndpoint = config.CollectorAppEndpoint

	return data
}

func getNetworkAttachmentDefinitions(client *clientsholder.ClientsHolder, namespaces []string) ([]nadClient.NetworkAttachmentDefinition, error) {
	var nadList []nadClient.NetworkAttachmentDefinition

	// Check if the network-attachment-definitions resource exists
	if !DoesAPIResourceExist(client, "network-attachment-definitions") {
		return nadList, nil
	}

	for _, ns := range namespaces {
		nad, err := client.CNCFNetworkingClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		// Append the list of networkAttachmentDefinitions to the nadList slice
		nadList = append(nadList, nad.Items...)
	}

	return nadList, nil
}

func getSriovNetworks(client *clientsholder.ClientsHolder, namespaces []string) (sriovNetworks []sriovNetworkOp.SriovNetwork, err error) {
	var sriovNetworkList []sriovNetworkOp.SriovNetwork

	// Check if the sriovnetworks resource exists
	if !DoesAPIResourceExist(client, "sriovnetworks") {
		return sriovNetworkList, nil
	}

	for _, ns := range namespaces {
		snl, err := client.SriovNetworkingClient.SriovNetworks(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		// Append the list of sriovNetworks to the sriovNetworks slice
		sriovNetworkList = append(sriovNetworkList, snl.Items...)
	}
	return sriovNetworkList, nil
}

func getSriovNetworkNodePolicies(client *clientsholder.ClientsHolder, namespaces []string) (sriovNetworkNodePolicies []sriovNetworkOp.SriovNetworkNodePolicy, err error) {
	var sriovNetworkNodePolicyList []sriovNetworkOp.SriovNetworkNodePolicy

	// Check if the sriovnetworknodepolicies resource exists
	if !DoesAPIResourceExist(client, "sriovnetworknodepolicies") {
		return sriovNetworkNodePolicyList, nil
	}

	for _, ns := range namespaces {
		snnp, err := client.SriovNetworkingClient.SriovNetworkNodePolicies(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		// Append the list of sriovNetworkNodePolicies to the sriovNetworkNodePolicies slice
		sriovNetworkNodePolicyList = append(sriovNetworkNodePolicyList, snnp.Items...)
	}
	return sriovNetworkNodePolicyList, nil
}

func namespacesListToStringList(namespaceList []configuration.Namespace) (stringList []string) {
	for _, ns := range namespaceList {
		stringList = append(stringList, ns.Name)
	}
	return stringList
}

func getOpenshiftVersion(oClient clientconfigv1.ConfigV1Interface) (ver string, err error) {
	var clusterOperator *configv1.ClusterOperator
	clusterOperator, err = oClient.ClusterOperators().Get(context.TODO(), "openshift-apiserver", metav1.GetOptions{})
	if err != nil {
		switch {
		case kerrors.IsNotFound(err):
			log.Warn("Unable to get ClusterOperator CR from openshift-apiserver. Running in a non-OCP cluster.")
			return NonOpenshiftClusterVersion, nil
		default:
			return "", err
		}
	}

	for _, ver := range clusterOperator.Status.Versions {
		if ver.Name == tnfCsvTargetLabelName {
			// openshift-apiserver does not report version,
			// clusteroperator/openshift-apiserver does, and only version number
			log.Info("OpenShift Version found: %v", ver.Version)
			return ver.Version, nil
		}
	}

	return "", errors.New("could not get openshift version from clusterOperator")
}

// Get a map of csvs with its managed operator/controller pods from its installation namespace.
func getOperatorCsvPods(csvList []*olmv1Alpha.ClusterServiceVersion) (map[types.NamespacedName][]*corev1.Pod, error) {
	const nsAnnotation = "olm.operatorNamespace"

	client := clientsholder.GetClientsHolder()
	csvToPodsMapping := make(map[types.NamespacedName][]*corev1.Pod)

	// The operator's pod (controller) should run in the subscription/operatorgroup ns.
	for _, csv := range csvList {
		ns, found := csv.Annotations[nsAnnotation]
		if !found {
			return nil, fmt.Errorf("failed to get ns annotation %q from csv %v/%v", nsAnnotation, csv.Namespace, csv.Name)
		}

		pods, err := getPodsOwnedByCsv(csv.Name, strings.TrimSpace(ns), client)
		if err != nil {
			return nil, fmt.Errorf("failed to get pods from ns %v: %v", ns, err)
		}

		csvToPodsMapping[types.NamespacedName{Name: csv.Name, Namespace: csv.Namespace}] = pods
	}
	return csvToPodsMapping, nil
}

// This function gets the operator/controller pods of the specified csv name in from the installation namespace.
func getPodsOwnedByCsv(csvName, operatorNamespace string, client *clientsholder.ClientsHolder) (managedPods []*corev1.Pod, err error) {
	// Get all pods from the target namespace
	podsList, err := client.K8sClient.CoreV1().Pods(operatorNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for index := range podsList.Items {
		// Get the top owners of the pod
		pod := podsList.Items[index]
		topOwners, err := podhelper.GetPodTopOwner(pod.Namespace, pod.OwnerReferences)
		if err != nil {
			return nil, fmt.Errorf("could not get top owners of Pod %s (in namespace %s), err=%v", pod.Name, pod.Namespace, err)
		}

		// check if owner matches with the csv
		for _, owner := range topOwners {
			// The owner must be in the targetNamespace
			if owner.Kind == olmv1Alpha.ClusterServiceVersionKind && owner.Namespace == operatorNamespace && owner.Name == csvName {
				managedPods = append(managedPods, &podsList.Items[index])
				break
			}
		}
	}
	return managedPods, nil
}

// getOperandPodsFromTestCsvs returns a subset of pods whose owner CRs are managed by any of the testCsvs.
func getOperandPodsFromTestCsvs(testCsvs []*olmv1Alpha.ClusterServiceVersion, pods []corev1.Pod) ([]*corev1.Pod, error) {
	// Helper var to store all the managed crds from the operators under test
	// They map key is "Kind.group/version" or "Kind.APIversion", which should be the same.
	//   e.g.: "Subscription.operators.coreos.com/v1alpha1"
	crds := map[string]*olmv1Alpha.ClusterServiceVersion{}

	// First, iterate on each testCsv to fill the helper crds map.
	for _, csv := range testCsvs {
		ownedCrds := csv.Spec.CustomResourceDefinitions.Owned
		if len(ownedCrds) == 0 {
			continue
		}

		for i := range ownedCrds {
			crd := &ownedCrds[i]

			_, group, found := strings.Cut(crd.Name, ".")
			if !found {
				return nil, fmt.Errorf("failed to parse resources and group from crd name %q", crd.Name)
			}

			log.Info("CSV %q owns crd %v", csv.Name, crd.Kind+"/"+group+"/"+crd.Version)

			crdPath := path.Join(crd.Kind, group, crd.Version)
			crds[crdPath] = csv
		}
	}

	// Now, iterate on every pod in the list to check whether they're owned by any of the CRs that
	// the csvs are managing.
	operandPods := []*corev1.Pod{}
	for i := range pods {
		pod := &pods[i]
		owners, err := podhelper.GetPodTopOwner(pod.Namespace, pod.OwnerReferences)
		if err != nil {
			return nil, fmt.Errorf("failed to get top owners of pod %v/%v: %v", pod.Namespace, pod.Name, err)
		}

		for _, owner := range owners {
			versionedCrdPath := path.Join(owner.Kind, owner.APIVersion)

			var csv *olmv1Alpha.ClusterServiceVersion
			if csv = crds[versionedCrdPath]; csv == nil {
				// The owner is not a CR or it's not a CR owned by any operator under test
				continue
			}

			log.Info("Pod %v/%v has owner CR %s of CRD %q (CSV %v)", pod.Namespace, pod.Name,
				owner.Name, versionedCrdPath, csv.Name)

			operandPods = append(operandPods, pod)
			break
		}
	}

	return operandPods, nil
}
