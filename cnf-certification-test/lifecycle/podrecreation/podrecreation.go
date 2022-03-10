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
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ReplicaSetString            = "ReplicaSet"
	DeploymentString            = "Deployment"
	StatefulsetString           = "StatefulSet"
	DaemonSetString             = "DaemonSet"
	DefaultGracePeriodInSeconds = 30
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

func DeletePods(nodeName string) (err error) {
	clients := clientsholder.GetClientsHolder()
	pods, err := clients.Coreclient.Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName, LabelSelector: "pod-template-hash",
	})
	if err != nil {
		logrus.Errorf("error getting list of pods err: %s", err)
		return err
	}
	for idx := range pods.Items {
		for _, or := range pods.Items[idx].OwnerReferences {
			if or.Kind == DaemonSetString {
				continue
			}
			logrus.Tracef("deleting pod %s", provider.PodToString(&pods.Items[idx]))
			deleteOptions := metav1.DeleteOptions{}
			gracePeriodSeconds := int64(DefaultGracePeriodInSeconds + time.Duration(*pods.Items[idx].Spec.TerminationGracePeriodSeconds))
			deleteOptions.GracePeriodSeconds = &gracePeriodSeconds

			err = clients.Coreclient.Pods(pods.Items[idx].Namespace).Delete(context.TODO(), pods.Items[idx].Name, deleteOptions)
			if err != nil {
				logrus.Errorf("error deleting pod %s err: %v", provider.PodToString(&pods.Items[idx]), err)
				return err
			}
		}
	}
	return nil
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
