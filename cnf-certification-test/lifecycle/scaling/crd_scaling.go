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
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	scalingv1 "k8s.io/api/autoscaling/v1"
	scale "k8s.io/client-go/scale"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	hps "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	retry "k8s.io/client-go/util/retry"
)

func TestScaleCrd(crScale *provider.CrScale, groupResourceSchema schema.GroupResource, timeout time.Duration, logger *log.Logger) bool {
	if crScale == nil {
		logger.Error("CR object is nill")
		return false
	}
	clients := clientsholder.GetClientsHolder()
	replicas := crScale.Spec.Replicas
	name := crScale.GetName()
	namespace := crScale.GetNamespace()

	if replicas <= 1 {
		// scale up
		replicas++
		if !scaleCrHelper(clients.ScalingClient, groupResourceSchema, crScale, replicas, true, timeout, logger) {
			logger.Error("Cannot scale CR %q in namespace %q", name, namespace)
			return false
		}
		// scale down
		replicas--
		if !scaleCrHelper(clients.ScalingClient, groupResourceSchema, crScale, replicas, false, timeout, logger) {
			logger.Error("Cannot scale CR  %q in namespace %q", name, namespace)
			return false
		}
	} else {
		// scale down
		replicas--
		if !scaleCrHelper(clients.ScalingClient, groupResourceSchema, crScale, replicas, false, timeout, logger) {
			logger.Error("Cannot scale CR %q in namespace %q", name, namespace)
			return false
		} // scale up
		replicas++
		if !scaleCrHelper(clients.ScalingClient, groupResourceSchema, crScale, replicas, true, timeout, logger) {
			logger.Error("Cannot scale CR %q in namespace %q", name, namespace)
			return false
		}
	}

	return true
}

func scaleCrHelper(scalesGetter scale.ScalesGetter, rc schema.GroupResource, autoscalerpram *provider.CrScale, replicas int32, up bool, timeout time.Duration, logger *log.Logger) bool {
	if up {
		logger.Debug("Scale UP CRS to %d replicas", replicas)
	} else {
		logger.Debug("Scale DOWN CRS to %d replicas", replicas)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		namespace := autoscalerpram.GetNamespace()
		name := autoscalerpram.GetName()
		scalingObject, err := scalesGetter.Scales(namespace).Get(context.TODO(), rc, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		scalingObject.Spec.Replicas = replicas
		_, err = scalesGetter.Scales(namespace).Update(context.TODO(), rc, scalingObject, metav1.UpdateOptions{})
		if err != nil {
			logger.Error("Cannot update DynamicClient, err=%v", err)
			return err
		}
		if !podsets.WaitForScalingToComplete(namespace, name, timeout, rc, logger) {
			logger.Error("Cannot update CR %s:%s", namespace, name)
			return errors.New("can not update cr")
		}
		return nil
	})
	if retryErr != nil {
		logger.Error("Can notscale DynamicClient, err=%v", retryErr)
		return false
	}
	return true
}

func TestScaleHPACrd(cr *provider.CrScale, hpa *scalingv1.HorizontalPodAutoscaler, groupResourceSchema schema.GroupResource, timeout time.Duration, logger *log.Logger) bool {
	if cr == nil {
		logger.Error("CR object is nill")
		return false
	}
	clients := clientsholder.GetClientsHolder()
	namespace := cr.GetNamespace()

	hpscaler := clients.K8sClient.AutoscalingV1().HorizontalPodAutoscalers(namespace)
	var min int32
	if hpa.Spec.MinReplicas != nil {
		min = *hpa.Spec.MinReplicas
	} else {
		min = 1
	}
	replicas := cr.Spec.Replicas
	name := cr.GetName()
	max := hpa.Spec.MaxReplicas
	if replicas <= 1 {
		// scale up
		replicas++
		logger.Debug("Scale UP HPA %s:%s to min=%d max=%d", namespace, hpa.Name, replicas, replicas)
		pass := scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, replicas, replicas, timeout, groupResourceSchema, logger)
		if !pass {
			return false
		}
		// scale down
		replicas--
		logger.Debug("Scale DOWN HPA %s:%s to min=%d max=%d", namespace, hpa.Name, replicas, replicas)
		pass = scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, min, max, timeout, groupResourceSchema, logger)
		if !pass {
			return false
		}
	} else {
		// scale down
		replicas--
		logger.Debug("Scale DOWN HPA %s:%s to min=%d max=%d", namespace, hpa.Name, replicas, replicas)
		pass := scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, replicas, replicas, timeout, groupResourceSchema, logger)
		if !pass {
			return false
		}
		// scale up
		replicas++
		logger.Debug("Scale UP HPA %s:%s to min=%d max=%d", namespace, hpa.Name, replicas, replicas)
		pass = scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, replicas, replicas, timeout, groupResourceSchema, logger)
		if !pass {
			return false
		}
	}
	// back the min and the max value of the hpa
	logger.Debug("Back HPA %s:%s to min=%d max=%d", namespace, hpa.Name, min, max)
	return scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, min, max, timeout, groupResourceSchema, logger)
}

func scaleHpaCRDHelper(hpscaler hps.HorizontalPodAutoscalerInterface, hpaName, crName, namespace string, min, max int32, timeout time.Duration, groupResourceSchema schema.GroupResource, logger *log.Logger) bool {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := hpscaler.Get(context.TODO(), hpaName, metav1.GetOptions{})
		if err != nil {
			logger.Error("Cannot update autoscaler to scale %s:%s, err=%v", namespace, crName, err)
			return err
		}
		hpa.Spec.MinReplicas = &min
		hpa.Spec.MaxReplicas = max
		_, err = hpscaler.Update(context.TODO(), hpa, metav1.UpdateOptions{})
		if err != nil {
			logger.Error("Cannot update autoscaler to scale %s:%s, err=%v", namespace, crName, err)
			return err
		}
		if !podsets.WaitForScalingToComplete(namespace, crName, timeout, groupResourceSchema, logger) {
			logger.Error("Cannot update CR %s:%s", namespace, crName)
			return errors.New("can not update cr")
		}
		return nil
	})
	if retryErr != nil {
		logger.Error("Cannot scale hpa %s:%s, err=%v", namespace, hpaName, retryErr)
		return false
	}
	return true
}
