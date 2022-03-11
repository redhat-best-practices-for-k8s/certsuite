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
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	retry "k8s.io/client-go/util/retry"
)

const (
	ReplicaSetString            = "ReplicaSet"
	DeploymentString            = "Deployment"
	StatefulsetString           = "StatefulSet"
	DaemonSetString             = "DaemonSet"
	DefaultGracePeriodInSeconds = 30
	Cordon                      = "cordon"
	Uncordon                    = "uncordon"
)

func CordonHelper(name, operation string) error {
	clients := clientsholder.GetClientsHolder()

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Fetch node object
		node, err := clients.Coreclient.Nodes().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		switch operation {
		case Cordon:
			node.Spec.Unschedulable = true
		case Uncordon:
			node.Spec.Unschedulable = false
		default:
			return fmt.Errorf("cordonHelper: Unsupported operation:%s", operation)
		}
		// Update the node
		_, err = clients.Coreclient.Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		logrus.Error("can't ", operation, " node: ", name, " error=", retryErr)
	}
	return retryErr
}

func CountPodsWithDelete(nodeName string, isDelete bool) (count int, err error) {
	clients := clientsholder.GetClientsHolder()
	pods, err := clients.Coreclient.Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName, LabelSelector: "pod-template-hash",
	})
	if err != nil {
		logrus.Errorf("error getting list of pods err: %s", err)
		return 0, err
	}
	count = 0
	for idx := range pods.Items {
		for _, or := range pods.Items[idx].OwnerReferences {
			if or.Kind == DaemonSetString {
				continue
			}
			count++
			if !isDelete {
				continue
			}
			logrus.Tracef("deleting %s", provider.PodToString(&pods.Items[idx]))
			deleteOptions := metav1.DeleteOptions{}
			gracePeriodSeconds := int64(DefaultGracePeriodInSeconds + time.Duration(*pods.Items[idx].Spec.TerminationGracePeriodSeconds))
			deleteOptions.GracePeriodSeconds = &gracePeriodSeconds

			err = clients.Coreclient.Pods(pods.Items[idx].Namespace).Delete(context.TODO(), pods.Items[idx].Name, deleteOptions)
			if err != nil {
				logrus.Errorf("error deleting %s err: %v", provider.PodToString(&pods.Items[idx]), err)
				return 0, err
			}
		}
	}
	return count, nil
}

func CordonCleanup(node string) {
	err := CordonHelper(node, Uncordon)
	if err != nil {
		logrus.Fatalf("cleanup: error uncordoning the node: %s", node)
	}
}
