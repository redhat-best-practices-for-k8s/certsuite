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

//nolint:dupl
package scaling

import (
	"context"
	"errors"
	"time"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/podsets"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"

	v1autoscaling "k8s.io/api/autoscaling/v1"

	v1machinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	retry "k8s.io/client-go/util/retry"

	appsv1 "k8s.io/api/apps/v1"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	hps "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
)

func TestScaleDeployment(deployment *appsv1.Deployment, timeout time.Duration) bool {
	clients := clientsholder.GetClientsHolder()
	log.Debug("scale deployment not using HPA %s:%s", deployment.Namespace, deployment.Name)
	var replicas int32
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	} else {
		replicas = 1
	}

	if replicas <= 1 {
		// scale up
		replicas++
		if !scaleDeploymentHelper(clients.K8sClient.AppsV1(), deployment, replicas, timeout, true) {
			log.Error("can not scale deployment %s:%s", deployment.Namespace, deployment.Name)
			return false
		}
		// scale down
		replicas--
		if !scaleDeploymentHelper(clients.K8sClient.AppsV1(), deployment, replicas, timeout, false) {
			log.Error("can not scale deployment %s:%s", deployment.Namespace, deployment.Name)
			return false
		}
	} else {
		// scale down
		replicas--
		if !scaleDeploymentHelper(clients.K8sClient.AppsV1(), deployment, replicas, timeout, false) {
			log.Error("can not scale deployment %s:%s", deployment.Namespace, deployment.Name)
			return false
		} // scale up
		replicas++
		if !scaleDeploymentHelper(clients.K8sClient.AppsV1(), deployment, replicas, timeout, true) {
			log.Error("can not scale deployment %s:%s", deployment.Namespace, deployment.Name)
			return false
		}
	}
	return true
}

func scaleDeploymentHelper(client typedappsv1.AppsV1Interface, deployment *appsv1.Deployment, replicas int32, timeout time.Duration, up bool) bool {
	if up {
		log.Debug("scale UP deployment to %d replicas", replicas)
	} else {
		log.Debug("scale DOWN deployment to %d replicas", replicas)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		dp, err := client.Deployments(deployment.Namespace).Get(context.TODO(), deployment.Name, v1machinery.GetOptions{})
		if err != nil {
			log.Error("failed to get latest version of deployment %s:%s", deployment.Namespace, deployment.Name)
			return err
		}
		dp.Spec.Replicas = &replicas
		_, err = client.Deployments(deployment.Namespace).Update(context.TODO(), dp, v1machinery.UpdateOptions{})
		if err != nil {
			log.Error("can not update deployment %s:%s", deployment.Namespace, deployment.Name)
			return err
		}
		if !podsets.WaitForDeploymentSetReady(deployment.Namespace, deployment.Name, timeout) {
			log.Error("can not update deployment %s:%s", deployment.Namespace, deployment.Name)
			return errors.New("can not update deployment")
		}
		return nil
	})
	if retryErr != nil {
		log.Error("can not scale deployment %s:%s, err=%v", deployment.Namespace, deployment.Name, retryErr)
		return false
	}
	return true
}

func TestScaleHpaDeployment(deployment *provider.Deployment, hpa *v1autoscaling.HorizontalPodAutoscaler, timeout time.Duration) bool {
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
		log.Debug("scale UP HPA %s:%s to min=%d max=%d", deployment.Namespace, hpa.Name, replicas, replicas)
		pass := scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, replicas, replicas, timeout)
		if !pass {
			return false
		}
		// scale down
		replicas--
		log.Debug("scale DOWN HPA %s:%s to min=%d max=%d", deployment.Namespace, hpa.Name, replicas, replicas)
		pass = scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, min, max, timeout)
		if !pass {
			return false
		}
	} else {
		// scale down
		replicas--
		log.Debug("scale DOWN HPA %s:%s to min=%d max=%d", deployment.Namespace, hpa.Name, replicas, replicas)
		pass := scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, replicas, replicas, timeout)
		if !pass {
			return false
		}
		// scale up
		replicas++
		log.Debug("scale UP HPA %s:%s to min=%d max=%d", deployment.Namespace, hpa.Name, replicas, replicas)
		pass = scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, replicas, replicas, timeout)
		if !pass {
			return false
		}
	}
	// back the min and the max value of the hpa
	log.Debug("back HPA %s:%s to min=%d max=%d", deployment.Namespace, hpa.Name, min, max)
	return scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, min, max, timeout)
}

func scaleHpaDeploymentHelper(hpscaler hps.HorizontalPodAutoscalerInterface, hpaName, deploymentName, namespace string, min, max int32, timeout time.Duration) bool {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := hpscaler.Get(context.TODO(), hpaName, v1machinery.GetOptions{})
		if err != nil {
			log.Error("can not Update autoscaler to scale %s:%s , err=%v", namespace, deploymentName, err)
			return err
		}
		hpa.Spec.MinReplicas = &min
		hpa.Spec.MaxReplicas = max
		_, err = hpscaler.Update(context.TODO(), hpa, v1machinery.UpdateOptions{})
		if err != nil {
			log.Error("can not Update autoscaler to scale %s:%s, err=%v", namespace, deploymentName, err)
			return err
		}
		if !podsets.WaitForDeploymentSetReady(namespace, deploymentName, timeout) {
			log.Error("deployment not ready after scale operation %s:%s", namespace, deploymentName)
		}
		return nil
	})
	if retryErr != nil {
		log.Error("can not scale hpa %s:%s , err=%v", namespace, hpaName, retryErr)
		return false
	}
	return true
}
