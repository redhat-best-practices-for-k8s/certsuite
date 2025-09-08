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

// WaitForStatefulSetReady waits until a StatefulSet reaches the ready state
//
// The function polls the Kubernetes API at one‑second intervals, retrieving
// the latest StatefulSet definition for the given namespace and name. It checks
// whether all replicas are available and the update is complete; if so it logs
// success and returns true. If the timeout expires before readiness, an error
// is logged and false is returned.
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

// isDeploymentReady checks if a deployment has finished rolling out
//
// The function retrieves the current state of a deployment in a given namespace
// using Kubernetes clients, then determines readiness by examining its status
// conditions. It returns true when all replicas are updated and available,
// otherwise false, along with any error that occurred during retrieval.
func isDeploymentReady(name, namespace string) (bool, error) {
	appsV1Api := clientsholder.GetClientsHolder().K8sClient.AppsV1()

	dep, err := provider.GetUpdatedDeployment(appsV1Api, namespace, name)
	if err != nil {
		return false, err
	}

	return dep.IsDeploymentReady(), nil
}

// isStatefulSetReady determines if a StatefulSet is fully ready
//
// The function retrieves the current state of a specified StatefulSet using
// Kubernetes client APIs, then checks whether all its replicas are available.
// It returns true when the StatefulSet meets readiness criteria or an error if
// retrieval fails.
func isStatefulSetReady(name, namespace string) (bool, error) {
	appsV1Api := clientsholder.GetClientsHolder().K8sClient.AppsV1()

	sts, err := provider.GetUpdatedStatefulset(appsV1Api, namespace, name)
	if err != nil {
		return false, err
	}

	return sts.IsStatefulSetReady(), nil
}

// getDeploymentsInfo Collects deployment identifiers as namespace:name strings
//
// The function iterates over a slice of deployment pointers, formatting each
// deployment’s namespace and name into a string separated by a colon. It
// appends these formatted strings to a new slice, which is then returned. This
// helper is used for logging or reporting purposes during test execution.
func getDeploymentsInfo(deployments []*provider.Deployment) []string {
	deps := []string{}
	for _, dep := range deployments {
		deps = append(deps, fmt.Sprintf("%s:%s", dep.Namespace, dep.Name))
	}

	return deps
}

// getStatefulSetsInfo creates a list of namespace:name strings for each StatefulSet
//
// The function iterates over the supplied slice, formatting each element’s
// namespace and name into a single string separated by a colon. These formatted
// strings are collected in a new slice which is then returned. The resulting
// slice provides a concise representation of the StatefulSets for logging or
// reporting purposes.
func getStatefulSetsInfo(statefulSets []*provider.StatefulSet) []string {
	stsInfo := []string{}
	for _, sts := range statefulSets {
		stsInfo = append(stsInfo, fmt.Sprintf("%s:%s", sts.Namespace, sts.Name))
	}

	return stsInfo
}

// getNotReadyDeployments identifies deployments that are not yet ready
//
// This helper inspects each deployment in the supplied slice, calling a
// readiness check for its name and namespace. Deployments reported as ready are
// omitted from the result; any errors during the check also cause the
// deployment to be considered not ready. The function returns a new slice
// containing only those deployments that failed the readiness test.
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

// getNotReadyStatefulSets filters stateful sets that are not ready
//
// The function iterates over a slice of stateful set objects, checking each
// one's readiness status via an external helper. If the check fails or
// indicates the set is not ready, it records the set in a new slice. The
// resulting slice contains only those stateful sets that are considered not
// ready, and this list is returned to the caller.
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

// WaitForAllPodSetsReady waits until all deployments and stateful sets are ready or a timeout occurs
//
// The function polls the readiness status of every deployment and stateful set
// in the test environment at fixed intervals, logging each check. It stops
// early if all podsets become ready before the specified duration; otherwise it
// returns the remaining not‑ready objects after the timeout. The returned
// slices allow callers to report which resources failed to reach readiness.
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

// GetAllNodesForAllPodSets Collects unique node names for pods owned by replicasets or statefulsets
//
// The function iterates over each pod and inspects its owner references. When
// it finds an owner of kind ReplicaSet or StatefulSet, the pod’s node name is
// added to a map that tracks distinct nodes. The resulting map contains one
// entry per node that hosts at least one such pod.
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
