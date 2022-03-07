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

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/autodiscover"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"helm.sh/helm/v3/pkg/release"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	daemonSetNamespace = "default"
	daemonSetName      = "debug"
	timeout            = 60 * time.Second
)

type TestEnvironment struct { // rename this with testTarget
	Namespaces       []string //
	Pods             []*v1.Pod
	Containers       []*Container
	Csvs             []*v1alpha1.ClusterServiceVersion
	DebugPods        map[string]*v1.Pod // map from nodename to debugPod
	Config           configuration.TestConfiguration
	variables        configuration.TestParameters
	Crds             []*apiextv1.CustomResourceDefinition
	Subscriptions    []*v1alpha1.Subscription
	K8sVersion       string
	OpenshiftVersion string
	HelmList         []*release.Release
}

type Container struct {
	Data                     v1.Container
	Status                   v1.ContainerStatus
	Namespace                string
	Podname                  string
	NodeName                 string
	ContainerImageIdentifier configuration.ContainerImageIdentifier
}

var (
	env    = TestEnvironment{}
	loaded = false
)

func GetContainer() *Container {
	return &Container{}
}

func BuildTestEnvironment() {
	// delete env
	env = TestEnvironment{}
	// build Pods and Containers under test
	data := autodiscover.DoAutoDiscover()
	env.Config = data.TestData
	env.Crds = data.Crds
	env.Namespaces = data.Namespaces
	env.variables = data.Env
	pods := data.Pods
	for i := 0; i < len(pods); i++ {
		env.Pods = append(env.Pods, &pods[i])
		for j := 0; j < len(pods[i].Spec.Containers); j++ {
			cut := pods[i].Spec.Containers[j]
			state := pods[i].Status.ContainerStatuses[j]
			container := Container{Podname: pods[i].Name, Namespace: pods[i].Namespace,
				NodeName: pods[i].Spec.NodeName, Data: cut, Status: state,
				ContainerImageIdentifier: buildContainerImageSource(pods[i].Spec.Containers[j].Image)}
			env.Containers = append(env.Containers, &container)
		}
	}
	debugPods := data.DebugPods
	env.DebugPods = make(map[string]*v1.Pod)
	for i := 0; i < len(debugPods); i++ {
		nodeName := debugPods[i].Spec.NodeName
		env.DebugPods[nodeName] = &debugPods[i]
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
		if len(raw) > 0 {
			for _, helm := range raw {
				if !isSkipHelmChart(helm.Name, data.TestData.SkipHelmChartList) {
					env.HelmList = append(env.HelmList, helm)
				}
			}
		}
	}

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
		BuildTestEnvironment()
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
	oc := clientsholder.NewClientsHolder()
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

func isDaemonSetReady(status *appsv1.DaemonSetStatus) (isReady bool) {
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

//nolint:gomnd
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
