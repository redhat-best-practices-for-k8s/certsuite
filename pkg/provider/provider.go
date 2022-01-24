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
	"github.com/test-network-function/cnf-certification-test/pkg/autodiscover"
	v1 "k8s.io/api/core/v1"
)

type TestEnvironment struct { // rename this with testTarget
	Namespaces []string //
	Pods       []*v1.Pod
	Containers []*Container
	DebugPods  map[string]*v1.Pod // map from nodename to debugPod
}

type Container struct {
	Data      v1.Container
	Status    v1.ContainerStatus
	Namespace string
	Podname   string
	NodeName  string
}

type Context struct {
	Namespace     string
	Podname       string
	Containername string
}

var (
	env TestEnvironment
)

func GetContainer(namespace, podName, containerName string) (v1.Container, error) {
	return v1.Container{}, nil
}

func GetPod(namespace, podName string) (v1.Pod, error) {
	return v1.Pod{}, nil
}

func BuildTestEnvironment() {
	// build Pods and Containers under test
	pods, debugPods := autodiscover.DoAutoDiscover()
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
	for i := 0; i < len(debugPods); i++ {
		nodeName := debugPods[i].Spec.NodeName
		env.DebugPods[nodeName] = &debugPods[i]
	}
}

func GetTestEnvironment() TestEnvironment {
	return env
}
