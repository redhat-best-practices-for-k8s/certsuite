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

package scaling

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"

	v1app "k8s.io/api/apps/v1"
	v1autoscaling "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	v1machinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	retry "k8s.io/client-go/util/retry"

	hps "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
)

func isDeploymentInstanceReady(deployment *v1app.Deployment) bool {
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
func isStatefulSetReady(statefulset *v1app.StatefulSet) bool {
	var replicas int32
	if statefulset.Spec.Replicas != nil {
		replicas = *(statefulset.Spec.Replicas)
	} else {
		replicas = 1
	}
	if statefulset.Status.ReadyReplicas != replicas ||
		statefulset.Status.AvailableReplicas != replicas ||
		statefulset.Status.UpdatedReplicas != replicas {
		return false
	}
	return true
}

func TestScaleDeployment(deployment *v1app.Deployment, timeout time.Duration) bool {
	clients := clientsholder.GetClientsHolder()
	name, namespace := deployment.Name, deployment.Namespace
	dpClients := clients.AppsClients.Deployments(namespace)
	logrus.Trace("scale deployment not using HPA ", namespace, ":", name)
	var replicas int32
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	} else {
		replicas = 1
	}

	if replicas <= 1 {
		// scale up
		replicas++
		if !scaleDeploymentHelper(clients, dpClients, deployment, replicas, timeout, true) {
			logrus.Error("can't scale deployment =", namespace, ":", name)
			return false
		}
		// scale down
		replicas--
		if !scaleDeploymentHelper(clients, dpClients, deployment, replicas, timeout, false) {
			logrus.Error("can't scale deployment =", namespace, ":", name)
			return false
		}
	} else {
		// scale down
		replicas--
		if !scaleDeploymentHelper(clients, dpClients, deployment, replicas, timeout, false) {
			logrus.Error("can't scale deployment =", namespace, ":", name)
			return false
		} // scale up
		replicas++
		if !scaleDeploymentHelper(clients, dpClients, deployment, replicas, timeout, true) {
			logrus.Error("can't scale deployment =", namespace, ":", name)
			return false
		}
	}
	return true
}

func scaleDeploymentHelper(clients *clientsholder.ClientsHolder, dpClient v1.DeploymentInterface, deployment *v1app.Deployment, replicas int32, timeout time.Duration, up bool) bool {
	if up {
		logrus.Trace("scale UP deployment to ", replicas, " replicas ")
	} else {
		logrus.Trace("scale DOWN deployment to ", replicas, " replicas ")
	}

	name := deployment.Name
	namespace := deployment.Namespace

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		dp, err := dpClient.Get(context.TODO(), name, v1machinery.GetOptions{})
		if err != nil {
			logrus.Error("failed to get latest version of deployment ", namespace, ":", name)
			return err
		}
		dp.Spec.Replicas = &replicas
		_, err = clients.AppsClients.Deployments(namespace).Update(context.TODO(), dp, v1machinery.UpdateOptions{})
		if err != nil {
			logrus.Error("can't update deployment ", namespace, ":", name)
			return err
		}
		if !isDeploymentReady(namespace, name, timeout) {
			logrus.Error("can't update deployment ", namespace, ":", name)
			return errors.New("can't update deployment")
		}
		return nil
	})
	if retryErr != nil {
		logrus.Error("can't scale deployment ", namespace, ":", name, " error=", retryErr)
		return false
	}
	return true
}

func isDeploymentReady(ns, name string, timeout time.Duration) bool {
	logrus.Trace("check if deployment ", ns, ":", name, " is ready ")
	clients := clientsholder.GetClientsHolder()
	start := time.Now()
	for time.Since(start) < timeout {
		dp, err := provider.GetUpdatedDeployment(clients.AppsClients, ns, name)
		if err == nil && isDeploymentInstanceReady(dp) {
			logrus.Trace("deployment ", ns, ":", name, " is ready ")
			return true
		}
		time.Sleep(time.Second)
	}
	logrus.Error("deployment ", ns, ":", name, " is not ready ")
	return false
}

func TestScaleHpaDeployment(deployment *v1app.Deployment, hpa *v1autoscaling.HorizontalPodAutoscaler, timeout time.Duration) bool {
	clients := clientsholder.GetClientsHolder()
	hpaName := hpa.Name
	name, namespace := deployment.Name, deployment.Namespace
	hpscaler := clients.K8sClient.AutoscalingV1().HorizontalPodAutoscalers(namespace)
	var min int32
	if hpa.Spec.MinReplicas != nil {
		min = *hpa.Spec.MinReplicas
	} else {
		min = 1
	}
	max := hpa.Spec.MaxReplicas
	if min <= 1 {
		// scale up
		min++
		max++
		scaleUp := true
		pass := scaleHpaDeploymentHelper(hpscaler, hpaName, name, namespace, min, max, timeout, scaleUp)
		if !pass {
			return false
		}
		// scale down
		min--
		max--
		pass = scaleHpaDeploymentHelper(hpscaler, hpaName, name, namespace, min, max, timeout, !scaleUp)
		if !pass {
			return false
		}
	} else {
		// scale down
		min--
		max--
		scaleUp := false
		pass := scaleHpaDeploymentHelper(hpscaler, hpaName, name, namespace, min, max, timeout, scaleUp)
		if !pass {
			return false
		}
		// scale up
		min++
		max++
		pass = scaleHpaDeploymentHelper(hpscaler, hpaName, name, namespace, min, max, timeout, !scaleUp)
		if !pass {
			return false
		}
	}
	return true
}

func scaleHpaDeploymentHelper(hpscaler hps.HorizontalPodAutoscalerInterface, hpaName, deploymentName, namespace string, min, max int32, timeout time.Duration, up bool) bool {
	if up {
		logrus.Trace("scale UP HPA ", namespace, ":", hpaName, "To min=", min, "max=", max)
	} else {
		logrus.Trace("scale DOWN HPA ", namespace, ":", hpaName, "To min=", min, "max=", max)
	}
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := hpscaler.Get(context.TODO(), hpaName, v1machinery.GetOptions{})
		if err != nil {
			logrus.Error("can't Update autoscaler to scale ", namespace, ":", deploymentName, " error=", err)
			return err
		}
		hpa.Spec.MinReplicas = &min
		hpa.Spec.MaxReplicas = max
		_, err = hpscaler.Update(context.TODO(), hpa, v1machinery.UpdateOptions{})
		if err != nil {
			logrus.Error("can't Update autoscaler to scale ", namespace, ":", deploymentName, " error=", err)
			return err
		}
		if !isDeploymentReady(namespace, deploymentName, timeout) {
			logrus.Error("deployment not ready after scale operation ", namespace, ":", deploymentName)
		}
		return nil
	})
	if retryErr != nil {
		logrus.Error("can't scale hpa ", namespace, ":", hpaName, " error=", retryErr)
		return false
	}
	return true
}
