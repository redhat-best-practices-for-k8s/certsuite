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

package podrecreation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
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
	DeleteBackground            = "deleteBackground"
	DeleteForeground            = "deleteForeground"
	NoDelete                    = "noDelete"
)

// CordonHelper updates a node’s schedulability status.
//
// It retrieves the specified node, modifies its Unschedulable field based
// on the provided cordon type (e.g., "cordon" or "uncordon"), and then
// persists the change using a retry loop to handle conflicts.
// The function returns an error if any step fails.
func CordonHelper(name, operation string) error {
	clients := clientsholder.GetClientsHolder()

	log.Info("Performing %s operation on node %s", operation, name)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Fetch node object
		node, err := clients.K8sClient.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
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
		_, err = clients.K8sClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
		return err
	})
	if retryErr != nil {
		log.Error("can not %s node: %s, err=%v", operation, name, retryErr)
	}
	return retryErr
}

// CountPodsWithDelete counts how many pods in the provided list should be deleted based on the delete strategy.
//
// It iterates over each pod, skips daemonset pods, and applies the deletion logic defined by the strategy string.
// The function returns the count of pods that were processed for deletion and any error encountered during the operation.
func CountPodsWithDelete(pods []*provider.Pod, nodeName, mode string) (count int, err error) {
	count = 0
	var wg sync.WaitGroup
	for _, put := range pods {
		_, isDeployment := put.Labels["pod-template-hash"]
		_, isStatefulset := put.Labels["controller-revision-hash"]
		if put.Spec.NodeName == nodeName &&
			(isDeployment || isStatefulset) {
			if skipDaemonPod(put.Pod) {
				continue
			}
			count++
			if mode == NoDelete {
				continue
			}
			err := deletePod(put.Pod, mode, &wg)
			if err != nil {
				log.Error("Error deleting %s", put)
			}
		}
	}

	wg.Wait()
	return count, nil
}

// skipDaemonPod determines whether the given Pod should be ignored because it
// belongs to a DaemonSet.
//
// It accepts a pointer to a corev1.Pod and returns a boolean.
// The function checks the pod's owner references for an owning resource of
// kind "DaemonSet" and returns true if such an owner is found, indicating that
// this pod should be skipped during recreation logic.
func skipDaemonPod(pod *corev1.Pod) bool {
	for _, or := range pod.OwnerReferences {
		if or.Kind == DaemonSetString {
			return true
		}
	}
	return false
}

// deletePod deletes a pod and waits for its removal before signaling completion.
//
// It accepts a pointer to the pod to be deleted, the name of the deletion
// propagation policy (e.g., "Background" or "Foreground"), and a WaitGroup
// to synchronize concurrent deletions. The function issues a delete request
// via the Kubernetes client, then blocks until the pod is confirmed removed,
// reporting any errors encountered during deletion or waiting. On success,
// it marks the WaitGroup as done.
func deletePod(pod *corev1.Pod, mode string, wg *sync.WaitGroup) error {
	clients := clientsholder.GetClientsHolder()
	log.Debug("deleting ns=%s pod=%s with %s mode", pod.Namespace, pod.Name, mode)
	gracePeriodSeconds := *pod.Spec.TerminationGracePeriodSeconds
	// Create watcher before deleting pod
	watcher, err := clients.K8sClient.CoreV1().Pods(pod.Namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: "metadata.name=" + pod.Name + ",metadata.namespace=" + pod.Namespace,
	})
	if err != nil {
		return fmt.Errorf("waitPodDeleted ns=%s pod=%s, err=%s", pod.Namespace, pod.Name, err)
	}
	// Actually deleting pod
	err = clients.K8sClient.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	})
	if err != nil {
		log.Error("Error deleting %s err: %v", pod.String(), err)
		return err
	}
	if mode == DeleteBackground {
		return nil
	}
	wg.Add(1)
	podName := pod.Name
	namespace := pod.Namespace
	go func() {
		defer wg.Done()
		waitPodDeleted(namespace, podName, gracePeriodSeconds, watcher)
	}()
	return nil
}

// CordonCleanup restores a node's scheduling status after test cleanup.
//
// It takes the name of a node and a pointer to a checksdb.Check object,
// then uncordons the node if it was previously cordoned during testing.
// If the operation fails, the function aborts the current check with an error message.
func CordonCleanup(node string, check *checksdb.Check) {
	err := CordonHelper(node, Uncordon)
	if err != nil {
		check.Abort(fmt.Sprintf("cleanup: error uncordoning the node: %s, err=%s", node, err))
	}
}

// waitPodDeleted blocks until a pod with the given name in the specified namespace is deleted or an error occurs.
//
// It takes the pod name, namespace, and a timeout in seconds along with a watch.Interface that streams
// pod events. The function returns a closure that, when invoked, will wait for either a delete event for
// the target pod or for the timeout to elapse. If the pod is deleted before the timeout, the closure
// completes silently; otherwise it logs an error indicating that the pod was not removed in time.
func waitPodDeleted(ns, podName string, timeout int64, watcher watch.Interface) {
	log.Debug("Entering waitPodDeleted ns=%s pod=%s", ns, podName)
	defer watcher.Stop()

	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type == watch.Deleted || event.Type == "" {
				log.Debug("ns=%s pod=%s deleted", ns, podName)
				return
			}
		case <-time.After(time.Duration(timeout) * time.Second):
			log.Info("watch for pod deletion timedout after %d seconds", timeout)
			return
		}
	}
}
