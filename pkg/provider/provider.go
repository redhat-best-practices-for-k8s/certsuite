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

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/ocpclient"
	"github.com/test-network-function/cnf-certification-test/pkg/autodiscover"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
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
	Namespaces []string //
	Pods       []*v1.Pod
	Containers []*Container
	Csvs       []*v1alpha1.ClusterServiceVersion
	DebugPods  map[string]*v1.Pod // map from nodename to debugPod
	Config     configuration.TestConfiguration
	variables  configuration.TestParameters
	Crds       []*apiextv1.CustomResourceDefinition
}

type Container struct {
	Data      v1.Container
	Status    v1.ContainerStatus
	Namespace string
	Podname   string
	NodeName  string
}

var (
	env    = TestEnvironment{}
	loaded = false
)

func GetContainer(namespace, podName, containerName string) (v1.Container, error) {
	return v1.Container{}, nil
}

func GetPod(namespace, podName string) (v1.Pod, error) {
	return v1.Pod{}, nil
}

func BuildTestEnvironment() {
	// delete env
	env = TestEnvironment{}
	// build Pods and Containers under test
	environmentVariables, conf, pods, debugPods, crds, ns, csvs := autodiscover.DoAutoDiscover()
	env.Config = conf
	env.Crds = crds
	env.Namespaces = ns
	env.variables = environmentVariables
	for i := 0; i < len(pods); i++ {
		env.Pods = append(env.Pods, &pods[i])
		for j := 0; j < len(pods[i].Spec.Containers); j++ {
			cut := pods[i].Spec.Containers[j]
			state := pods[i].Status.ContainerStatuses[j]
			container := Container{Podname: pods[i].Name, Namespace: pods[i].Namespace,
				NodeName: pods[i].Spec.NodeName, Data: cut, Status: state}
			env.Containers = append(env.Containers, &container)
		}
	}
	env.DebugPods = make(map[string]*v1.Pod)
	for i := 0; i < len(debugPods); i++ {
		nodeName := debugPods[i].Spec.NodeName
		env.DebugPods[nodeName] = &debugPods[i]
	}

	for i := range csvs {
		env.Csvs = append(env.Csvs, &csvs[i])
	}
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

func WaitDebugPodReady() {
	oc := ocpclient.NewOcpClient()
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
		daemonSet, err := oc.AppsClient.DaemonSets(daemonSetNamespace).Get(context.TODO(), daemonSetName, getOptions)
		if err != nil && daemonSet != nil {
			logrus.Fatal("Error getting Daemonset, please create debug daemonset")
		}
		if daemonSet.Status.DesiredNumberScheduled != nodesCount {
			logrus.Fatalf("Daemonset DesiredNumberScheduled not equal to number of nodes:%d, please instantiate debug pods on all nodes", nodesCount)
		}
		if daemonSet.Status.DesiredNumberScheduled == daemonSet.Status.CurrentNumberScheduled && //nolint:gocritic
			daemonSet.Status.DesiredNumberScheduled == daemonSet.Status.NumberAvailable &&
			daemonSet.Status.DesiredNumberScheduled == daemonSet.Status.NumberReady &&
			daemonSet.Status.NumberMisscheduled == 0 {
			isReady = true
		}
		logrus.Debugf("Waiting for debug pods to be ready: %v", &daemonSet.Status)
		time.Sleep(time.Second)
	}
	if time.Since(start) > timeout {
		logrus.Fatal("Timeout waiting for Daemonset to be ready")
	}
	if isReady {
		logrus.Info("Daemonset is ready")
	}
}
