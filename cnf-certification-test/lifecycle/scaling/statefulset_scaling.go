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

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/podsets"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/log"

	v1app "k8s.io/api/apps/v1"
	v1autoscaling "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	v1machinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	retry "k8s.io/client-go/util/retry"

	hps "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
)

func TestScaleStatefulSet(statefulset *v1app.StatefulSet, timeout time.Duration, logger *log.Logger) bool {
	clients := clientsholder.GetClientsHolder()
	name, namespace := statefulset.Name, statefulset.Namespace
	ssClients := clients.K8sClient.AppsV1().StatefulSets(namespace)
	logger.Debug("Scale statefulset not using HPA %s:%s", namespace, name)
	replicas := int32(1)
	if statefulset.Spec.Replicas != nil {
		replicas = *statefulset.Spec.Replicas
	}

	if replicas <= 1 {
		// scale up
		replicas++
		logger.Debug("Scale UP statefulset to %d replicas", replicas)
		if !scaleStateFulsetHelper(clients, ssClients, statefulset, replicas, timeout, logger) {
			logger.Error("Cannot scale statefulset = %s:%s", namespace, name)
			return false
		}
		// scale down
		replicas--
		logger.Debug("Scale DOWN statefulset to %d replicas", replicas)
		if !scaleStateFulsetHelper(clients, ssClients, statefulset, replicas, timeout, logger) {
			logger.Error("Cannot scale statefulset = %s:%s", namespace, name)
			return false
		}
	} else {
		// scale down
		replicas--
		logger.Debug("Scale DOWN statefulset to %d replicas", replicas)
		if !scaleStateFulsetHelper(clients, ssClients, statefulset, replicas, timeout, logger) {
			logger.Error("Cannot scale statefulset = %s:%s", namespace, name)
			return false
		} // scale up
		replicas++
		logger.Debug("Scale UP statefulset to %d replicas", replicas)
		if !scaleStateFulsetHelper(clients, ssClients, statefulset, replicas, timeout, logger) {
			logger.Error("Cannot scale statefulset = %s:%s", namespace, name)
			return false
		}
	}
	return true
}

func scaleStateFulsetHelper(clients *clientsholder.ClientsHolder, ssClient v1.StatefulSetInterface, statefulset *v1app.StatefulSet, replicas int32, timeout time.Duration, logger *log.Logger) bool {
	name := statefulset.Name
	namespace := statefulset.Namespace

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of statefulset before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		ss, err := ssClient.Get(context.TODO(), name, v1machinery.GetOptions{})
		if err != nil {
			logger.Error("Failed to get latest version of statefulset %s:%s with error %s", namespace, name, err)
			return err
		}
		ss.Spec.Replicas = &replicas
		_, err = clients.K8sClient.AppsV1().StatefulSets(namespace).Update(context.TODO(), ss, v1machinery.UpdateOptions{})
		if err != nil {
			logger.Error("Cannot update statefulset %s:%s", namespace, name)
			return err
		}
		if !podsets.WaitForStatefulSetReady(namespace, name, timeout, logger) {
			logger.Error("Cannot update statefulset %s:%s", namespace, name)
			return errors.New("can not update statefulset")
		}
		return nil
	})
	if retryErr != nil {
		logger.Error("Cannot scale statefulset %s:%s, err=%v", namespace, name, retryErr)
		return false
	}
	return true
}

func TestScaleHpaStatefulSet(statefulset *v1app.StatefulSet, hpa *v1autoscaling.HorizontalPodAutoscaler, timeout time.Duration, logger *log.Logger) bool {
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
		logger.Debug("Scale UP HPA %s:%s to min=%d max=%d", namespace, hpaName, replicas, replicas)
		pass := scaleHpaStatefulSetHelper(hpscaler, hpaName, name, namespace, replicas, replicas, timeout, logger)
		if !pass {
			return false
		}
		// scale down
		replicas--
		logger.Debug("Scale DOWN HPA %s:%s to min=%d max=%d", namespace, hpaName, replicas, replicas)
		pass = scaleHpaStatefulSetHelper(hpscaler, hpaName, name, namespace, replicas, replicas, timeout, logger)
		if !pass {
			return false
		}
	} else {
		// scale down
		replicas--
		logger.Debug("Scale DOWN HPA %s:%s to min=%d max=%d", namespace, hpaName, replicas, replicas)
		pass := scaleHpaStatefulSetHelper(hpscaler, hpaName, name, namespace, replicas, replicas, timeout, logger)
		if !pass {
			return false
		}
		// scale up
		replicas++
		logger.Debug("Scale UP HPA %s:%s to min=%d max=%d", namespace, hpaName, min, max)
		pass = scaleHpaStatefulSetHelper(hpscaler, hpaName, name, namespace, replicas, replicas, timeout, logger)
		if !pass {
			return false
		}
	}
	// back the min and the max value of the hpa
	logger.Debug("Back HPA %s:%s to min=%d max=%d", namespace, hpaName, min, max)
	pass := scaleHpaStatefulSetHelper(hpscaler, hpaName, name, namespace, min, max, timeout, logger)
	return pass
}

func scaleHpaStatefulSetHelper(hpscaler hps.HorizontalPodAutoscalerInterface, hpaName, statefulsetName, namespace string, min, max int32, timeout time.Duration, logger *log.Logger) bool {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := hpscaler.Get(context.TODO(), hpaName, v1machinery.GetOptions{})
		if err != nil {
			logger.Error("Cannot update autoscaler to scale %s:%s, err=%v", namespace, statefulsetName, err)
			return err
		}
		hpa.Spec.MinReplicas = &min
		hpa.Spec.MaxReplicas = max
		_, err = hpscaler.Update(context.TODO(), hpa, v1machinery.UpdateOptions{})
		if err != nil {
			logger.Error("Cannot update autoscaler to scale %s:%s, err=%v", namespace, statefulsetName, err)
			return err
		}
		if !podsets.WaitForStatefulSetReady(namespace, statefulsetName, timeout, logger) {
			logger.Error("StatefulSet not ready after scale operation %s:%s", namespace, statefulsetName)
		}
		return nil
	})
	if retryErr != nil {
		logger.Error("Cannot scale hpa %s:%s, err=%v", namespace, hpaName, retryErr)
		return false
	}
	return true
}
