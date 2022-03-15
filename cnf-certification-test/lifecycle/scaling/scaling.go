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
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/podsets"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"

	v1app "k8s.io/api/apps/v1"
	v1autoscaling "k8s.io/api/autoscaling/v1"
	v1machinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	appv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	hps "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	retry "k8s.io/client-go/util/retry"
)

const (
	deployment  = "deployment"
	statefulset = "statefulset"
)

//nolint:funlen
func TestScaleDeployment(podset interface{}, timeout time.Duration) bool {
	clients := clientsholder.GetClientsHolder()
	var name, namespace, podsettype string
	var replicaset, podSetClient interface{}
	switch v := podset.(type) {
	case *v1app.StatefulSet:
		podsettype = statefulset
		logrus.Infof("type is %v", v.Name)
		podset, _ := podset.(*v1app.StatefulSet)
		name, namespace = podset.Name, podset.Namespace
		podSetClient = clients.AppsClients.StatefulSets(namespace)
		replicaset = *podset.Spec.Replicas
	case *v1app.Deployment:
		podsettype = deployment
		logrus.Infof("type is %v", v.Name)
		podset, _ := podset.(*v1app.Deployment)
		name, namespace = podset.Name, podset.Namespace
		podSetClient = clients.AppsClients.Deployments(namespace)
		replicaset = *podset.Spec.Replicas
	}
	logrus.Trace("scale deployment not using HPA ", namespace, ":", name)
	var replicas int32
	if replicaset != nil {
		replicas, _ = replicaset.(int32)
	} else {
		replicas = 1
	}

	if replicas <= 1 {
		// scale up
		replicas++
		if !scaleDeploymentHelper(podSetClient, name, namespace, podsettype, replicas, timeout, true) {
			logrus.Error("can't scale deployment =", namespace, ":", name)
			return false
		}
		// scale down
		replicas--
		if !scaleDeploymentHelper(podSetClient, name, namespace, podsettype, replicas, timeout, false) {
			logrus.Error("can't scale deployment =", namespace, ":", name)
			return false
		}
	} else {
		// scale down
		replicas--
		if !scaleDeploymentHelper(podSetClient, name, namespace, podsettype, replicas, timeout, false) {
			logrus.Error("can't scale deployment =", namespace, podsettype, ":", name)
			return false
		} // scale up
		replicas++
		if !scaleDeploymentHelper(podSetClient, name, namespace, podsettype, replicas, timeout, true) {
			logrus.Error("can't scale deployment =", namespace, ":", name)
			return false
		}
	}
	return true
}

func scaleDeploymentHelper(podSetClient interface{}, name, namespace, podsettype string,
	replicas int32, timeout time.Duration, up bool) bool {
	if up {
		logrus.Trace("scale UP ", podsettype, " to ", replicas, " replicas ")
	} else {
		logrus.Trace("scale DOWN ", podsettype, " to ", replicas, " replicas ")
	}
	// Retrieve the latest version of Deployment before attempting update
	// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return runPodsetScale(podSetClient, name, namespace, podsettype, replicas, timeout)
	})

	if retryErr != nil {
		logrus.Error("can't scale ", podsettype, " ", namespace, ":", name, " error=", retryErr)
		return false
	}
	return true
}

//nolint:funlen
func runPodsetScale(podSetClient interface{}, name,
	namespace, podsettype string, replicas int32, timeout time.Duration) error {
	var errors error
	switch v := podSetClient.(type) {
	case appv1client.DeploymentInterface:
		logrus.Infof("type is %v", v)
		dpClient, _ := podSetClient.(appv1client.DeploymentInterface)
		dp, err := dpClient.Get(context.TODO(), name, v1machinery.GetOptions{})
		if err != nil {
			logrus.Error("failed to get latest version of deployment ", namespace, ":", name)
			return err
		}
		dp.Spec.Replicas = &replicas
		_, errors = dpClient.Update(context.TODO(), dp, v1machinery.UpdateOptions{})
	case appv1client.StatefulSetInterface:
		stClient := podSetClient.(appv1client.StatefulSetInterface)
		st, err := stClient.Get(context.TODO(), name, v1machinery.GetOptions{})
		if err != nil {
			logrus.Error("failed to get latest version of deployment ", namespace, ":", name)
			return err
		}
		st.Spec.Replicas = &replicas
		_, errors = stClient.Update(context.TODO(), st, v1machinery.UpdateOptions{})
	}
	if errors != nil {
		logrus.Error("can't update", podsettype, " ", namespace, ":", name)
		return errors
	}

	if !podsets.IsPodsetReady(namespace, name, podsettype, timeout) {
		logrus.Error("can't update", podsettype, " ", namespace, ":", name)
		return fmt.Errorf("can't update %s", podsettype)
	}
	return nil
}

//nolint:funlen
func TestScaleHpaDeployment(podsetlist interface{}, hpa *v1autoscaling.HorizontalPodAutoscaler, timeout time.Duration) bool {
	clients := clientsholder.GetClientsHolder()
	hpaName := hpa.Name
	var name, namespace, typeset string
	switch v := podsetlist.(type) {
	case *v1app.StatefulSet:
		typeset = statefulset
		logrus.Infof("*v1app.Deployment:%v", v.Name)
		podsetlist, _ := podsetlist.(*v1app.StatefulSet)
		name, namespace = podsetlist.Name, podsetlist.Namespace
	case *v1app.Deployment:
		typeset = deployment
		logrus.Infof("*v1app.Deployment:%v", v.Name)
		podsetlist, _ := podsetlist.(*v1app.Deployment)
		name, namespace = podsetlist.Name, podsetlist.Namespace
	default:
		fmt.Println("unknown")
	}

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
		pass := scaleHpaDeploymentHelper(hpscaler, hpaName, name, namespace, min, max, timeout, scaleUp, typeset)
		if !pass {
			return false
		}
		// scale down
		min--
		max--
		pass = scaleHpaDeploymentHelper(hpscaler, hpaName, name, namespace, min, max, timeout, !scaleUp, typeset)
		if !pass {
			return false
		}
	} else {
		// scale down
		min--
		max--
		scaleUp := false
		pass := scaleHpaDeploymentHelper(hpscaler, hpaName, name, namespace, min, max, timeout, scaleUp, typeset)
		if !pass {
			return false
		}
		// scale up
		min++
		max++
		pass = scaleHpaDeploymentHelper(hpscaler, hpaName, name, namespace, min, max, timeout, !scaleUp, typeset)
		if !pass {
			return false
		}
	}
	return true
}

func scaleHpaDeploymentHelper(hpscaler hps.HorizontalPodAutoscalerInterface, hpaName, deploymentName, namespace string, min, max int32, timeout time.Duration, up bool, podset string) bool {
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
		if !podsets.IsPodsetReady(namespace, deploymentName, podset, timeout) {
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
