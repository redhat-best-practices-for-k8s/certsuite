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
	"fmt"
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
			logrus.Errorf("error while getting the scaling fields %v", err)
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
	logrus.Trace("check if statefulset ", ns, ":", name, " is ready")
	clients := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		ss, err := provider.GetUpdatedStatefulset(clients.K8sClient.AppsV1(), ns, name)
		if err != nil {
			logrus.Errorf("error while getting the %s, err: %v", ss.ToString(), err)
		} else if ss.IsStatefulSetReady() {
			logrus.Tracef("%s is ready", ss.ToString())
			return true
		}
		time.Sleep(time.Second)
	}
	logrus.Error("statefulset ", ns, ":", name, " is not ready")
	return false
}

func isDeploymentReady(name, namespace string) (bool, error) {
	appsV1Api := clientsholder.GetClientsHolder().K8sClient.AppsV1()

	dep, err := provider.GetUpdatedDeployment(appsV1Api, namespace, name)
	if err != nil {
		return false, err
	}

	return dep.IsDeploymentReady(), nil
}

func isStatefulSetReady(name, namespace string) (bool, error) {
	appsV1Api := clientsholder.GetClientsHolder().K8sClient.AppsV1()

	sts, err := provider.GetUpdatedStatefulset(appsV1Api, namespace, name)
	if err != nil {
		return false, err
	}

	return sts.IsStatefulSetReady(), nil
}

// Helper function to get a slice of namespace:name strings from a slice of *provider.Deployments.
// E.g: [tnf:test tnf:hazelcast-platform-controller-manager]
func getDeploymentsInfo(deployments []*provider.Deployment) []string {
	deps := []string{}
	for _, dep := range deployments {
		deps = append(deps, fmt.Sprintf("%s:%s", dep.Namespace, dep.Name))
	}

	return deps
}

// Helper function to get a slice of namespace: name strings from a slice of *provider.Statefulsets.
func getStatefulSetsInfo(statefulSets []*provider.StatefulSet) []string {
	stsInfo := []string{}
	for _, sts := range statefulSets {
		stsInfo = append(stsInfo, fmt.Sprintf("%s:%s", sts.Namespace, sts.Name))
	}

	return stsInfo
}

// Helper function that checks the status of each deployment in the slice and returns
// a slice with the not-ready ones.
func getNotReadyDeployments(deployments []*provider.Deployment) []*provider.Deployment {
	notReadyDeployments := []*provider.Deployment{}
	for _, dep := range deployments {
		ready, err := isDeploymentReady(dep.Name, dep.Namespace)
		if err != nil {
			logrus.Errorf("Failed to get %s: %v", dep.ToString(), err)
			// We'll mark it as not ready, anyways.
			notReadyDeployments = append(notReadyDeployments, dep)
			continue
		}

		if ready {
			logrus.Debugf("%s is ready.", dep.ToString())
		} else {
			notReadyDeployments = append(notReadyDeployments, dep)
		}
	}

	return notReadyDeployments
}

// Helper function that checks the status of each statefulSet in the slice and returns
// a slice with the not-ready ones.
func getNotReadyStatefulSets(statefulSets []*provider.StatefulSet) []*provider.StatefulSet {
	notReadyStatefulSets := []*provider.StatefulSet{}
	for _, sts := range statefulSets {
		ready, err := isStatefulSetReady(sts.Name, sts.Namespace)
		if err != nil {
			logrus.Errorf("Failed to get %s: %v", sts.ToString(), err)
			// We'll mark it as not ready, anyways.
			notReadyStatefulSets = append(notReadyStatefulSets, sts)
			continue
		}

		if ready {
			logrus.Debugf("%s is ready.", sts.ToString())
		} else {
			notReadyStatefulSets = append(notReadyStatefulSets, sts)
		}
	}

	return notReadyStatefulSets
}

func WaitForAllPodSetsReady(env *provider.TestEnvironment, timeout time.Duration) (claimsLog loghelper.CuratedLogLines, atLeastOnePodsetNotReady bool) {
	const queryInterval = 15 * time.Second

	deploymentsToCheck := env.Deployments
	statefulSetsToCheck := env.StatefulSets

	logrus.Infof("Waiting %s for %d podsets to be ready.", timeout, len(deploymentsToCheck)+len(statefulSetsToCheck))
	for startTime := time.Now(); time.Since(startTime) < timeout; {
		logrus.Infof("Checking Deployments readiness of Deployments %v", getDeploymentsInfo(deploymentsToCheck))
		notReadyDeployments := getNotReadyDeployments(deploymentsToCheck)

		logrus.Infof("Checking StatefulSets readiness of StatefulSets %v", getStatefulSetsInfo(statefulSetsToCheck))
		notReadyStatefulSets := getNotReadyStatefulSets(statefulSetsToCheck)

		logrus.Infof("Not ready Deployments: %v", getDeploymentsInfo(notReadyDeployments))
		logrus.Infof("Not ready StatefulSets: %v", getStatefulSetsInfo(notReadyStatefulSets))

		deploymentsToCheck = notReadyDeployments
		statefulSetsToCheck = notReadyStatefulSets

		if len(deploymentsToCheck) == 0 && len(statefulSetsToCheck) == 0 {
			// No more podsets to check.
			break
		}

		time.Sleep(queryInterval)
	}

	// Here, either we reached the timeout or there's no more not-ready deployments or statefulsets.
	claimsLog.AddLogLine("Not ready Deployments: %v", getDeploymentsInfo(deploymentsToCheck))
	claimsLog.AddLogLine("Not ready StatefulSets: %v", getStatefulSetsInfo(statefulSetsToCheck))

	atLeastOnePodsetNotReady = len(deploymentsToCheck) > 0 || len(statefulSetsToCheck) > 0
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
