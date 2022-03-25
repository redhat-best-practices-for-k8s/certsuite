// Copyright (C) 2020-2021 Red Hat, Inc.
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
	"time"

	"errors"
	"fmt"
	"strings"

	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/autodiscover"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	v1apps "k8s.io/api/apps/v1"
	v1scaling "k8s.io/api/autoscaling/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
)

const (
	daemonSetNamespace               = "default"
	daemonSetName                    = "debug"
	timeout                          = 60 * time.Second
	cniNetworksStatusKey             = "k8s.v1.cni.cncf.io/networks-status"
	skipConnectivityTestsLabel       = "test-network-function.com/skip_connectivity_tests"
	skipMultusConnectivityTestsLabel = "test-network-function.com/skip_multus_connectivity_tests"
)

type TestEnvironment struct { // rename this with testTarget
	Namespaces         []string
	Pods               []*v1.Pod
	Containers         []*Container
	Operators          []Operator
	DebugPods          map[string]*v1.Pod // map from nodename to debugPod
	Config             configuration.TestConfiguration
	variables          configuration.TestParameters
	Crds               []*apiextv1.CustomResourceDefinition
	ContainersMap      map[*v1.Container]*Container
	MultusIPs          map[*v1.Pod]map[string][]string
	SkipNetTests       map[*v1.Pod]bool
	SkipMultusNetTests map[*v1.Pod]bool
	Deployments        []*v1apps.Deployment
	StatetfulSets      []*v1apps.StatefulSet
	HorizontalScaler   map[string]*v1scaling.HorizontalPodAutoscaler
	Nodes              *v1.NodeList
	Subscriptions      []*olmv1Alpha.Subscription
	K8sVersion         string
	OpenshiftVersion   string
	HelmList           []*release.Release
}

type CsvInstallPlan struct {
	// Operator's installPlan name
	Name string `yaml:"name" json:"name"`
	// BundleImage is the URL referencing the bundle image
	BundleImage string `yaml:"bundleImage" json:"bundleImage"`
	// IndexImage is the URL referencing the index image
	IndexImage string `yaml:"indexImage" json:"indexImage"`
}

type Operator struct {
	Name             string                            `yaml:"name" json:"name"`
	Namespace        string                            `yaml:"namespace" json:"namespace"`
	Csv              *olmv1Alpha.ClusterServiceVersion `yaml:"csv" json:"csv"`
	SubscriptionName string                            `yaml:"subscriptionName" json:"subscriptionName"`
	InstallPlans     []CsvInstallPlan                  `yaml:"installPlans,omitempty" json:"installPlans,omitempty"`
	Package          string                            `yaml:"packag" json:"packag"`
	Org              string                            `yaml:"Org" json:"Org"`
	Version          string                            `yaml:"Version" json:"Version"`
}
type Container struct {
	Data                     *v1.Container
	Status                   v1.ContainerStatus
	Namespace                string
	Podname                  string
	NodeName                 string
	Runtime                  string
	UID                      string
	ContainerImageIdentifier configuration.ContainerImageIdentifier
}
type cniNetworkInterface struct {
	Name      string                 `json:"name"`
	Interface string                 `json:"interface"`
	IPs       []string               `json:"ips"`
	Default   bool                   `json:"default"`
	DNS       map[string]interface{} `json:"dns"`
}

var (
	env    = TestEnvironment{}
	loaded = false
)

func GetContainer() *Container {
	return &Container{}
}

func GetUpdatedDeployment(ac *appv1client.AppsV1Client, namespace, podName string) (*v1apps.Deployment, error) {
	return autodiscover.FindDeploymentByNameByNamespace(ac, namespace, podName)
}
func GetUpdatedStatefulset(ac *appv1client.AppsV1Client, namespace, podName string) (*v1apps.StatefulSet, error) {
	return autodiscover.FindStatefulsetByNameByNamespace(ac, namespace, podName)
}

func buildTestEnvironment() { //nolint:funlen
	// delete env
	env = TestEnvironment{}
	// build Pods and Containers under test
	data := autodiscover.DoAutoDiscover()
	env.Config = data.TestData
	env.Crds = data.Crds
	env.Namespaces = data.Namespaces
	env.variables = data.Env
	env.ContainersMap = make(map[*v1.Container]*Container)
	env.MultusIPs = make(map[*v1.Pod]map[string][]string)
	env.SkipNetTests = make(map[*v1.Pod]bool)
	env.SkipMultusNetTests = make(map[*v1.Pod]bool)
	env.Nodes = data.Nodes
	pods := data.Pods

	for i := 0; i < len(pods); i++ {
		env.Pods = append(env.Pods, &pods[i])
		var err error
		env.MultusIPs[&pods[i]], err = getPodIPsPerNet(pods[i].GetAnnotations()[cniNetworksStatusKey])
		if err != nil {
			logrus.Errorf("Could not decode networks-status annotation")
		}
		if pods[i].GetLabels()[skipConnectivityTestsLabel] != "" {
			env.SkipNetTests[&pods[i]] = true
		}
		if pods[i].GetLabels()[skipMultusConnectivityTestsLabel] != "" {
			env.SkipMultusNetTests[&pods[i]] = true
		}
		env.Containers = append(env.Containers, getPodContainers(&pods[i])...)
	}
	env.ContainersMap = createContainersMapByNode(env.Containers)
	env.DebugPods = make(map[string]*v1.Pod)
	for i := 0; i < len(data.DebugPods); i++ {
		nodeName := data.DebugPods[i].Spec.NodeName
		env.DebugPods[nodeName] = &data.DebugPods[i]
	}
	csvs := data.Csvs
	subscriptions := data.Subscriptions
	for i := range csvs {
		isCsv, sub := IsinstalledCsv(&csvs[i], subscriptions)
		if isCsv {
			env.Subscriptions = append(env.Subscriptions, &sub)
		}
	}
	env.OpenshiftVersion = data.OpenshiftVersion
	env.K8sVersion = data.K8sVersion
	helmList := data.HelmList
	for _, raw := range helmList {
		for _, helm := range raw {
			if !isSkipHelmChart(helm.Name, data.TestData.SkipHelmChartList) {
				env.HelmList = append(env.HelmList, helm)
			}
		}
	}
	for i := range data.Deployments {
		env.Deployments = append(env.Deployments, &data.Deployments[i])
	}
	for i := range data.StatefulSet {
		env.StatetfulSets = append(env.StatetfulSets, &data.StatefulSet[i])
	}
	env.HorizontalScaler = data.Hpas

	operators, err := createOperators(data.Csvs, data.Subscriptions)
	if err != nil {
		logrus.Errorf("Failed to get cluster operators: %s", err)
	}
	env.Operators = operators
}

func getPodContainers(aPod *v1.Pod) (containerList []*Container) {
	for j := 0; j < len(aPod.Spec.Containers); j++ {
		cut := &(aPod.Spec.Containers[j])
		var state v1.ContainerStatus
		if len(aPod.Status.ContainerStatuses) > 0 {
			state = aPod.Status.ContainerStatuses[j]
		} else {
			logrus.Errorf("%s is not ready, skipping status collection", PodToString(aPod))
		}
		aRuntime, uid := GetRuntimeUID(&state)
		container := Container{Podname: aPod.Name, Namespace: aPod.Namespace,
			NodeName: aPod.Spec.NodeName, Data: cut, Status: state, Runtime: aRuntime, UID: uid,
			ContainerImageIdentifier: buildContainerImageSource(aPod.Spec.Containers[j].Image)}
		containerList = append(containerList, &container)
	}
	return containerList
}

func createContainersMapByNode(containerList []*Container) (containersMap map[*v1.Container]*Container) {
	containersMap = make(map[*v1.Container]*Container)
	for _, c := range containerList {
		containersMap[c.Data] = c
	}
	return containersMap
}

func isSkipHelmChart(helmName string, skipHelmChartList []configuration.SkipHelmChartList) bool {
	if len(skipHelmChartList) == 0 {
		return false
	}
	for _, helm := range skipHelmChartList {
		if helmName == helm.Name {
			logrus.Infof("Helm chart with name %s was skipped", helmName)
			return true
		}
	}
	return false
}

func GetTestEnvironment() TestEnvironment {
	if !loaded {
		buildTestEnvironment()
		loaded = true
	}
	return env
}

func IsOCPCluster() bool {
	return !env.variables.NonOcpCluster
}

func IsinstalledCsv(csv *olmv1Alpha.ClusterServiceVersion, subscriptions []olmv1Alpha.Subscription) (bool, olmv1Alpha.Subscription) {
	var returnsub olmv1Alpha.Subscription
	for i := range subscriptions {
		if subscriptions[i].Status.InstalledCSV == csv.Name {
			returnsub = subscriptions[i]
			return true, returnsub
		}
	}
	return false, returnsub
}
func WaitDebugPodReady() {
	oc := clientsholder.GetClientsHolder()
	listOptions := metav1.ListOptions{}
	nodes, err := oc.Coreclient.Nodes().List(context.TODO(), listOptions)

	if err != nil {
		logrus.Fatalf("Error getting node list, err:%s", err)
	}

	nodesCount := int32(len(nodes.Items))

	getOptions := metav1.GetOptions{}
	isReady := false
	start := time.Now()
	for !isReady && time.Since(start) < timeout {
		daemonSet, err := oc.AppsClients.DaemonSets(daemonSetNamespace).Get(context.TODO(), daemonSetName, getOptions)
		if err != nil && daemonSet != nil {
			logrus.Fatal("Error getting Daemonset, please create debug daemonset")
		}
		if daemonSet.Status.DesiredNumberScheduled != nodesCount {
			logrus.Fatalf("Daemonset DesiredNumberScheduled not equal to number of nodes:%d, please instantiate debug pods on all nodes", nodesCount)
		}
		isReady = isDaemonSetReady(&daemonSet.Status)
		logrus.Debugf("Waiting for debug pods to be ready: %v", daemonSet.Status)
		time.Sleep(time.Second)
	}
	if time.Since(start) > timeout {
		logrus.Fatal("Timeout waiting for Daemonset to be ready")
	}
	if isReady {
		logrus.Info("Daemonset is ready")
	}
}

func isDaemonSetReady(status *v1apps.DaemonSetStatus) (isReady bool) {
	isReady = false
	if status.DesiredNumberScheduled == status.CurrentNumberScheduled && //nolint:gocritic
		status.DesiredNumberScheduled == status.NumberAvailable &&
		status.DesiredNumberScheduled == status.NumberReady &&
		status.NumberMisscheduled == 0 {
		isReady = true
	}
	return isReady
}
func (c *Container) GetUID() (string, error) {
	split := strings.Split(c.Status.ContainerID, "://")
	uid := ""
	if len(split) > 0 {
		uid = split[len(split)-1]
	}
	if uid == "" {
		logrus.Debugln(fmt.Sprintf("could not find uid of %s/%s/%s\n", c.Namespace, c.Podname, c.Data.Name))
		return "", errors.New("cannot determine container UID")
	}
	logrus.Debugln(fmt.Sprintf("uid of %s/%s/%s=%s\n", c.Namespace, c.Podname, c.Data.Name, uid))
	return uid, nil
}

func buildContainerImageSource(url string) configuration.ContainerImageIdentifier {
	source := configuration.ContainerImageIdentifier{}
	urlSegments := strings.Split(url, "/")
	n := len(urlSegments)
	if n > 1 {
		source.Repository = urlSegments[n-2]
	}
	colonIndex := strings.Index(urlSegments[n-1], ":")
	atIndex := strings.Index(urlSegments[n-1], "@")
	if atIndex == -1 {
		if colonIndex == -1 {
			source.Name = urlSegments[n-1]
		} else {
			source.Name = urlSegments[n-1][:colonIndex]
			source.Tag = urlSegments[n-1][colonIndex+1:]
		}
	} else {
		source.Name = urlSegments[n-1][:atIndex]
		source.Digest = urlSegments[n-1][atIndex+1:]
	}
	return source
}
func GetRuntimeUID(cs *v1.ContainerStatus) (runtime, uid string) {
	split := strings.Split(cs.ContainerID, "://")
	if len(split) > 0 {
		uid = split[len(split)-1]
		runtime = split[0]
	}
	return runtime, uid
}

func (c *Container) String() string {
	return fmt.Sprintf("node: %s ns: %s podName: %s containerName: %s containerUID: %s containerRuntime: %s",
		c.NodeName,
		c.Namespace,
		c.Podname,
		c.Data.Name,
		c.Status.ContainerID,
		c.Runtime,
	)
}
func (c *Container) StringShort() string {
	return fmt.Sprintf("container: %s pod: %s ns: %s",
		c.Data.Name,
		c.Podname,
		c.Namespace,
	)
}

func PodToString(pod *v1.Pod) string {
	return fmt.Sprintf("pod: %s ns: %s",
		pod.Name,
		pod.Namespace,
	)
}

func DeploymentToString(d *v1apps.Deployment) string {
	return fmt.Sprintf("deployment: %s ns: %s",
		d.Name,
		d.Namespace,
	)
}

func StatefulsetToString(s *v1apps.StatefulSet) string {
	return fmt.Sprintf("statefulset: %s ns: %s",
		s.Name,
		s.Namespace,
	)
}

func CsvToString(csv *olmv1Alpha.ClusterServiceVersion) string {
	return fmt.Sprintf("operator csv: %s ns: %s",
		csv.Name,
		csv.Namespace,
	)
}

func (op *Operator) String() string {
	return fmt.Sprintf("csv: %s ns:%s subscription:%s", op.Name, op.Namespace, op.SubscriptionName)
}

// getPodIPsPerNet gets the IPs of a pod.
// CNI annotation "k8s.v1.cni.cncf.io/networks-status".
// Returns (ips, error).
func getPodIPsPerNet(annotation string) (ips map[string][]string, err error) {
	// This is a map indexed with the network name (network attachment) and
	// listing all the IPs created in this subnet and belonging to the pod namespace
	// The list of ips pr net is parsed from the content of the "k8s.v1.cni.cncf.io/networks-status" annotation.
	ips = make(map[string][]string)

	var cniInfo []cniNetworkInterface
	err = json.Unmarshal([]byte(annotation), &cniInfo)
	if err != nil {
		return nil, errors.New("could not unmarshal network-status annotation")
	}
	// If this is the default interface, skip it as it is tested separately
	// Otherwise add all non default interfaces
	for _, cniInterface := range cniInfo {
		if !cniInterface.Default {
			ips[cniInterface.Name] = cniInterface.IPs
		}
	}
	return ips, nil
}

func (env *TestEnvironment) SetNeedsRefresh() {
	loaded = false
}

func (env *TestEnvironment) IsIntrusive() bool {
	return !env.variables.NonIntrusiveOnly
}

// getInstallPlansInNamespace is a helper function to get the installPlans in a namespace. The
// map installPlans is used to store them in order to avoid repeating http requests for a namespace
// whose installPlans were already obtained.
func getInstallPlansInNamespace(namespace string, clusterInstallPlans map[string][]olmv1Alpha.InstallPlan) ([]olmv1Alpha.InstallPlan, error) {
	// Check if installplans were stored before.
	nsInstallPlans, exist := clusterInstallPlans[namespace]
	if exist {
		return nsInstallPlans, nil
	}

	clients := clientsholder.GetClientsHolder()
	installPlanList, err := clients.OlmClient.OperatorsV1alpha1().InstallPlans(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable get installplans in namespace %s, err: %s", namespace, err)
	}

	nsInstallPlans = installPlanList.Items
	clusterInstallPlans[namespace] = nsInstallPlans

	return nsInstallPlans, nil
}

// getCsvInstallPlans is a helper function that returns the installPlans for a CSV in a namespace.
// The map clusterInstallPlans is used to store previously retrieved installPlans, in order to save
// http requests.
func getCsvInstallPlans(namespace, csv string, clusterInstallPlans map[string][]olmv1Alpha.InstallPlan) ([]*olmv1Alpha.InstallPlan, error) {
	nsInstallPlans, err := getInstallPlansInNamespace(namespace, clusterInstallPlans)
	if err != nil {
		return nil, err
	}

	installPlans := []*olmv1Alpha.InstallPlan{}
	for i := range nsInstallPlans {
		nsInstallPlan := &nsInstallPlans[i]
		for _, csvName := range nsInstallPlan.Spec.ClusterServiceVersionNames {
			if csv != csvName {
				continue
			}

			if nsInstallPlan.Status.BundleLookups == nil {
				logrus.Warnf("InstallPlan %s for csv %s (ns %s) does not have bundle lookups. It will be skipped.", nsInstallPlan.Name, csv, namespace)
				continue
			}

			installPlans = append(installPlans, nsInstallPlan)
		}
	}

	if len(installPlans) == 0 {
		return nil, fmt.Errorf("no installplans found for csv %s (ns %s)", csv, namespace)
	}

	return installPlans, nil
}

func getCatalogSourceImageIndexFromInstallPlan(installPlan *olmv1Alpha.InstallPlan) (string, error) {
	// ToDo/Technical debt: what to do if installPlan has more than one BundleLookups entries.
	catalogSourceName := installPlan.Status.BundleLookups[0].CatalogSourceRef.Name
	catalogSourceNamespace := installPlan.Status.BundleLookups[0].CatalogSourceRef.Namespace

	clients := clientsholder.GetClientsHolder()
	catalogSource, err := clients.OlmClient.OperatorsV1alpha1().CatalogSources(catalogSourceNamespace).Get(context.TODO(), catalogSourceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get catalogsource: %s", err)
	}

	return catalogSource.Spec.Image, nil
}

func createOperators(csvs []olmv1Alpha.ClusterServiceVersion, subscriptions []olmv1Alpha.Subscription) ([]Operator, error) {
	installPlans := map[string][]olmv1Alpha.InstallPlan{} // Helper: maps a namespace name to all its installplans.
	operators := []Operator{}
	for i := range csvs {
		csv := &csvs[i]
		op := Operator{Name: csv.Name, Namespace: csv.Namespace, Csv: csv}

		packageAndVersion := strings.SplitN(csv.Name, ".", 2) //nolint:gomnd // ok
		op.Version = packageAndVersion[1]

		for s := range subscriptions {
			subscription := &subscriptions[s]
			if subscription.Status.InstalledCSV != csv.Name {
				continue
			}

			op.SubscriptionName = subscription.Name
			op.Package = subscription.Spec.Package
			op.Org = subscription.Spec.CatalogSource

			csvInstallPlans, err := getCsvInstallPlans(csv.Namespace, csv.Name, installPlans)
			if err != nil {
				return nil, fmt.Errorf("failed to get installPlans for csv %s (ns %s)", csv.Name, csv.Namespace)
			}

			for _, installPlan := range csvInstallPlans {
				indexImage, err := getCatalogSourceImageIndexFromInstallPlan(installPlan)
				if err != nil {
					return nil, fmt.Errorf("failed to get installPlan image index for csv %s (ns %s) installPlan %s, err: %s",
						csv.Name, csv.Namespace, installPlan.Name, err)
				}

				op.InstallPlans = append(op.InstallPlans, CsvInstallPlan{
					Name:        installPlan.Name,
					BundleImage: installPlan.Status.BundleLookups[0].Path,
					IndexImage:  indexImage,
				})
			}
		}

		operators = append(operators, op)
	}

	return operators, nil
}
