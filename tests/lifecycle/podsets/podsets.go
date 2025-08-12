// Copyright (C) 2020-2024 Red Hat, Inc.
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

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ReplicaSetString  = "ReplicaSet"
	StatefulsetString = "StatefulSet"
)

var WaitForDeploymentSetReady = func(ns, name string, timeout time.Duration, logger *log.Logger) bool {
	logger.Info("Check if Deployment %s:%s is ready", ns, name)
	clients := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		dp, err := provider.GetUpdatedDeployment(clients.K8sClient.AppsV1(), ns, name)
		if err != nil {
			logger.Error("Error while getting Deployment %q, err: %v", name, err)
		} else if !dp.IsDeploymentReady() {
			logger.Warn("Deployment %q is not ready yet", dp.ToString())
		} else {
			logger.Info("Deployment %q is ready!", dp.ToString())
			return true
		}

		time.Sleep(time.Second)
	}
	logger.Error("Deployment %s:%s is not ready", ns, name)
	return false
}

var WaitForScalingToComplete = func(ns, name string, timeout time.Duration, groupResourceSchema schema.GroupResource, logger *log.Logger) bool {
	logger.Info("Check if scale object for CRs %s:%s is ready", ns, name)
	clients := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		crScale, err := provider.GetUpdatedCrObject(clients.ScalingClient, ns, name, groupResourceSchema)
		if err != nil {
			logger.Error("Error while getting the scaling fields %v", err)
		} else if !crScale.IsScaleObjectReady() {
			logger.Warn("%s is not ready yet", crScale.ToString())
		} else {
			logger.Info("%s is ready!", crScale.ToString())
			return true
		}

		time.Sleep(time.Second)
	}
	logger.Error("Timeout waiting for CR %s:%s scaling to be complete", ns, name)
	return false
}

// WaitForStatefulSetReady waits until the specified StatefulSet is fully ready or a timeout occurs.
//
// It repeatedly checks the readiness status of the StatefulSet identified by the
// provided namespace and name, sleeping between attempts. The function logs
// debug information using the supplied logger and returns true if the
// StatefulSet becomes ready before the timeout duration expires; otherwise it
// returns false. The parameters are: a string for the namespace, a string for
// the StatefulSet name, a time.Duration specifying the maximum wait time,
// and a pointer to log.Logger for logging output.
func WaitForStatefulSetReady(ns, name string, timeout time.Duration, logger *log.Logger) bool {
	logger.Debug("Check if statefulset %s:%s is ready", ns, name)
	clients := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		ss, err := provider.GetUpdatedStatefulset(clients.K8sClient.AppsV1(), ns, name)
		if err != nil {
			logger.Error("Error while getting the %s, err: %v", ss.ToString(), err)
		} else if ss.IsStatefulSetReady() {
			logger.Info("%s is ready", ss.ToString())
			return true
		}
		time.Sleep(time.Second)
	}
	logger.Error("Statefulset %s:%s is not ready", ns, name)
	return false
}

// isDeploymentReady checks if a deployment is fully rolled out.
//
// It takes the namespace and deployment name, retrieves the latest
// deployment configuration, evaluates its status via IsDeploymentReady,
// and returns true when the desired replica count matches the available
// replicas. If an error occurs during retrieval or evaluation it is
// returned along with false.
func isDeploymentReady(name, namespace string) (bool, error) {
	appsV1Api := clientsholder.GetClientsHolder().K8sClient.AppsV1()

	dep, err := provider.GetUpdatedDeployment(appsV1Api, namespace, name)
	if err != nil {
		return false, err
	}

	return dep.IsDeploymentReady(), nil
}

// isStatefulSetReady checks whether a StatefulSet is fully ready.
//
// It takes the namespace and name of a StatefulSet as input,
// retrieves the latest state from the Kubernetes API, and evaluates
// its readiness by examining replica counts and conditions.
// The function returns true if the StatefulSet meets all readiness criteria,
// or false otherwise. An error is returned if any API call fails or the
// StatefulSet cannot be retrieved.
func isStatefulSetReady(name, namespace string) (bool, error) {
	appsV1Api := clientsholder.GetClientsHolder().K8sClient.AppsV1()

	sts, err := provider.GetUpdatedStatefulset(appsV1Api, namespace, name)
	if err != nil {
		return false, err
	}

	return sts.IsStatefulSetReady(), nil
}

// getDeploymentsInfo returns a slice of namespace:name strings from deployments.
//
// It takes a slice of *provider.Deployment pointers and constructs
// a string for each deployment in the form "namespace:name".
// The returned slice contains one entry per deployment, suitable for
// use in logging or comparison operations.
func getDeploymentsInfo(deployments []*provider.Deployment) []string {
	deps := []string{}
	for _, dep := range deployments {
		deps = append(deps, fmt.Sprintf("%s:%s", dep.Namespace, dep.Name))
	}

	return deps
}

// getStatefulSetsInfo returns a slice of namespace:name strings for each StatefulSet.
//
// It takes a slice of *provider.StatefulSet pointers, extracts the Namespace and Name
// from each object, formats them as "namespace: name" using Sprintf,
// and appends each formatted string to a result slice which is then returned.
func getStatefulSetsInfo(statefulSets []*provider.StatefulSet) []string {
	stsInfo := []string{}
	for _, sts := range statefulSets {
		stsInfo = append(stsInfo, fmt.Sprintf("%s:%s", sts.Namespace, sts.Name))
	}

	return stsInfo
}

// getNotReadyDeployments returns a slice of deployments that are not ready.
//
// It iterates over the provided deployments, checks each one with isDeploymentReady,
// and collects those that fail into a new slice which it then returns.
func getNotReadyDeployments(deployments []*provider.Deployment) []*provider.Deployment {
	notReadyDeployments := []*provider.Deployment{}
	for _, dep := range deployments {
		ready, err := isDeploymentReady(dep.Name, dep.Namespace)
		if err != nil {
			log.Error("Failed to get %s: %v", dep.ToString(), err)
			// We'll mark it as not ready, anyways.
			notReadyDeployments = append(notReadyDeployments, dep)
			continue
		}

		if ready {
			log.Debug("%s is ready.", dep.ToString())
		} else {
			notReadyDeployments = append(notReadyDeployments, dep)
		}
	}

	return notReadyDeployments
}

// getNotReadyStatefulSets returns the subset of stateful sets that are not ready.
//
// It iterates over the provided slice of StatefulSet pointers, checks each one
// with isStatefulSetReady, and appends those that are not ready to a new slice.
// The resulting slice contains only the stateful sets whose status does not meet
// readiness criteria. The function returns this filtered slice.
func getNotReadyStatefulSets(statefulSets []*provider.StatefulSet) []*provider.StatefulSet {
	notReadyStatefulSets := []*provider.StatefulSet{}
	for _, sts := range statefulSets {
		ready, err := isStatefulSetReady(sts.Name, sts.Namespace)
		if err != nil {
			log.Error("Failed to get %s: %v", sts.ToString(), err)
			// We'll mark it as not ready, anyways.
			notReadyStatefulSets = append(notReadyStatefulSets, sts)
			continue
		}

		if ready {
			log.Debug("%s is ready.", sts.ToString())
		} else {
			notReadyStatefulSets = append(notReadyStatefulSets, sts)
		}
	}

	return notReadyStatefulSets
}

// WaitForAllPodSetsReady waits until all deployments and stateful sets in the test environment are ready or the timeout expires.
//
// It polls the current status of deployments and stateful sets, logging progress at each interval.
// The function returns two slices: one containing ready deployments and another containing ready stateful sets.
// If the timeout is reached before readiness, it logs an error and still returns the last observed ready objects.
func WaitForAllPodSetsReady(env *provider.TestEnvironment, timeout time.Duration, logger *log.Logger) (
	notReadyDeployments []*provider.Deployment,
	notReadyStatefulSets []*provider.StatefulSet) {
	const queryInterval = 15 * time.Second

	deploymentsToCheck := env.Deployments
	statefulSetsToCheck := env.StatefulSets

	logger.Info("Waiting %s for %d podsets to be ready.", timeout, len(deploymentsToCheck)+len(statefulSetsToCheck))
	for startTime := time.Now(); time.Since(startTime) < timeout; {
		logger.Info("Checking Deployments readiness of Deployments %v", getDeploymentsInfo(deploymentsToCheck))
		notReadyDeployments = getNotReadyDeployments(deploymentsToCheck)

		logger.Info("Checking StatefulSets readiness of StatefulSets %v", getStatefulSetsInfo(statefulSetsToCheck))
		notReadyStatefulSets = getNotReadyStatefulSets(statefulSetsToCheck)

		logger.Info("Not ready Deployments: %v", getDeploymentsInfo(notReadyDeployments))
		logger.Info("Not ready StatefulSets: %v", getStatefulSetsInfo(notReadyStatefulSets))

		deploymentsToCheck = notReadyDeployments
		statefulSetsToCheck = notReadyStatefulSets

		if len(deploymentsToCheck) == 0 && len(statefulSetsToCheck) == 0 {
			// No more podsets to check.
			break
		}

		time.Sleep(queryInterval)
	}

	// Here, either we reached the timeout or there's no more not-ready deployments or statefulsets.
	logger.Error("Not ready Deployments: %v", getDeploymentsInfo(deploymentsToCheck))
	logger.Error("Not ready StatefulSets: %v", getStatefulSetsInfo(statefulSetsToCheck))

	return deploymentsToCheck, statefulSetsToCheck
}

// GetAllNodesForAllPodSets returns a set of node names that host any pod in the provided slice.
//
// It accepts a slice of pointers to provider.Pod and iterates over each pod’s node assignment.
// For every pod, it records the node name in a map with boolean values,
// ensuring that each node appears only once. The resulting map keys represent
// all distinct nodes across the entire collection of pods.
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
