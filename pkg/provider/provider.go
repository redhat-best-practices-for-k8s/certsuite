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

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/autodiscover"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	Csvs               []*v1alpha1.ClusterServiceVersion
	DebugPods          map[string]*v1.Pod // map from nodename to debugPod
	Config             configuration.TestConfiguration
	variables          configuration.TestParameters
	Crds               []*apiextv1.CustomResourceDefinition
	ContainersMap      map[*v1.Container]*Container
	MultusIPs          map[*v1.Pod]map[string][]string
	SkipNetTests       map[*v1.Pod]bool
	SkipMultusNetTests map[*v1.Pod]bool
	Deployments        []*v1apps.Deployment
	SatetfulSets       []*v1apps.StatefulSet
	HorizontalScaler   map[string]*v1scaling.HorizontalPodAutoscaler
	Subscriptions      []*v1alpha1.Subscription
	K8sVersion         string
	OpenshiftVersion   string
	HelmList           []*release.Release
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

func GetUpdatedStatefulSet(ac *appv1client.AppsV1Client, namespace, podName string) (*v1apps.StatefulSet, error) {
	return autodiscover.FindStateFulSetByNameByNamespace(ac, namespace, podName)
}

//nolint:funlen
func buildTestEnvironment() {
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
		for j := 0; j < len(pods[i].Spec.Containers); j++ {
			cut := &(pods[i].Spec.Containers[j])
			var state v1.ContainerStatus
			if len(pods[i].Status.ContainerStatuses) > 0 {
				state = pods[i].Status.ContainerStatuses[j]
			} else {
				logrus.Errorf("Pod %s is not ready, skipping status collection", PodToString(&pods[i]))
			}
			aRuntime, uid := GetRuntimeUID(&state)
			container := Container{Podname: pods[i].Name, Namespace: pods[i].Namespace,
				NodeName: pods[i].Spec.NodeName, Data: cut, Status: state, Runtime: aRuntime, UID: uid,
				ContainerImageIdentifier: buildContainerImageSource(pods[i].Spec.Containers[j].Image)}
			env.Containers = append(env.Containers, &container)
			env.ContainersMap[cut] = &container
		}
	}
	env.DebugPods = make(map[string]*v1.Pod)
	for i := 0; i < len(data.DebugPods); i++ {
		nodeName := data.DebugPods[i].Spec.NodeName
		env.DebugPods[nodeName] = &data.DebugPods[i]
	}
	csvs := data.Csvs
	subscriptions := data.Subscriptions
	for i := range csvs {
		env.Csvs = append(env.Csvs, &csvs[i])
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
		env.SatetfulSets = append(env.SatetfulSets, &data.StatefulSet[i])
	}
	env.HorizontalScaler = data.Hpas
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

func IsinstalledCsv(csv *v1alpha1.ClusterServiceVersion, subscriptions []v1alpha1.Subscription) (bool, v1alpha1.Subscription) {
	var returnsub v1alpha1.Subscription
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
	return fmt.Sprintf("ns: %s podName: %s containerName: %s",
		c.Namespace,
		c.Podname,
		c.Data.Name,
	)
}

func PodToString(pod *v1.Pod) string {
	return fmt.Sprintf("ns: %s podName: %s",
		pod.Namespace,
		pod.Name,
	)
}

func DeploymentToString(d *v1apps.Deployment) string {
	return fmt.Sprintf("ns: %s deployment name: %s",
		d.Namespace,
		d.Name,
	)
}

func StatefultsetToString(s *v1apps.StatefulSet) string {
	return fmt.Sprintf("ns: %s statefulset name: %s",
		s.Namespace,
		s.Name,
	)
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
