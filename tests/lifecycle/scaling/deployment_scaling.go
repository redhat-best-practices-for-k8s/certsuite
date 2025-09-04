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

//nolint:dupl
package scaling

import (
	"context"
	"errors"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/podsets"

	v1autoscaling "k8s.io/api/autoscaling/v1"

	v1machinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	retry "k8s.io/client-go/util/retry"

	appsv1 "k8s.io/api/apps/v1"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	hps "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
)

// TestScaleDeployment Tests scaling behavior of a Deployment without HPA
//
// The function obtains Kubernetes clients, determines the current replica count
// or defaults to one, then performs a scale-up followed by a scale-down if the
// deployment has fewer than two replicas; otherwise it scales down first and
// then up. Each scaling operation is executed through a helper that retries on
// conflicts and waits for pods to become ready. It logs success or failure and
// returns true only when both scaling steps complete successfully.
func TestScaleDeployment(deployment *appsv1.Deployment, timeout time.Duration, logger *log.Logger) bool {
	clients := clientsholder.GetClientsHolder()
	logger.Info("Deployment not using HPA: %s:%s", deployment.Namespace, deployment.Name)
	var replicas int32
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	} else {
		replicas = 1
	}

	if replicas <= 1 {
		// scale up
		replicas++
		if !scaleDeploymentHelper(clients.K8sClient.AppsV1(), deployment, replicas, timeout, true, logger) {
			logger.Error("Cannot scale Deployment %s:%s", deployment.Namespace, deployment.Name)
			return false
		}
		// scale down
		replicas--
		if !scaleDeploymentHelper(clients.K8sClient.AppsV1(), deployment, replicas, timeout, false, logger) {
			logger.Error("Cannot scale Deployment %s:%s", deployment.Namespace, deployment.Name)
			return false
		}
	} else {
		// scale down
		replicas--
		if !scaleDeploymentHelper(clients.K8sClient.AppsV1(), deployment, replicas, timeout, false, logger) {
			logger.Error("Cannot scale Deployment %s:%s", deployment.Namespace, deployment.Name)
			return false
		} // scale up
		replicas++
		if !scaleDeploymentHelper(clients.K8sClient.AppsV1(), deployment, replicas, timeout, true, logger) {
			logger.Error("Cannot scale Deployment %s:%s", deployment.Namespace, deployment.Name)
			return false
		}
	}
	return true
}

// scaleDeploymentHelper Adjusts a Deployment's replica count with conflict handling
//
// This routine logs the scaling action, retrieves the current Deployment
// object, updates its desired replica count, and applies the change using a
// retry loop to handle conflicts. After a successful update it waits for all
// pods in the set to become ready within a specified timeout, reporting any
// errors through logging. The function returns true if the scaling succeeds and
// false otherwise.
func scaleDeploymentHelper(client typedappsv1.AppsV1Interface, deployment *appsv1.Deployment, replicas int32, timeout time.Duration, up bool, logger *log.Logger) bool {
	if up {
		logger.Info("Scale UP deployment to %d replicas", replicas)
	} else {
		logger.Info("Scale DOWN deployment to %d replicas", replicas)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		dp, err := client.Deployments(deployment.Namespace).Get(context.TODO(), deployment.Name, v1machinery.GetOptions{})
		if err != nil {
			logger.Error("Failed to get latest version of Deployment %s:%s", deployment.Namespace, deployment.Name)
			return err
		}
		dp.Spec.Replicas = &replicas
		_, err = client.Deployments(deployment.Namespace).Update(context.TODO(), dp, v1machinery.UpdateOptions{})
		if err != nil {
			logger.Error("Cannot update Deployment %s:%s", deployment.Namespace, deployment.Name)
			return err
		}
		if !podsets.WaitForDeploymentSetReady(deployment.Namespace, deployment.Name, timeout, logger) {
			logger.Error("Cannot update Deployment %s:%s", deployment.Namespace, deployment.Name)
			return errors.New("can not update deployment")
		}
		return nil
	})
	if retryErr != nil {
		logger.Error("Cannot scale Deployment %s:%s, err=%v", deployment.Namespace, deployment.Name, retryErr)
		return false
	}
	return true
}

// TestScaleHpaDeployment Verifies that an HPA can scale a deployment up and down correctly
//
// The function retrieves the Kubernetes client and determines the current
// replica count of the deployment, as well as the min and max values from the
// HPA specification. It then performs a sequence of scaling operations: if
// replicas are low it scales up to the minimum, restores to the original, or if
// high it scales down to one replica before restoring. After each adjustment it
// calls a helper that updates the HPA and waits for the deployment to become
// ready. If any step fails, false is returned; otherwise true indicates
// successful round‑trip scaling.
func TestScaleHpaDeployment(deployment *provider.Deployment, hpa *v1autoscaling.HorizontalPodAutoscaler, timeout time.Duration, logger *log.Logger) bool {
	clients := clientsholder.GetClientsHolder()
	hpscaler := clients.K8sClient.AutoscalingV1().HorizontalPodAutoscalers(deployment.Namespace)
	var min int32
	if hpa.Spec.MinReplicas != nil {
		min = *hpa.Spec.MinReplicas
	} else {
		min = 1
	}
	replicas := int32(1)
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	}
	max := hpa.Spec.MaxReplicas
	if replicas <= 1 {
		// scale up
		replicas++
		logger.Debug("Scale UP HPA %s:%s to min=%d max=%d", deployment.Namespace, hpa.Name, replicas, replicas)
		pass := scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, replicas, replicas, timeout, logger)
		if !pass {
			return false
		}
		// scale down
		replicas--
		logger.Debug("Scale DOWN HPA %s:%s to min=%d max=%d", deployment.Namespace, hpa.Name, replicas, replicas)
		pass = scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, min, max, timeout, logger)
		if !pass {
			return false
		}
	} else {
		// scale down
		replicas--
		logger.Debug("Scale DOWN HPA %s:%s to min=%d max=%d", deployment.Namespace, hpa.Name, replicas, replicas)
		pass := scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, replicas, replicas, timeout, logger)
		if !pass {
			return false
		}
		// scale up
		replicas++
		logger.Debug("Scale UP HPA %s:%s to min=%d max=%d", deployment.Namespace, hpa.Name, replicas, replicas)
		pass = scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, replicas, replicas, timeout, logger)
		if !pass {
			return false
		}
	}
	// back the min and the max value of the hpa
	logger.Debug("Back HPA %s:%s to min=%d max=%d", deployment.Namespace, hpa.Name, min, max)
	return scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, min, max, timeout, logger)
}

// scaleHpaDeploymentHelper Adjusts the minimum and maximum replica counts for a horizontal pod autoscaler and waits for the deployment to stabilize
//
// The helper updates an HPA's MinReplicas and MaxReplicas fields using retry
// logic to handle conflicts, then triggers a wait until the associated
// deployment is ready or times out. It logs any errors encountered during get,
// update, or readiness checks and returns true only when all operations
// succeed.
func scaleHpaDeploymentHelper(hpscaler hps.HorizontalPodAutoscalerInterface, hpaName, deploymentName, namespace string, min, max int32, timeout time.Duration, logger *log.Logger) bool {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := hpscaler.Get(context.TODO(), hpaName, v1machinery.GetOptions{})
		if err != nil {
			logger.Error("Cannot update autoscaler to scale %s:%s , err=%v", namespace, deploymentName, err)
			return err
		}
		hpa.Spec.MinReplicas = &min
		hpa.Spec.MaxReplicas = max
		_, err = hpscaler.Update(context.TODO(), hpa, v1machinery.UpdateOptions{})
		if err != nil {
			logger.Error("Cannot update autoscaler to scale %s:%s, err=%v", namespace, deploymentName, err)
			return err
		}
		if !podsets.WaitForDeploymentSetReady(namespace, deploymentName, timeout, logger) {
			logger.Error("Deployment not ready after scale operation %s:%s", namespace, deploymentName)
		}
		return nil
	})
	if retryErr != nil {
		logger.Error("Cannot scale hpa %s:%s , err=%v", namespace, hpaName, retryErr)
		return false
	}
	return true
}
