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
	scalingv1 "k8s.io/api/autoscaling/v1"
	scale "k8s.io/client-go/scale"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	hps "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	retry "k8s.io/client-go/util/retry"
)

// TestScaleCrd verifies that a CRD can be scaled successfully.
//
// It takes a pointer to a CrScale provider, the group and resource of the
// custom resource, a timeout duration for scaling operations, and a logger.
// The function returns true if the scaling operation succeeds within the
// given timeout; otherwise it logs errors and returns false.
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

// scaleCrHelper applies a scaling operation to a custom resource and waits for the change to complete.
//
// It retrieves the current scale of the specified resource, updates it with the desired replica count,
// and then polls until the scaling action has finished or the timeout is reached.
// The function logs progress and errors using the provided logger.  
// It returns true if scaling succeeded within the given duration, otherwise false.
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

// TestScaleHPACrd tests scaling of a HorizontalPodAutoscaler CRD.
//
// It attempts to scale the provided HPA resource according to the desired
// replica count in the CrScale struct. The function returns true if the
// scaling operation succeeds, otherwise false and logs errors.
// Parameters include a pointer to the test context, the HPA object,
// the GroupResource of the target resource, a timeout duration,
// and a logger for debug output.
func TestScaleHPACrd(cr *provider.CrScale, hpa *scalingv1.HorizontalPodAutoscaler, groupResourceSchema schema.GroupResource, timeout time.Duration, logger *log.Logger) bool {
	if cr == nil {
		logger.Error("CR object is nill")
		return false
	}
	clients := clientsholder.GetClientsHolder()
	namespace := cr.GetNamespace()

	hpscaler := clients.K8sClient.AutoscalingV1().HorizontalPodAutoscalers(namespace)
	min := int32(1)
	if hpa.Spec.MinReplicas != nil {
		min = *hpa.Spec.MinReplicas
	}
	replicas := cr.Spec.Replicas
	name := cr.GetName()

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
		pass = scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, min, hpa.Spec.MaxReplicas, timeout, groupResourceSchema, logger)
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
	logger.Debug("Back HPA %s:%s to min=%d max=%d", namespace, hpa.Name, min, hpa.Spec.MaxReplicas)
	return scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, min, hpa.Spec.MaxReplicas, timeout, groupResourceSchema, logger)
}

// scaleHpaCRDHelper scales a Custom Resource Definition's Horizontal Pod Autoscaler and waits for the scaling operation to complete.
//
// It receives an HPA interface, the namespace and name of the target resource, the CRD group version string,
// minimum and maximum replica counts, a timeout duration, the GroupResource describing the CRD,
// and a logger. The function attempts to update the HPA's desired replicas within the provided bounds,
// retrying on conflict errors. After updating, it blocks until scaling completes or the timeout expires.
// It returns true if scaling succeeded before the timeout, otherwise false.
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
