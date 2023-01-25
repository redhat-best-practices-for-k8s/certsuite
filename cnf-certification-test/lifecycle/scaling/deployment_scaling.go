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

//nolint:dupl
package scaling

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/podsets"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"

	v1autoscaling "k8s.io/api/autoscaling/v1"

	v1machinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	retry "k8s.io/client-go/util/retry"

	appsv1 "k8s.io/api/apps/v1"
	hps "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
)

func TestScaleDeployment(deployment *appsv1.Deployment, timeout time.Duration) bool {
	clients := clientsholder.GetClientsHolder()
	logrus.Trace("scale deployment not using HPA ", deployment.Namespace, ":", deployment.Name)
	var replicas int32
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	} else {
		replicas = 1
	}

	if replicas <= 1 {
		// scale up
		replicas++
		if !scaleDeploymentHelper(clients, deployment, replicas, timeout, true) {
			logrus.Error("can not scale deployment =", deployment.Namespace, ":", deployment.Name)
			return false
		}
		// scale down
		replicas--
		if !scaleDeploymentHelper(clients, deployment, replicas, timeout, false) {
			logrus.Error("can not scale deployment =", deployment.Namespace, ":", deployment.Name)
			return false
		}
	} else {
		// scale down
		replicas--
		if !scaleDeploymentHelper(clients, deployment, replicas, timeout, false) {
			logrus.Error("can not scale deployment =", deployment.Namespace, ":", deployment.Name)
			return false
		} // scale up
		replicas++
		if !scaleDeploymentHelper(clients, deployment, replicas, timeout, true) {
			logrus.Error("can not scale deployment =", deployment.Namespace, ":", deployment.Name)
			return false
		}
	}
	return true
}

func scaleDeploymentHelper(clients *clientsholder.ClientsHolder, deployment *appsv1.Deployment, replicas int32, timeout time.Duration, up bool) bool {
	if up {
		logrus.Trace("scale UP deployment to ", replicas, " replicas ")
	} else {
		logrus.Trace("scale DOWN deployment to ", replicas, " replicas ")
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		dp, err := clients.K8sClient.AppsV1().Deployments(deployment.Namespace).Get(context.TODO(), deployment.Name, v1machinery.GetOptions{})
		if err != nil {
			logrus.Error("failed to get latest version of deployment ", deployment.Namespace, ":", deployment.Name)
			return err
		}
		dp.Spec.Replicas = &replicas
		_, err = clients.K8sClient.AppsV1().Deployments(deployment.Namespace).Update(context.TODO(), dp, v1machinery.UpdateOptions{})
		if err != nil {
			logrus.Error("can not update deployment ", deployment.Namespace, ":", deployment.Name)
			return err
		}
		if !podsets.WaitForDeploymentSetReady(deployment.Namespace, deployment.Name, timeout) {
			logrus.Error("can not update deployment ", deployment.Namespace, ":", deployment.Name)
			return errors.New("can not update deployment")
		}
		return nil
	})
	if retryErr != nil {
		logrus.Error("can not scale deployment ", deployment.Namespace, ":", deployment.Name, " error=", retryErr)
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
		logrus.Trace("scale UP HPA ", deployment.Namespace, ":", hpa.Name, "To min=", replicas, " max=", replicas)
		pass := scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, replicas, replicas, timeout)
		if !pass {
			return false
		}
		// scale down
		replicas--
		logrus.Trace("scale DOWN HPA ", deployment.Namespace, ":", hpa.Name, "To min=", replicas, " max=", replicas)
		pass = scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, min, max, timeout)
		if !pass {
			return false
		}
	} else {
		// scale down
		replicas--
		logrus.Trace("scale DOWN HPA ", deployment.Namespace, ":", hpa.Name, "To min=", replicas, " max=", replicas)
		pass := scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, replicas, replicas, timeout)
		if !pass {
			return false
		}
		// scale up
		replicas++
		logrus.Trace("scale UP HPA ", deployment.Namespace, ":", hpa.Name, "To min=", replicas, " max=", replicas)
		pass = scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, replicas, replicas, timeout)
		if !pass {
			return false
		}
	}
	// back the min and the max value of the hpa
	logrus.Trace("back HPA ", deployment.Namespace, ":", hpa.Name, "To min=", min, " max=", max)
	return scaleHpaDeploymentHelper(hpscaler, hpa.Name, deployment.Name, deployment.Namespace, min, max, timeout)
}

func scaleHpaDeploymentHelper(hpscaler hps.HorizontalPodAutoscalerInterface, hpaName, deploymentName, namespace string, min, max int32, timeout time.Duration) bool {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := hpscaler.Get(context.TODO(), hpaName, v1machinery.GetOptions{})
		if err != nil {
			logrus.Error("can not Update autoscaler to scale ", namespace, ":", deploymentName, " error=", err)
			return err
		}
		hpa.Spec.MinReplicas = &min
		hpa.Spec.MaxReplicas = max
		_, err = hpscaler.Update(context.TODO(), hpa, v1machinery.UpdateOptions{})
		if err != nil {
			logrus.Error("can not Update autoscaler to scale ", namespace, ":", deploymentName, " error=", err)
			return err
		}
		if !podsets.WaitForDeploymentSetReady(namespace, deploymentName, timeout) {
			logrus.Error("deployment not ready after scale operation ", namespace, ":", deploymentName)
		}
		return nil
	})
	if retryErr != nil {
		logrus.Error("can not scale hpa ", namespace, ":", hpaName, " error=", retryErr)
		return false
	}
	return true
}
