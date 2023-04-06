// Copyright (C) 2020-2023 Red Hat, Inc.
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
	"github.com/test-network-function/cnf-certification-test/pkg/loghelper"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ReplicaSetString  = "ReplicaSet"
	StatefulsetString = "StatefulSet"
)

var WaitForDeploymentSetReady = func(ns, name string, timeout time.Duration) bool {
	logrus.Trace("check if deployment ", ns, ":", name, " is ready ")
	clients := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		dp, err := provider.GetUpdatedDeployment(clients.K8sClient.AppsV1(), ns, name)
		if err != nil {
			logrus.Errorf("Error while getting deployment %s (ns: %s), err: %v", name, ns, err)
		} else if !dp.IsDeploymentReady() {
			logrus.Infof("%s is not ready yet", dp.ToString())
		} else {
			logrus.Tracef("%s is ready!", dp.ToString())
			return true
		}

		time.Sleep(time.Second)
	}
	logrus.Error("deployment ", ns, ":", name, " is not ready ")
	return false
}

var WaitForScalingToComplete = func(ns, name string, timeout time.Duration, groupResourceSchema schema.GroupResource) bool {
	logrus.Trace("check if scale object for crs ", ns, ":", name, " is ready ")
	clients := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		crScale, err := provider.GetUpdatedCrObject(clients.ScalingClient, ns, name, groupResourceSchema)
		if err != nil {
			logrus.Errorf("error while getting the scaling fileds %e", err)
		} else if !crScale.IsScaleObjectReady() {
			logrus.Errorf("%s is not ready yet", crScale.ToString())
		} else {
			logrus.Tracef("%s is ready!", crScale.ToString())
			return true
		}

		time.Sleep(time.Second)
	}
	logrus.Error("timeout waiting for cr ", ns, ":", name, " scaling to be complete")
	return false
}

func WaitForStatefulSetReady(ns, name string, timeout time.Duration) bool {
	logrus.Trace("check if statefulset ", ns, ":", name, " is ready ")
	clients := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		ss, err := provider.GetUpdatedStatefulset(clients.K8sClient.AppsV1(), ns, name)
		if err == nil && ss.IsStatefulSetReady() {
			logrus.Tracef("%s is ready, err: %v", ss.ToString(), err)
			return true
		} else if err != nil {
			logrus.Errorf("Error while getting the %s, err: %v", ss.ToString(), err)
		}
		time.Sleep(time.Second)
	}
	logrus.Error("statefulset ", ns, ":", name, " is not ready ")
	return false
}

func WaitForAllPodSetReady(env *provider.TestEnvironment, timeoutPodSetReady time.Duration) (claimsLog loghelper.CuratedLogLines, atLeastOnePodsetNotReady bool) {
	atLeastOnePodsetNotReady = false
	for _, dut := range env.Deployments {
		isReady := WaitForDeploymentSetReady(dut.Namespace, dut.Name, timeoutPodSetReady)
		if isReady {
			claimsLog.AddLogLine("%s Status: OK", dut.ToString())
		} else {
			claimsLog.AddLogLine("%s Status: NOK", dut.ToString())
			atLeastOnePodsetNotReady = true
		}
	}
	for _, sut := range env.StatefulSets {
		isReady := WaitForStatefulSetReady(sut.Namespace, sut.Name, timeoutPodSetReady)
		if isReady {
			claimsLog.AddLogLine("%s Status: OK", sut.ToString())
		} else {
			claimsLog.AddLogLine("%s Status: NOK", sut.ToString())
			atLeastOnePodsetNotReady = true
		}
	}
	return claimsLog, atLeastOnePodsetNotReady
}

func GetAllNodesForAllPodSets(pods []*provider.Pod) (nodes map[string]bool) {
	nodes = make(map[string]bool)
	for _, put := range pods {
		for _, or := range put.OwnerReferences {
			if or.Kind != ReplicaSetString && or.Kind != StatefulsetString {
				continue
			}
			nodes[put.Spec.NodeName] = true
			break
		}
	}
	return nodes
}
