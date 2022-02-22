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
	"errors"
	"fmt"
	"strings"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/autodiscover"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	v1 "k8s.io/api/core/v1"
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

func GetContainer() *Container {
	return &Container{}
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
