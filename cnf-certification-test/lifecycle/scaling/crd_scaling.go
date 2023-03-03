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
	scalingv1 "k8s.io/api/autoscaling/v1"
	scale "k8s.io/client-go/scale"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	hps "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	retry "k8s.io/client-go/util/retry"
)

func TestScaleCrd(crScale *provider.CrScale, groupResourceSchema schema.GroupResource, timeout time.Duration) bool {
	if crScale == nil {
		logrus.Errorf("cc object is nill")
		return false
	}
	clients := clientsholder.GetClientsHolder()
	replicas := crScale.Spec.Replicas
	name := crScale.GetName()
	namespace := crScale.GetNamespace()

	if replicas <= 1 {
		// scale up
		replicas++
		if !scaleCrHelper(clients.ScalingClient, groupResourceSchema, crScale, replicas, true, timeout) {
			logrus.Errorf("Can not scale cr %s in namespace %s", name, namespace)
			return false
		}
		// scale down
		replicas--
		if !scaleCrHelper(clients.ScalingClient, groupResourceSchema, crScale, replicas, false, timeout) {
			logrus.Errorf("Can not scale cr  %s in namespace %s", name, namespace)
			return false
		}
	} else {
		// scale down
		replicas--
		if !scaleCrHelper(clients.ScalingClient, groupResourceSchema, crScale, replicas, false, timeout) {
			logrus.Errorf("Can not scale cr %s in namespace %s", name, namespace)
			return false
		} // scale up
		replicas++
		if !scaleCrHelper(clients.ScalingClient, groupResourceSchema, crScale, replicas, true, timeout) {
			logrus.Errorf("Can not scale cr %s in namespace %s", name, namespace)
			return false
		}
	}

	return true
}

func scaleCrHelper(scalesGetter scale.ScalesGetter, rc schema.GroupResource, autoscalerpram *provider.CrScale, replicas int32, up bool, timeout time.Duration) bool {
	if up {
		logrus.Trace("scale UP CRS to ", replicas, " replicas ")
	} else {
		logrus.Trace("scale DOWN CRS to ", replicas, " replicas ")
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
			logrus.Error("Can not update DynamicClient ")
			return err
		}
		if !podsets.WaitForScalingToComplete(namespace, name, timeout, rc) {
			logrus.Error("can not update cr ", namespace, ":", name)
			return errors.New("can not update cr")
		}
		return nil
	})
	if retryErr != nil {
		logrus.Error("Can not scale DynamicClient ", " error=", retryErr)
		return false
	}
	return true
}

//nolint:funlen
func TestScaleHPACrd(cr *provider.CrScale, hpa *scalingv1.HorizontalPodAutoscaler, groupResourceSchema schema.GroupResource, timeout time.Duration) bool {
	if cr == nil {
		logrus.Errorf("cc object is nill")
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
		logrus.Trace("scale UP HPA ", namespace, ":", hpa.Name, "To min=", replicas, " max=", replicas)
		pass := scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, replicas, replicas, timeout, groupResourceSchema)
		if !pass {
			return false
		}
		// scale down
		replicas--
		logrus.Trace("scale DOWN HPA ", namespace, ":", hpa.Name, "To min=", replicas, " max=", replicas)
		pass = scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, min, max, timeout, groupResourceSchema)
		if !pass {
			return false
		}
	} else {
		// scale down
		replicas--
		logrus.Trace("scale DOWN HPA ", namespace, ":", hpa.Name, "To min=", replicas, " max=", replicas)
		pass := scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, replicas, replicas, timeout, groupResourceSchema)
		if !pass {
			return false
		}
		// scale up
		replicas++
		logrus.Trace("scale UP HPA ", namespace, ":", hpa.Name, "To min=", replicas, " max=", replicas)
		pass = scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, replicas, replicas, timeout, groupResourceSchema)
		if !pass {
			return false
		}
	}
	// back the min and the max value of the hpa
	logrus.Trace("back HPA ", namespace, ":", hpa.Name, "To min=", min, " max=", max)
	return scaleHpaCRDHelper(hpscaler, hpa.Name, name, namespace, min, max, timeout, groupResourceSchema)
}

func scaleHpaCRDHelper(hpscaler hps.HorizontalPodAutoscalerInterface, hpaName, crName, namespace string, min, max int32, timeout time.Duration, groupResourceSchema schema.GroupResource) bool {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		hpa, err := hpscaler.Get(context.TODO(), hpaName, metav1.GetOptions{})
		if err != nil {
			logrus.Error("Can not Update autoscaler to scale ", namespace, ":", crName, " error=", err)
			return err
		}
		hpa.Spec.MinReplicas = &min
		hpa.Spec.MaxReplicas = max
		_, err = hpscaler.Update(context.TODO(), hpa, metav1.UpdateOptions{})
		if err != nil {
			logrus.Error("Can not Update autoscaler to scale ", namespace, ":", crName, " error=", err)
			return err
		}
		if !podsets.WaitForScalingToComplete(namespace, crName, timeout, groupResourceSchema) {
			logrus.Error("Can not update cr ", namespace, ":", crName)
			return errors.New("can not update cr")
		}
		return nil
	})
	if retryErr != nil {
		logrus.Error("Can not scale hpa ", namespace, ":", hpaName, " error=", retryErr)
		return false
	}
	return true
}
