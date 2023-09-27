// Copyright (C) 2020-2023 Red Hat, Inc.
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
	"regexp"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/compatibility"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
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
)

const (
	// NonOpenshiftClusterVersion is a fake version number for non openshift clusters (kind/minikube)
	NonOpenshiftClusterVersion = "0.0.0"
	tnfCsvTargetLabelName      = "operator"
	tnfCsvTargetLabelValue     = ""
	tnfLabelPrefix             = "test-network-function.com"
	labelTemplate              = "%s/%s"
)

type DiscoveredTestData struct {
	Env                    configuration.TestParameters
	Pods                   []corev1.Pod
	AllPods                []corev1.Pod
	DebugPods              []corev1.Pod
	ResourceQuotaItems     []corev1.ResourceQuota
	PodDisruptionBudgets   []policyv1.PodDisruptionBudget
	NetworkPolicies        []networkingv1.NetworkPolicy
	Crds                   []*apiextv1.CustomResourceDefinition
	Namespaces             []string
	AllNamespaces          []string
	AbnormalEvents         []corev1.Event
	Csvs                   []*olmv1Alpha.ClusterServiceVersion
	AllCsvs                []*olmv1Alpha.ClusterServiceVersion
	AllInstallPlans        []*olmv1Alpha.InstallPlan
	AllCatalogSources      []*olmv1Alpha.CatalogSource
	Deployments            []appsv1.Deployment
	StatefulSet            []appsv1.StatefulSet
	PersistentVolumes      []corev1.PersistentVolume
	PersistentVolumeClaims []corev1.PersistentVolumeClaim
	ClusterRoleBindings    []rbacv1.ClusterRoleBinding
	RoleBindings           []rbacv1.RoleBinding // Contains all rolebindings from all namespaces
	Roles                  []rbacv1.Role        // Contains all roles from all namespaces
	Services               []*corev1.Service
	Hpas                   []*scalingv1.HorizontalPodAutoscaler
	Subscriptions          []olmv1Alpha.Subscription
	AllSubscriptions       []olmv1Alpha.Subscription
	HelmChartReleases      map[string][]*release.Release
	K8sVersion             string
	OpenshiftVersion       string
	OCPStatus              string
	Nodes                  *corev1.NodeList
	IstioServiceMeshFound  bool
	ValidProtocolNames     []string
	StorageClasses         []storagev1.StorageClass
	ServicesIgnoreList     []string
	ScaleCrUnderTest       []ScaleObject
	CollectorAppEndPoint   string
	ExecutedBy             string
	PartnerName            string
	CollectorAppPassword   string
}

type labelObject struct {
	LabelKey   string
	LabelValue string
}

var data = DiscoveredTestData{}

func warnDeprecation(config *configuration.TestConfiguration) {
	if len(config.OperatorsUnderTestLabels) == 0 {
		logrus.Warnf("DEPRECATED: deprecated default operator label in use ( %s:%s ) is about to be obsolete. Please use the new \"operatorsUnderTestLabels\" field to specify operators labels instead.",
			deprecatedHardcodedOperatorLabelName, deprecatedHardcodedOperatorLabelValue)
	}
	if len(config.PodsUnderTestLabels) == 0 {
		logrus.Warn("No Pod under test labels configured. Tests on pods and containers will not run. Please use the \"podsUnderTestLabels\" field to specify labels for pods under test")
	}
}

const labelRegex = `(\S*)\s*:\s*(\S*)`
const labelRegexMatches = 3

func createLabels(labelStrings []string) (labelObjects []labelObject) {
	for _, label := range labelStrings {
		r := regexp.MustCompile(labelRegex)

		values := r.FindStringSubmatch(label)
		if len(values) != labelRegexMatches {
			logrus.Errorf("failed to parse label=%s, will not be used!, ", label)
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
//nolint:funlen
func DoAutoDiscover(config *configuration.TestConfiguration) DiscoveredTestData {
	oc := clientsholder.GetClientsHolder()

	var err error
	data.StorageClasses, err = getAllStorageClasses()
	if err != nil {
		logrus.Fatalf("Failed to retrieve storageClasses - err: %v", err)
	}

	podsUnderTestLabelsObjects := createLabels(config.PodsUnderTestLabels)
	operatorsUnderTestLabelsObjects := createLabels(config.OperatorsUnderTestLabels)

	// prints warning about deprecated labels
	warnDeprecation(config)
	// adds DEPRECATED hardcoded operator label
	operatorsUnderTestLabelsObjects = append(operatorsUnderTestLabelsObjects, labelObject{LabelKey: deprecatedHardcodedOperatorLabelName, LabelValue: deprecatedHardcodedOperatorLabelValue})

	logrus.Infof("parsed pods under test labels: %+v", podsUnderTestLabelsObjects)
	logrus.Infof("parsed operators under test labels: %+v", operatorsUnderTestLabelsObjects)

	data.AllNamespaces, _ = getAllNamespaces(oc.K8sClient.CoreV1())
	data.AllSubscriptions = findSubscriptions(oc.OlmClient, []string{""})
	data.AllCsvs = getAllOperators(oc.OlmClient)
	data.AllInstallPlans = getAllInstallPlans(oc.OlmClient)
	data.AllCatalogSources = getAllCatalogSources(oc.OlmClient)
	data.Namespaces = namespacesListToStringList(config.TargetNameSpaces)
	data.Pods, data.AllPods = findPodsByLabel(oc.K8sClient.CoreV1(), podsUnderTestLabelsObjects, data.Namespaces)
	data.AbnormalEvents = findAbnormalEvents(oc.K8sClient.CoreV1(), data.Namespaces)
	debugLabels := []labelObject{{LabelKey: debugHelperPodsLabelName, LabelValue: debugHelperPodsLabelValue}}
	debugNS := []string{config.DebugDaemonSetNamespace}
	data.DebugPods, _ = findPodsByLabel(oc.K8sClient.CoreV1(), debugLabels, debugNS)
	data.ResourceQuotaItems, err = getResourceQuotas(oc.K8sClient.CoreV1())
	if err != nil {
		logrus.Fatalf("Cannot get resource quotas, error: %v", err)
	}
	data.PodDisruptionBudgets, err = getPodDisruptionBudgets(oc.K8sClient.PolicyV1(), data.Namespaces)
	if err != nil {
		logrus.Fatalf("Cannot get pod disruption budgets, error: %v", err)
	}
	data.NetworkPolicies, err = getNetworkPolicies(oc.K8sNetworkingClient)
	if err != nil {
		logrus.Fatalln("Cannot get network policies")
	}
	data.Crds = FindTestCrdNames(config.CrdFilters)
	data.ScaleCrUnderTest = GetScaleCrUnderTest(data.Namespaces, data.Crds)
	data.Csvs = findOperatorsByLabel(oc.OlmClient, operatorsUnderTestLabelsObjects, config.TargetNameSpaces)
	data.Subscriptions = findSubscriptions(oc.OlmClient, data.Namespaces)
	data.HelmChartReleases = getHelmList(oc.RestConfig, data.Namespaces)

	openshiftVersion, err := getOpenshiftVersion(oc.OcpClient)
	if err != nil {
		logrus.Fatalf("Failed to get the OpenShift version: %v", err)
	}

	data.OpenshiftVersion = openshiftVersion
	k8sVersion, err := oc.K8sClient.Discovery().ServerVersion()
	if err != nil {
		logrus.Fatalf("Cannot get the K8s version, error: %v", err)
	}
	data.IstioServiceMeshFound = isIstioServiceMeshInstalled(data.AllNamespaces)
	data.ValidProtocolNames = config.ValidProtocolNames
	data.ServicesIgnoreList = config.ServicesIgnoreList

	// Find the status of the OCP version (pre-ga, end-of-life, maintenance, or generally available)
	data.OCPStatus = compatibility.DetermineOCPStatus(openshiftVersion, time.Now())

	data.K8sVersion = k8sVersion.GitVersion
	data.Deployments = findDeploymentByLabel(oc.K8sClient.AppsV1(), podsUnderTestLabelsObjects, data.Namespaces)
	data.StatefulSet = findStatefulSetByLabel(oc.K8sClient.AppsV1(), podsUnderTestLabelsObjects, data.Namespaces)
	// Find ClusterRoleBindings
	clusterRoleBindings, err := getClusterRoleBindings()
	if err != nil {
		logrus.Fatalf("Cannot get cluster role bindings, error: %v", err)
	}
	data.ClusterRoleBindings = clusterRoleBindings
	// Find RoleBindings
	roleBindings, err := getRoleBindings()
	if err != nil {
		logrus.Fatalf("Cannot get cluster role bindings, error: %v", err)
	}
	data.RoleBindings = roleBindings
	// find roles
	roles, err := getRoles()
	if err != nil {
		logrus.Fatalf("Cannot get roles, error: %v", err)
	}
	data.Roles = roles
	data.Hpas = findHpaControllers(oc.K8sClient, data.Namespaces)
	data.Nodes, err = oc.K8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Fatalf("Cannot get list of nodes, error: %v", err)
	}
	data.PersistentVolumes, err = getPersistentVolumes(oc.K8sClient.CoreV1())
	if err != nil {
		logrus.Fatalf("Cannot get list of persistent volumes, error: %v", err)
	}
	data.PersistentVolumeClaims, err = getPersistentVolumeClaims(oc.K8sClient.CoreV1())
	if err != nil {
		logrus.Fatalf("Cannot get list of persistent volume claims, error: %v", err)
	}
	data.Services, err = getServices(oc.K8sClient.CoreV1(), data.Namespaces, data.ServicesIgnoreList)
	if err != nil {
		logrus.Fatalf("Cannot get list of services, error: %v", err)
	}

	if config.CollectorAppEndPoint == "" {
		config.CollectorAppEndPoint = "http://localhost:8080"
	}
	data.CollectorAppEndPoint = config.CollectorAppEndPoint
	data.ExecutedBy = config.ExecutedBy
	data.PartnerName = config.PartnerName
	data.CollectorAppPassword = config.CollectorAppPassword
	return data
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
			logrus.Warnf("Unable to get ClusterOperator CR from openshift-apiserver. Running in a non-OCP cluster.")
			return NonOpenshiftClusterVersion, nil
		default:
			return "", err
		}
	}

	for _, ver := range clusterOperator.Status.Versions {
		if ver.Name == tnfCsvTargetLabelName {
			// openshift-apiserver does not report version,
			// clusteroperator/openshift-apiserver does, and only version number
			logrus.Infof("OpenShift Version found: %v", ver.Version)
			return ver.Version, nil
		}
	}

	return "", errors.New("could not get openshift version from clusterOperator")
}
