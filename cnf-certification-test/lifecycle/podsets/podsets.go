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

package podsets

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"

	v1app "k8s.io/api/apps/v1"
)

func WaitForDeploymentSetReady(ns, name string, timeout time.Duration) bool { //nolint:dupl // not duplicate
	logrus.Trace("check if deployment ", ns, ":", name, " is ready ")
	clients := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		dp, err := provider.GetUpdatedDeployment(clients.AppsClients, ns, name)
		if err == nil && IsDeploymentReady(dp) {
			logrus.Trace("deployment ", ns, ":", name, " is ready ")
			return true
		}
		time.Sleep(time.Second)
	}
	logrus.Error("deployment ", ns, ":", name, " is not ready ")
	return false
}
func IsDeploymentReady(deployment *v1app.Deployment) bool {
	notReady := true
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == v1app.DeploymentAvailable {
			notReady = false
			break
		}
	}
	var replicas int32
	if deployment.Spec.Replicas != nil {
		replicas = *(deployment.Spec.Replicas)
	} else {
		replicas = 1
	}
	if notReady ||
		deployment.Status.UnavailableReplicas != 0 ||
		deployment.Status.ReadyReplicas != replicas ||
		deployment.Status.AvailableReplicas != replicas ||
		deployment.Status.UpdatedReplicas != replicas {
		return false
	}
	return true
}
func WaitForStatefulSetReady(ns, name string, timeout time.Duration) bool { //nolint:dupl // not duplicate
	logrus.Trace("check if statefulset ", ns, ":", name, " is ready ")
	clients := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		ss, err := provider.GetUpdatedStatefulset(clients.AppsClients, ns, name)
		if err == nil && IsStatefulSetReady(ss) {
			logrus.Trace("statefulset ", ns, ":", name, " is ready ")
			return true
		}
		time.Sleep(time.Second)
	}
	logrus.Error("statefulset ", ns, ":", name, " is not ready ")
	return false
}
func IsStatefulSetReady(statefulset *v1app.StatefulSet) bool {
	var replicas int32
	if statefulset.Spec.Replicas != nil {
		replicas = *(statefulset.Spec.Replicas)
	} else {
		replicas = 1
	}
	if statefulset.Status.ReadyReplicas != replicas ||
		statefulset.Status.CurrentReplicas != replicas ||
		statefulset.Status.UpdatedReplicas != replicas {
		return false
	}
	return true
}
