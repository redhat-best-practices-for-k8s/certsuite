// Copyright (C) 2020-2022 Red Hat, Inc.
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

package podrecreation

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ReplicaSetString  = "ReplicaSet"
	DeploymentString  = "Deployment"
	StatefulsetString = "StatefulSet"
	DaemonSetString   = "DaemonSet"
)

func CordonNode(name string) error {
	clients := clientsholder.GetClientsHolder()
	// Fetch node object
	node, err := clients.Coreclient.Nodes().Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		return err
	}

	node.Spec.Unschedulable = true

	// Update the node
	_, err = clients.Coreclient.Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})

	return err
}

func UncordonNode(name string) error {
	clients := clientsholder.GetClientsHolder()
	// Fetch node object
	node, err := clients.Coreclient.Nodes().Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		return err
	}

	node.Spec.Unschedulable = false

	// Update the node
	_, err = clients.Coreclient.Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})

	return err
}

func DeletePods(nodeName string) {
	clients := clientsholder.GetClientsHolder()
	pods, err := clients.Coreclient.Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("error getting list of pods err: %s", err)
	}
	for idx := range pods.Items {
		for _, or := range pods.Items[idx].OwnerReferences {
			if pods.Items[idx].Spec.NodeName == nodeName && or.Kind != DaemonSetString {
				err = clients.Coreclient.Pods(pods.Items[idx].Namespace).Delete(context.TODO(), pods.Items[idx].Name, metav1.DeleteOptions{})
				if err != nil {
					logrus.Errorf("error deleting pod %s err: %v", provider.PodToString(&pods.Items[idx]), err)
				}
			}
		}
	}
}

func CountPods(nodeName string) (count int) {
	count = 0
	clients := clientsholder.GetClientsHolder()
	pods, err := clients.Coreclient.Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("error getting list of pods err: %s", err)
	}
	for idx := range pods.Items {
		for _, or := range pods.Items[idx].OwnerReferences {
			if pods.Items[idx].Spec.NodeName == nodeName && or.Kind != DaemonSetString {
				count++
			}
		}
	}
	return count
}

func GetDeploymentNodes(pods []*v1.Pod, dName string) (nodes []string) {
	clients := clientsholder.GetClientsHolder()
	for _, put := range pods {
		deploymentFound := false
		for _, or := range put.OwnerReferences {
			if deploymentFound {
				break
			}
			if or.Kind == ReplicaSetString {
				r, err := clients.AppsClients.ReplicaSets(put.Namespace).Get(context.TODO(), or.Name, metav1.GetOptions{})
				if err != nil {
					logrus.Errorf("err: %s", err)
					continue
				}
				for _, or := range r.OwnerReferences {
					if or.Kind == DeploymentString && or.Name == dName {
						nodes = append(nodes, put.Spec.NodeName)
						deploymentFound = true
						break
					}
				}
			}
		}
	}
	return nodes
}

func GetStatefulsetNodes(pods []*v1.Pod, ssName string) (nodes []string) {
	for _, put := range pods {
		statefulsetFound := false
		for _, or := range put.OwnerReferences {
			if statefulsetFound {
				break
			}
			if or.Kind == StatefulsetString && or.Name == ssName {
				nodes = append(nodes, put.Spec.NodeName)
				statefulsetFound = true
				break
			}
		}
	}
	return nodes
}
