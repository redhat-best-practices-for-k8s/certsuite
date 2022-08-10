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

//nolint:dupl
package scaling

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/autodiscover/pkg/clientsholder"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/podsets"

	v1app "k8s.io/api/apps/v1"
	v1autoscaling "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	v1machinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	retry "k8s.io/client-go/util/retry"

	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	hps "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
)

func TestScaleStatefulSet(statefulset *v1app.StatefulSet, timeout time.Duration) bool {
	clients := clientsholder.GetClientsHolder()
	name, namespace := statefulset.Name, statefulset.Namespace
	ssClients := clients.K8sClient.AppsV1().StatefulSets(namespace)
	logrus.Trace("scale statefulset not using HPA ", namespace, ":", name)
	replicas := int32(1)
	if statefulset.Spec.Replicas != nil {
		replicas = *statefulset.Spec.Replicas
	}

	if replicas <= 1 {
		// scale up
		replicas++
		logrus.Trace("scale UP statefulset to ", replicas, " replicas ")
		if !scaleStateFulsetHelper(clients, ssClients, statefulset, replicas, timeout) {
			logrus.Error("can't scale statefulset =", namespace, ":", name)
			return false
		}
		// scale down
		replicas--
		logrus.Trace("scale DOWN statefulset to ", replicas, " replicas ")
		if !scaleStateFulsetHelper(clients, ssClients, statefulset, replicas, timeout) {
			logrus.Error("can't scale statefulset =", namespace, ":", name)
			return false
		}
	} else {
		// scale down
		replicas--
		logrus.Trace("scale DOWN statefulset to ", replicas, " replicas ")
		if !scaleStateFulsetHelper(clients, ssClients, statefulset, replicas, timeout) {
			logrus.Error("can't scale statefulset =", namespace, ":", name)
			return false
		} // scale up
		replicas++
		logrus.Trace("scale UP statefulset to ", replicas, " replicas ")
		if !scaleStateFulsetHelper(clients, ssClients, statefulset, replicas, timeout) {
			logrus.Error("can't scale statefulset =", namespace, ":", name)
			return false
		}
	}
	return true
}

func scaleStateFulsetHelper(clients *clientsholder.ClientsHolder, ssClient v1.StatefulSetInterface, statefulset *v1app.StatefulSet, replicas int32, timeout time.Duration) bool {
	name := statefulset.Name
	namespace := statefulset.Namespace

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of statefulset before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		ss, err := ssClient.Get(context.TODO(), name, v1machinery.GetOptions{})
		if err != nil {
			tnf.ClaimFilePrintf("failed to get latest version of statefulset %s:%s with error %s", namespace, name, err)
			return err
		}
		ss.Spec.Replicas = &replicas
		_, err = clients.K8sClient.AppsV1().StatefulSets(namespace).Update(context.TODO(), ss, v1machinery.UpdateOptions{})
		if err != nil {
			logrus.Error("can't update statefulset ", namespace, ":", name)
			return err
		}
		if !podsets.WaitForStatefulSetReady(namespace, name, timeout) {
			logrus.Error("can't update statefulset ", namespace, ":", name)
			return errors.New("can't update statefulset")
		}
		return nil
	})
	if retryErr != nil {
		logrus.Error("can't scale statefulset ", namespace, ":", name, " error=", retryErr)
		return false
	}
	return true
}

func TestScaleHpaStatefulSet(statefulset *v1app.StatefulSet, hpa *v1autoscaling.HorizontalPodAutoscaler, timeout time.Duration) bool {
	clients := clientsholder.GetClientsHolder()
	hpaName := hpa.Name
	name, namespace := statefulset.Name, statefulset.Namespace
	hpscaler := clients.K8sClient.AutoscalingV1().HorizontalPodAutoscalers(namespace)
	min := int32(1)
	if hpa.Spec.MinReplicas != nil {
		min = *hpa.Spec.MinReplicas
	}
	replicas := int32(1)
	if statefulset.Spec.Replicas != nil {
		replicas = *statefulset.Spec.Replicas
	}
	max := hpa.Spec.MaxReplicas
	if replicas <= 1 {
		// scale up
		replicas++
		logrus.Trace("scale UP HPA ", namespace, ":", hpaName, "To min=", replicas, " max=", replicas)
		pass := scaleHpaStatefulSetHelper(hpscaler, hpaName, name, namespace, replicas, replicas, timeout)
		if !pass {
			return false
		}
		// scale down
		replicas--
		logrus.Trace("scale DOWN HPA ", namespace, ":", hpaName, "To min=", replicas, " max=", replicas)
		pass = scaleHpaStatefulSetHelper(hpscaler, hpaName, name, namespace, replicas, replicas, timeout)
		if !pass {
			return false
		}
	} else {
		// scale down
		replicas--
		logrus.Trace("scale DOWN HPA ", namespace, ":", hpaName, "To min=", replicas, " max=", replicas)
		pass := scaleHpaStatefulSetHelper(hpscaler, hpaName, name, namespace, replicas, replicas, timeout)
		if !pass {
			return false
		}
		// scale up
		replicas++
		logrus.Trace("scale UP HPA ", namespace, ":", hpaName, "To min=", min, " max=", max)
		pass = scaleHpaStatefulSetHelper(hpscaler, hpaName, name, namespace, replicas, replicas, timeout)
		if !pass {
			return false
		}
	}
	// back the min and the max value of the hpa
	logrus.Trace("back HPA ", namespace, ":", hpaName, "To min=", min, " max=", max)
	pass := scaleHpaStatefulSetHelper(hpscaler, hpaName, name, namespace, min, max, timeout)
	return pass
}

func scaleHpaStatefulSetHelper(hpscaler hps.HorizontalPodAutoscalerInterface, hpaName, statefulsetName, namespace string, min, max int32, timeout time.Duration) bool {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := hpscaler.Get(context.TODO(), hpaName, v1machinery.GetOptions{})
		if err != nil {
			logrus.Error("can't Update autoscaler to scale ", namespace, ":", statefulsetName, " error=", err)
			return err
		}
		hpa.Spec.MinReplicas = &min
		hpa.Spec.MaxReplicas = max
		_, err = hpscaler.Update(context.TODO(), hpa, v1machinery.UpdateOptions{})
		if err != nil {
			logrus.Error("can't Update autoscaler to scale ", namespace, ":", statefulsetName, " error=", err)
			return err
		}
		if !podsets.WaitForStatefulSetReady(namespace, statefulsetName, timeout) {
			logrus.Error("statefulsetN not ready after scale operation ", namespace, ":", statefulsetName)
		}
		return nil
	})
	if retryErr != nil {
		logrus.Error("can't scale hpa ", namespace, ":", hpaName, " error=", retryErr)
		return false
	}
	return true
}
